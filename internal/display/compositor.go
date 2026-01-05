package display

import (
	"os"
	"os/exec"
	"strings"
)

type Compositor int

const (
	CompositorUnknown Compositor = iota
	CompositorGNOME
	CompositorKDE
	CompositorSway
	CompositorHyprland
	CompositorWlroots
)

func (c Compositor) String() string {
	switch c {
	case CompositorGNOME:
		return "GNOME"
	case CompositorKDE:
		return "KDE Plasma"
	case CompositorSway:
		return "Sway"
	case CompositorHyprland:
		return "Hyprland"
	case CompositorWlroots:
		return "wlroots-based"
	default:
		return "Unknown"
	}
}

func Detect() Compositor {
	if os.Getenv("WAYLAND_DISPLAY") == "" {
		return CompositorUnknown
	}

	desktop := strings.ToLower(os.Getenv("XDG_CURRENT_DESKTOP"))
	session := strings.ToLower(os.Getenv("XDG_SESSION_DESKTOP"))

	switch {
	case strings.Contains(desktop, "gnome") || strings.Contains(session, "gnome"):
		return CompositorGNOME
	case strings.Contains(desktop, "kde") || strings.Contains(session, "plasma"):
		return CompositorKDE
	case strings.Contains(desktop, "sway") || strings.Contains(session, "sway"):
		return CompositorSway
	case strings.Contains(desktop, "hyprland"):
		return CompositorHyprland
	}

	if os.Getenv("HYPRLAND_INSTANCE_SIGNATURE") != "" {
		return CompositorHyprland
	}

	if os.Getenv("SWAYSOCK") != "" {
		return CompositorSway
	}

	return detectViaProcess()
}

func detectViaProcess() Compositor {
	compositors := map[string]Compositor{
		"gnome-shell":  CompositorGNOME,
		"kwin_wayland": CompositorKDE,
		"sway":         CompositorSway,
		"Hyprland":     CompositorHyprland,
	}

	for proc, comp := range compositors {
		cmd := exec.Command("pgrep", "-x", proc)
		if err := cmd.Run(); err == nil {
			return comp
		}
	}

	return CompositorUnknown
}
