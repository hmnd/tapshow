package privacy

import (
	"testing"
	"testing/synctest"
	"time"

	"github.com/tapshow/tapshow/internal/config"
)

func TestWindowInfo_MatchesAny(t *testing.T) {
	tests := []struct {
		name    string
		info    WindowInfo
		pattern string
		want    bool
	}{
		{
			name:    "matches class",
			info:    WindowInfo{Class: "org.keepassxc.KeePassXC"},
			pattern: "keepass",
			want:    true,
		},
		{
			name:    "matches process",
			info:    WindowInfo{ProcessName: "1password"},
			pattern: "1password",
			want:    true,
		},
		{
			name:    "matches path",
			info:    WindowInfo{Path: "/usr/bin/keepassxc"},
			pattern: "keepass",
			want:    true,
		},
		{
			name:    "matches title",
			info:    WindowInfo{Title: "Unlock Database - KeePassXC"},
			pattern: "unlock",
			want:    true,
		},
		{
			name:    "case insensitive",
			info:    WindowInfo{Class: "KeePassXC"},
			pattern: "KEEPASS",
			want:    true,
		},
		{
			name:    "no match",
			info:    WindowInfo{Class: "firefox", ProcessName: "firefox", Path: "/usr/bin/firefox"},
			pattern: "chrome",
			want:    false,
		},
		{
			name:    "empty pattern matches nothing",
			info:    WindowInfo{Class: "firefox"},
			pattern: "",
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.info.MatchesAny(tt.pattern); got != tt.want {
				t.Errorf("MatchesAny(%q) = %v, want %v", tt.pattern, got, tt.want)
			}
		})
	}
}

