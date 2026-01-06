package privacy

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/tapshow/tapshow/internal/config"
	"github.com/tapshow/tapshow/internal/display"
)

type WindowInfo struct {
	Class       string
	ProcessName string
	Path        string
	Title       string
}

func (w WindowInfo) String() string {
	return fmt.Sprintf("class=%s process=%s path=%s title=%s", w.Class, w.ProcessName, w.Path, w.Title)
}

func (w WindowInfo) IsEmpty() bool {
	return w.Class == "" && w.ProcessName == "" && w.Path == "" && w.Title == ""
}

func (w WindowInfo) MatchesAny(pattern string) bool {
	if pattern == "" {
		return false
	}
	pattern = strings.ToLower(pattern)
	if w.Class != "" && strings.Contains(strings.ToLower(w.Class), pattern) {
		return true
	}
	if w.ProcessName != "" && strings.Contains(strings.ToLower(w.ProcessName), pattern) {
		return true
	}
	if w.Path != "" && strings.Contains(strings.ToLower(w.Path), pattern) {
		return true
	}
	if w.Title != "" && strings.Contains(strings.ToLower(w.Title), pattern) {
		return true
	}
	return false
}

func (w WindowInfo) Matches(m config.AppMatcher) bool {
	if m.Value != "" {
		return w.MatchesAny(m.Value)
	}
	if m.Class != "" && !strings.Contains(strings.ToLower(w.Class), strings.ToLower(m.Class)) {
		return false
	}
	if m.Process != "" && !strings.Contains(strings.ToLower(w.ProcessName), strings.ToLower(m.Process)) {
		return false
	}
	if m.Path != "" && !strings.Contains(strings.ToLower(w.Path), strings.ToLower(m.Path)) {
		return false
	}
	if m.Title != "" && !strings.Contains(strings.ToLower(w.Title), strings.ToLower(m.Title)) {
		return false
	}
	return m.Class != "" || m.Process != "" || m.Path != "" || m.Title != ""
}

type Monitor struct {
	mu             sync.RWMutex
	matchers       config.AppMatchers
	compositor     display.Compositor
	paused         bool
	done           chan struct{}
	onChange       func(paused bool)
	resumeCooldown time.Duration
	lastMatchAt    time.Time
}

const defaultResumeCooldownMs = 500 * time.Millisecond

func NewMonitor(matchers config.AppMatchers, onChange func(paused bool)) *Monitor {
	return &Monitor{
		matchers:       matchers,
		compositor:     display.Detect(),
		done:           make(chan struct{}),
		onChange:       onChange,
		resumeCooldown: defaultResumeCooldownMs,
	}
}

