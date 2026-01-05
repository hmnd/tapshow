package privacy

import (
	"encoding/json"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/tapshow/tapshow/internal/display"
)

type Monitor struct {
	mu         sync.RWMutex
	pauseApps  []string
	compositor display.Compositor
	paused     bool
	done       chan struct{}
	onChange   func(paused bool)
}

func NewMonitor(pauseApps []string, onChange func(paused bool)) *Monitor {
	return &Monitor{
		pauseApps:  pauseApps,
		compositor: display.Detect(),
		done:       make(chan struct{}),
		onChange:   onChange,
	}
}

func (m *Monitor) Start() {
	if len(m.pauseApps) == 0 {
		return // Nothing to monitor
	}

	go m.monitorLoop()
}

func (m *Monitor) Stop() {
	close(m.done)
}

func (m *Monitor) IsPaused() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.paused
}

func (m *Monitor) monitorLoop() {
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-m.done:
			return
		case <-ticker.C:
			m.checkFocusedWindow()
		}
	}
}

func (m *Monitor) checkFocusedWindow() {
	appName := m.getFocusedApp()
	if appName == "" {
		return
	}

	shouldPause := false
	appLower := strings.ToLower(appName)

	for _, pauseApp := range m.pauseApps {
		if strings.Contains(appLower, strings.ToLower(pauseApp)) {
			shouldPause = true
			break
		}
	}

	m.mu.Lock()
	changed := m.paused != shouldPause
	m.paused = shouldPause
	m.mu.Unlock()

	if changed && m.onChange != nil {
		m.onChange(shouldPause)
	}
}

func (m *Monitor) getFocusedApp() string {
	switch m.compositor {
	case display.CompositorSway:
		return m.getSwayFocusedApp()
	case display.CompositorHyprland:
		return m.getHyprlandFocusedApp()
	case display.CompositorKDE:
		return m.getKDEFocusedApp()
	case display.CompositorGNOME:
		return m.getGNOMEFocusedApp()
	default:
		return ""
	}
}

func (m *Monitor) getSwayFocusedApp() string {
	cmd := exec.Command("swaymsg", "-t", "get_tree")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}

	var tree swayTree
	if err := json.Unmarshal(output, &tree); err != nil {
		return ""
	}

	return findFocusedSway(&tree)
}

type swayTree struct {
	Name          string     `json:"name"`
	AppID         string     `json:"app_id"`
	Focused       bool       `json:"focused"`
	Nodes         []swayTree `json:"nodes"`
	FloatingNodes []swayTree `json:"floating_nodes"`
}

func findFocusedSway(node *swayTree) string {
	if node.Focused && node.AppID != "" {
		return node.AppID
	}
	if node.Focused && node.Name != "" {
		return node.Name
	}

	for i := range node.Nodes {
		if result := findFocusedSway(&node.Nodes[i]); result != "" {
			return result
		}
	}
	for i := range node.FloatingNodes {
		if result := findFocusedSway(&node.FloatingNodes[i]); result != "" {
			return result
		}
	}

	return ""
}

func (m *Monitor) getHyprlandFocusedApp() string {
	cmd := exec.Command("hyprctl", "activewindow", "-j")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}

	var window struct {
		Class string `json:"class"`
		Title string `json:"title"`
	}
	if err := json.Unmarshal(output, &window); err != nil {
		return ""
	}

	if window.Class != "" {
		return window.Class
	}
	return window.Title
}

func (m *Monitor) getKDEFocusedApp() string {
	// Use qdbus to query KWin
	cmd := exec.Command("qdbus", "org.kde.KWin", "/KWin", "org.kde.KWin.activeWindow")
	output, err := cmd.Output()
	if err != nil {
		// Try alternative method
		return m.getKDEFocusedAppAlt()
	}

	return strings.TrimSpace(string(output))
}

func (m *Monitor) getKDEFocusedAppAlt() string {
	// Alternative using kdotool if available
	cmd := exec.Command("kdotool", "getactivewindow", "getwindowname")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}

func (m *Monitor) getGNOMEFocusedApp() string {
	// Use gdbus to query GNOME Shell
	script := `
		global.get_window_actors()
			.map(a => a.meta_window)
			.find(w => w.has_focus())
			?.get_wm_class() || ''
	`

	cmd := exec.Command("gdbus", "call", "--session",
		"--dest", "org.gnome.Shell",
		"--object-path", "/org/gnome/Shell",
		"--method", "org.gnome.Shell.Eval",
		script)

	output, err := cmd.Output()
	if err != nil {
		return ""
	}

	// Parse the output (format: "(true, 'app-name')")
	result := string(output)
	if strings.Contains(result, "true") {
		start := strings.Index(result, "'")
		end := strings.LastIndex(result, "'")
		if start != -1 && end > start {
			return result[start+1 : end]
		}
	}

	return ""
}