func TestWindowInfo_Matches_SimpleValue(t *testing.T) {
	info := WindowInfo{
		Class:       "com.1password.1Password",
		ProcessName: "1password",
		Path:        "/opt/1Password/1password",
		Title:       "1Password - Unlock",
	}

	tests := []struct {
		name    string
		matcher config.AppMatcher
		want    bool
	}{
		{
			name:    "simple match on class",
			matcher: config.AppMatcher{Value: "1password"},
			want:    true,
		},
		{
			name:    "simple match on path",
			matcher: config.AppMatcher{Value: "/opt/1Password"},
			want:    true,
		},
		{
			name:    "simple no match",
			matcher: config.AppMatcher{Value: "keepass"},
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := info.Matches(tt.matcher); got != tt.want {
				t.Errorf("Matches() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWindowInfo_Matches_Structured(t *testing.T) {
	info := WindowInfo{
		Class:       "org.keepassxc.KeePassXC",
		ProcessName: "keepassxc",
		Path:        "/usr/bin/keepassxc",
		Title:       "Unlock Database - KeePassXC",
	}

	tests := []struct {
		name    string
		matcher config.AppMatcher
		want    bool
	}{
		{
			name:    "match by class only",
			matcher: config.AppMatcher{Class: "keepassxc"},
			want:    true,
		},
		{
			name:    "match by process only",
			matcher: config.AppMatcher{Process: "keepassxc"},
			want:    true,
		},
		{
			name:    "match by path only",
			matcher: config.AppMatcher{Path: "/usr/bin/keepassxc"},
			want:    true,
		},
		{
			name:    "match by title only",
			matcher: config.AppMatcher{Title: "unlock"},
			want:    true,
		},
		{
			name:    "match by multiple fields (AND)",
			matcher: config.AppMatcher{Class: "keepassxc", Title: "unlock"},
			want:    true,
		},
		{
			name:    "partial match fails (AND logic)",
			matcher: config.AppMatcher{Class: "keepassxc", Title: "settings"},
			want:    false,
		},
		{
			name:    "class mismatch",
			matcher: config.AppMatcher{Class: "1password"},
			want:    false,
		},
		{
			name:    "empty matcher returns false",
			matcher: config.AppMatcher{},
			want:    false,
		},
		{
			name:    "case insensitive class",
			matcher: config.AppMatcher{Class: "KEEPASSXC"},
			want:    true,
		},
		{
			name:    "case insensitive title",
			matcher: config.AppMatcher{Title: "UNLOCK DATABASE"},
			want:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := info.Matches(tt.matcher); got != tt.want {
				t.Errorf("Matches() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWindowInfo_IsEmpty(t *testing.T) {
	tests := []struct {
		name string
		info WindowInfo
		want bool
	}{
		{
			name: "all empty",
			info: WindowInfo{},
			want: true,
		},
		{
			name: "has class",
			info: WindowInfo{Class: "firefox"},
			want: false,
		},
		{
			name: "has process",
			info: WindowInfo{ProcessName: "firefox"},
			want: false,
		},
		{
			name: "has path",
			info: WindowInfo{Path: "/usr/bin/firefox"},
			want: false,
		},
		{
			name: "has title",
			info: WindowInfo{Title: "Mozilla Firefox"},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.info.IsEmpty(); got != tt.want {
				t.Errorf("IsEmpty() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMonitor_InitialState(t *testing.T) {
	matchers := config.AppMatchers{
		{Value: "1password"},
	}

	monitor := NewMonitor(matchers, func(paused bool) {})

	if monitor.IsPaused() {
		t.Error("Monitor should not be paused initially")
	}
}

func TestMonitor_ResumeCooldown(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		matchers := config.AppMatchers{
			{Value: "1password"},
		}

		var pauseChanges []bool
		monitor := NewMonitor(matchers, func(paused bool) {
			pauseChanges = append(pauseChanges, paused)
		})

		sensitiveApp := WindowInfo{Class: "1password"}
		normalApp := WindowInfo{Class: "firefox"}

		monitor.testCheckWindow(sensitiveApp)
		if !monitor.IsPaused() {
			t.Error("Monitor should pause for sensitive app")
		}

		monitor.testCheckWindow(normalApp)
		if !monitor.IsPaused() {
			t.Error("Monitor should remain paused during cooldown period")
		}

		time.Sleep(defaultResumeCooldownMs)

		monitor.testCheckWindow(normalApp)
		if monitor.IsPaused() {
			t.Error("Monitor should unpause after cooldown period")
		}

		pauseChanges = nil
		monitor.testCheckWindow(sensitiveApp)
		if !monitor.IsPaused() {
			t.Error("Monitor should pause immediately for sensitive app")
		}
		if len(pauseChanges) != 1 || pauseChanges[0] != true {
			t.Errorf("Expected one pause=true change, got %v", pauseChanges)
		}
	})
}

func TestMatching_Integration(t *testing.T) {
	matchers := config.AppMatchers{
		{Value: "1password"},
		{Class: "keepassxc"},
		{Process: "secret-tool", Title: "unlock"},
	}

	testCases := []struct {
		info        WindowInfo
		shouldPause bool
		description string
	}{
		{
			info:        WindowInfo{Class: "firefox", ProcessName: "firefox"},
			shouldPause: false,
			description: "regular app should not pause",
		},
		{
			info:        WindowInfo{Class: "com.1password.1Password", ProcessName: "1password"},
			shouldPause: true,
			description: "1password should pause (simple match)",
		},
		{
			info:        WindowInfo{Class: "org.keepassxc.KeePassXC"},
			shouldPause: true,
			description: "keepassxc should pause (class match)",
		},
		{
			info:        WindowInfo{ProcessName: "secret-tool", Title: "unlock vault"},
			shouldPause: true,
			description: "secret-tool with unlock title should pause",
		},
		{
			info:        WindowInfo{ProcessName: "secret-tool", Title: "settings"},
			shouldPause: false,
			description: "secret-tool without unlock title should not pause",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			shouldPause := false
			for _, matcher := range matchers {
				if tc.info.Matches(matcher) {
					shouldPause = true
					break
				}
			}
			if shouldPause != tc.shouldPause {
				t.Errorf("got pause=%v, want %v", shouldPause, tc.shouldPause)
			}
		})
	}
}
