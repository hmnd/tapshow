package display

import (
	"os"
	"testing"
)

func TestCompositorString(t *testing.T) {
	tests := []struct {
		compositor Compositor
		expected   string
	}{
		{CompositorGNOME, "GNOME"},
		{CompositorKDE, "KDE Plasma"},
		{CompositorSway, "Sway"},
		{CompositorHyprland, "Hyprland"},
		{CompositorWlroots, "wlroots-based"},
		{CompositorUnknown, "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			got := tt.compositor.String()
			if got != tt.expected {
				t.Errorf("Compositor.String() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestCompositorDetect(t *testing.T) {
	origDesktop := os.Getenv("XDG_CURRENT_DESKTOP")
	origSession := os.Getenv("XDG_SESSION_DESKTOP")
	origWayland := os.Getenv("WAYLAND_DISPLAY")
	origSway := os.Getenv("SWAYSOCK")
	origHypr := os.Getenv("HYPRLAND_INSTANCE_SIGNATURE")

	defer func() {
		os.Setenv("XDG_CURRENT_DESKTOP", origDesktop)
		os.Setenv("XDG_SESSION_DESKTOP", origSession)
		os.Setenv("WAYLAND_DISPLAY", origWayland)
		os.Setenv("SWAYSOCK", origSway)
		os.Setenv("HYPRLAND_INSTANCE_SIGNATURE", origHypr)
	}()

	tests := []struct {
		name     string
		envSetup func()
		expected Compositor
	}{
		{
			name: "No Wayland",
			envSetup: func() {
				os.Unsetenv("WAYLAND_DISPLAY")
				os.Unsetenv("XDG_CURRENT_DESKTOP")
			},
			expected: CompositorUnknown,
		},
		{
			name: "KDE",
			envSetup: func() {
				os.Setenv("WAYLAND_DISPLAY", "wayland-0")
				os.Setenv("XDG_CURRENT_DESKTOP", "KDE")
				os.Unsetenv("SWAYSOCK")
				os.Unsetenv("HYPRLAND_INSTANCE_SIGNATURE")
			},
			expected: CompositorKDE,
		},
		{
			name: "GNOME",
			envSetup: func() {
				os.Setenv("WAYLAND_DISPLAY", "wayland-0")
				os.Setenv("XDG_CURRENT_DESKTOP", "GNOME")
				os.Unsetenv("SWAYSOCK")
				os.Unsetenv("HYPRLAND_INSTANCE_SIGNATURE")
			},
			expected: CompositorGNOME,
		},
		{
			name: "Sway via env",
			envSetup: func() {
				os.Setenv("WAYLAND_DISPLAY", "wayland-0")
				os.Setenv("XDG_CURRENT_DESKTOP", "sway")
				os.Unsetenv("SWAYSOCK")
				os.Unsetenv("HYPRLAND_INSTANCE_SIGNATURE")
			},
			expected: CompositorSway,
		},
		{
			name: "Sway via socket",
			envSetup: func() {
				os.Setenv("WAYLAND_DISPLAY", "wayland-0")
				os.Setenv("XDG_CURRENT_DESKTOP", "")
				os.Setenv("SWAYSOCK", "/run/user/1000/sway-ipc.sock")
				os.Unsetenv("HYPRLAND_INSTANCE_SIGNATURE")
			},
			expected: CompositorSway,
		},
		{
			name: "Hyprland via signature",
			envSetup: func() {
				os.Setenv("WAYLAND_DISPLAY", "wayland-0")
				os.Setenv("XDG_CURRENT_DESKTOP", "")
				os.Unsetenv("SWAYSOCK")
				os.Setenv("HYPRLAND_INSTANCE_SIGNATURE", "abc123")
			},
			expected: CompositorHyprland,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.envSetup()
			got := Detect()
			if got != tt.expected {
				t.Errorf("Detect() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestNew(t *testing.T) {
	backend := New()
	if backend == nil {
		t.Error("New() returned nil")
	}

	_, isGTK := backend.(*GTKWindowBackend)
	if !isGTK {
		t.Error("New() should return GTKWindowBackend")
	}
}
