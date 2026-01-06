<h1 align=center>tapshow</h1>

`tapshow` is a lightweight keystroke visualizer for Wayland systems. Displays your keystrokes as a minimal overlay window - perfect for screen recordings, presentations, and live coding!

<p align="center">
  <img title="Screenshot of tapshow in action" src="assets/example.png" />
</p>

## Features

- Real-time keystroke visualization
- Modifier key combination display (e.g., `Ctrl+Shift+A`)
- Privacy mode - auto-pause for sensitive applications
- Inherits your GTK styles

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

### From Release

[Download the latest release](https://github.com/hmnd/tapshow/releases/latest)

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

# Continuously log active app to help with pause_on_apps
tapshow debug active-app

# Show version
tapshow version
```

## Configuration

Run `tapshow config init` or refer to [the default config](configs/default.toml)

## Privacy Mode

Tapshow can automatically pause when sensitive applications are focused. Add application names to `pause_on_apps` in your config:

```toml
[privacy]
pause_on_apps = [
  "simple-match",
  { class = "org.keepassxc" },
  { process = "1password", title = "unlock" },
]
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

## License

MIT License - see LICENSE file for details.