func (m *Monitor) Start() {
	if len(m.matchers) == 0 {
		return
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

func (m *Monitor) testCheckWindow(info WindowInfo) {
	m.checkWindowInfo(info)
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
	info := GetFocusedWindow(m.compositor)
	if info.IsEmpty() {
		return
	}
	m.checkWindowInfo(info)
}

func (m *Monitor) checkWindowInfo(info WindowInfo) {
	shouldPause := false
	for _, matcher := range m.matchers {
		if info.Matches(matcher) {
			shouldPause = true
			break
		}
	}

	m.mu.Lock()
	if shouldPause {
		m.lastMatchAt = time.Now()
	}

	newPaused := shouldPause
	if m.paused && !shouldPause && m.resumeCooldown > 0 {
		if time.Since(m.lastMatchAt) < m.resumeCooldown {
			newPaused = true
		}
	}

	changed := m.paused != newPaused
	m.paused = newPaused
	m.mu.Unlock()

	if changed && m.onChange != nil {
		m.onChange(newPaused)
	}
}

func GetFocusedWindow(compositor display.Compositor) WindowInfo {
	switch compositor {
	case display.CompositorSway:
		return getSwayFocusedWindow()
	case display.CompositorHyprland:
		return getHyprlandFocusedWindow()
	case display.CompositorKDE:
		return getKDEFocusedWindow()
	case display.CompositorGNOME:
		return getGNOMEFocusedWindow()
	default:
		return WindowInfo{}
	}
}

func getProcessInfo(pid int) (name, path string) {
	if pid <= 0 {
		return "", ""
	}
	commPath := fmt.Sprintf("/proc/%d/comm", pid)
	if data, err := os.ReadFile(commPath); err == nil {
		name = strings.TrimSpace(string(data))
	}
	exePath := fmt.Sprintf("/proc/%d/exe", pid)
	if target, err := os.Readlink(exePath); err == nil {
		path = target
	}
	return name, path
}

func getSwayFocusedWindow() WindowInfo {
	cmd := exec.Command("swaymsg", "-t", "get_tree")
	output, err := cmd.Output()
	if err != nil {
		return WindowInfo{}
	}

	var tree swayTree
	if err := json.Unmarshal(output, &tree); err != nil {
		return WindowInfo{}
	}

	return findFocusedSway(&tree)
}

type swayTree struct {
	Name          string     `json:"name"`
	AppID         string     `json:"app_id"`
	PID           int        `json:"pid"`
	Focused       bool       `json:"focused"`
	Nodes         []swayTree `json:"nodes"`
	FloatingNodes []swayTree `json:"floating_nodes"`
}

func findFocusedSway(node *swayTree) WindowInfo {
	if node.Focused && (node.AppID != "" || node.Name != "") {
		class := node.AppID
		if class == "" {
			class = node.Name
		}
		procName, path := getProcessInfo(node.PID)
		return WindowInfo{Class: class, ProcessName: procName, Path: path, Title: node.Name}
	}

	for i := range node.Nodes {
		if result := findFocusedSway(&node.Nodes[i]); !result.IsEmpty() {
			return result
		}
	}
	for i := range node.FloatingNodes {
		if result := findFocusedSway(&node.FloatingNodes[i]); !result.IsEmpty() {
			return result
		}
	}

	return WindowInfo{}
}

func getHyprlandFocusedWindow() WindowInfo {
	cmd := exec.Command("hyprctl", "activewindow", "-j")
	output, err := cmd.Output()
	if err != nil {
		return WindowInfo{}
	}

	var window struct {
		Class string `json:"class"`
		Title string `json:"title"`
		PID   int    `json:"pid"`
	}
	if err := json.Unmarshal(output, &window); err != nil {
		return WindowInfo{}
	}

	procName, path := getProcessInfo(window.PID)
	return WindowInfo{Class: window.Class, ProcessName: procName, Path: path, Title: window.Title}
}

func getKDEFocusedWindow() WindowInfo {
	winID := getKDEActiveWindowID()
	if winID == "" {
		return WindowInfo{}
	}

	var info WindowInfo

	cmd := exec.Command("kdotool", "getwindowclassname", winID)
	if output, err := cmd.Output(); err == nil {
		info.Class = strings.TrimSpace(string(output))
	}

	cmd = exec.Command("kdotool", "getwindowname", winID)
	if output, err := cmd.Output(); err == nil {
		info.Title = strings.TrimSpace(string(output))
	}

	cmd = exec.Command("kdotool", "getwindowpid", winID)
	if output, err := cmd.Output(); err == nil {
		if pid, err := strconv.Atoi(strings.TrimSpace(string(output))); err == nil {
			info.ProcessName, info.Path = getProcessInfo(pid)
		}
	}

	return info
}

func getKDEActiveWindowID() string {
	cmd := exec.Command("kdotool", "getactivewindow")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}

func getGNOMEFocusedWindow() WindowInfo {
	script := `
		(function() {
			const w = global.get_window_actors()
				.map(a => a.meta_window)
				.find(w => w.has_focus());
			if (!w) return '';
			return JSON.stringify({
				class: w.get_wm_class() || '',
				title: w.get_title() || '',
				pid: w.get_pid() || 0
			});
		})()
	`

	cmd := exec.Command("gdbus", "call", "--session",
		"--dest", "org.gnome.Shell",
		"--object-path", "/org/gnome/Shell",
		"--method", "org.gnome.Shell.Eval",
		script)

	output, err := cmd.Output()
	if err != nil {
		return WindowInfo{}
	}

	result := string(output)
	if !strings.Contains(result, "true") {
		return WindowInfo{}
	}

	start := strings.Index(result, "'")
	end := strings.LastIndex(result, "'")
	if start == -1 || end <= start {
		return WindowInfo{}
	}

	jsonStr := strings.ReplaceAll(result[start+1:end], `\'`, `'`)
	jsonStr = strings.ReplaceAll(jsonStr, `\\`, `\`)

	var data struct {
		Class string `json:"class"`
		Title string `json:"title"`
		PID   int    `json:"pid"`
	}
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		return WindowInfo{Class: jsonStr}
	}

	procName, path := getProcessInfo(data.PID)
	return WindowInfo{Class: data.Class, ProcessName: procName, Path: path, Title: data.Title}
}
