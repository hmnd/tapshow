# Tapshow

A lightweight keystroke visualizer for Wayland Linux systems. Displays your keystrokes as a minimal overlay window - perfect for screen recordings, presentations, and live coding.

## Features

- Real-time keystroke visualization
- Modifier key combination display (e.g., `Ctrl+Shift+A`)
- Keystroke history display
- Held key indication
- Privacy mode - auto-pause for sensitive applications
- Configurable appearance and behavior
- Automatic compositor detection

## Compositor Support

| Compositor | Support | Notes                                                                                 |
| ---------- | ------- | ------------------------------------------------------------------------------------- |
| Sway       | Yes     |                                                                                       |
| Hyprland   | Yes     |                                                                                       |
| KDE Plasma | Yes     | To display above other windows: right-click window > More Actions > Keep Above Others |
| GNOME      | Yes     | To display above other windows: right-click window > Always on Top                    |

## Prerequisites

### System Dependencies

```bash
# Debian/Ubuntu
sudo apt install libgtk-4-dev libglib2.0-dev libgirepository1.0-dev

# Fedora
sudo dnf install gtk4-devel glib2-devel gobject-introspection-devel

# Arch Linux
sudo pacman -S gtk4 glib2 gobject-introspection
```

### Input Group Membership

Tapshow reads keyboard input from `/dev/input/event*` devices, which requires membership in the `input` group:

```bash
sudo usermod -aG input $USER
```

**Important:** Log out and back in for the group change to take effect.

To verify:

```bash
groups | grep input
```

## Installation

### From Source

```bash
# Clone repository
git clone https://github.com/tapshow/tapshow.git
cd tapshow

# Build (requires zig for C compilation)
just build

# Install (optional)
sudo cp bin/tapshow /usr/local/bin/
```

## Usage

```bash
# Run tapshow
tapshow

# Show config file location
tapshow config path

# Create default config file
tapshow config init

# Show version
tapshow version
```

## Configuration

Configuration file location: `~/.config/tapshow/config.toml`

```toml
[display]
position = "bottom-right"  # top-left, top-right, bottom-left, bottom-right
margin_x = 20
margin_y = 40
timeout_ms = 2000          # Fade timeout in milliseconds
history_count = 4          # Number of previous keys to show
show_held_keys = true

[appearance]
theme = "dark"             # dark, light
font_size = 18
opacity = 0.85
corner_radius = 8

[behavior]
combine_modifiers = true   # Show "Ctrl+A" instead of separate keys
show_modifier_only = false # Show when only modifier is pressed
excluded_keys = []         # Keys to never display, e.g., ["CapsLock"]

[privacy]
pause_on_apps = []         # Pause when these apps are focused
# Example: ["1password", "keepassxc", "bitwarden"]
```

## Privacy Mode

Tapshow can automatically pause when sensitive applications are focused. Add application names to `pause_on_apps` in your config:

```toml
[privacy]
pause_on_apps = ["1password", "keepassxc", "bitwarden", "gnome-keyring"]
```

The privacy monitor checks the focused window every 500ms and pauses the display when a matching app name is detected.

## Troubleshooting

### "no keyboards found" Error

Ensure you're in the `input` group:

```bash
groups | grep input
```

If not present:

```bash
sudo usermod -aG input $USER
# Then log out and back in
```

### Window Not Visible

1. Verify Wayland session: `echo $XDG_SESSION_TYPE` should output `wayland`
2. For GNOME/KDE, set the window to stay on top (see compositor support table)

### Keys Not Appearing

1. Check that tapshow is running: `pgrep tapshow`
2. Verify input permissions: `ls -la /dev/input/event*`
3. Try running with `sudo` to test (not recommended for regular use)

## Building from Source

### Requirements

- Go 1.25+
- GTK4 development libraries
- GLib development libraries
- Zig (for C/C++ compilation)
- Just (command runner)

### Build

```bash
just build
```

### Build with custom version

```bash
# Version is read from VERSION file
just build
```

## License

MIT License - see LICENSE file for details.
