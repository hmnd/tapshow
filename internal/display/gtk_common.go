package display

import (
	"fmt"
	"sync"

	"github.com/diamondburned/gotk4/pkg/gdk/v4"
	"github.com/diamondburned/gotk4/pkg/glib/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"github.com/tapshow/tapshow/internal/config"
	"github.com/tapshow/tapshow/internal/processor"
)

type GTKCommon struct {
	mu          sync.Mutex
	cfg         *config.Config
	app         *gtk.Application
	window      *gtk.Window
	keysBox     *gtk.Box
	placeholder *gtk.Label
	hasKeys     bool
	paused      bool
	compositor  Compositor
}

func (g *GTKCommon) InitGTK(cfg *config.Config, compositor Compositor) {
	g.cfg = cfg
	g.compositor = compositor
}

func (g *GTKCommon) BuildUI() *gtk.WindowHandle {
	g.keysBox = gtk.NewBox(gtk.OrientationHorizontal, 4)
	g.keysBox.SetHAlign(gtk.AlignCenter)
	g.keysBox.SetHExpand(true)

	g.placeholder = gtk.NewLabel("Listening for keystrokes...")
	g.placeholder.AddCSSClass("placeholder")
	g.keysBox.Append(g.placeholder)

	handle := gtk.NewWindowHandle()
	handle.SetChild(g.keysBox)
	handle.SetHAlign(gtk.AlignCenter)

	return handle
}

func (g *GTKCommon) ApplyCSS() {
	css := g.generateCSS()

	provider := gtk.NewCSSProvider()
	provider.LoadFromString(css)

	gtk.StyleContextAddProviderForDisplay(
		gdk.DisplayGetDefault(),
		provider,
		gtk.STYLE_PROVIDER_PRIORITY_APPLICATION,
	)
}

func (g *GTKCommon) generateCSS() string {
	var fallbackContainerBg, fallbackKeyBg string
	if g.cfg.Appearance.Theme == "light" {
		fallbackContainerBg = "#f0f0f0"
		fallbackKeyBg = "#ffffff"
	} else {
		fallbackContainerBg = "#2d2d2d"
		fallbackKeyBg = "#3d3d3d"
	}

	return fmt.Sprintf(`
window {
	background-color: transparent;
}

.overlay-container {
	/* Solid fallback under theme color */
	background: linear-gradient(@window_bg_color, @window_bg_color),
	            linear-gradient(%s, %s);
	opacity: %f;
	border-radius: %dpx;
	padding: 8px;
	border: 1px solid @borders;
}

.key-frame {
	/* Solid fallback under theme color */
	background: linear-gradient(@theme_bg_color, @theme_bg_color),
	            linear-gradient(%s, %s);
	border-radius: 6px;
	padding: 0;
	margin: 2px;
	border: 1px solid @borders;
	box-shadow: 0 1px 2px alpha(black, 0.15);
}

.key-label {
	padding: 8px 14px;
	font-weight: 600;
	font-size: %dpx;
	color: @theme_text_color;
}

.key-recent {
	opacity: 0.5;
}

.placeholder {
	padding: 8px 14px;
	font-size: %dpx;
	font-style: italic;
	opacity: 0.7;
	color: @theme_text_color;
}
`,
		fallbackContainerBg,
		fallbackContainerBg,
		g.cfg.Appearance.Opacity,
		g.cfg.Appearance.CornerRadius,
		fallbackKeyBg,
		fallbackKeyBg,
		g.cfg.Appearance.FontSize,
		g.cfg.Appearance.FontSize-2,
	)
}

func (g *GTKCommon) clearChildren() {
	for child := g.keysBox.FirstChild(); child != nil; child = g.keysBox.FirstChild() {
		g.keysBox.Remove(child)
	}
}

func (g *GTKCommon) createKeyWidget(text string, isRecent bool) *gtk.Frame {
	label := gtk.NewLabel(text)
	label.AddCSSClass("key-label")

	frame := gtk.NewFrame("")
	frame.SetLabel("")
	frame.SetChild(label)
	frame.AddCSSClass("key-frame")

	if isRecent {
		frame.AddCSSClass("key-recent")
	}

	return frame
}

func (g *GTKCommon) ShowKey(event processor.DisplayEvent) {
	g.mu.Lock()
	defer g.mu.Unlock()

	if g.paused || g.keysBox == nil {
		return
	}

	glib.IdleAdd(func() {
		if g.keysBox == nil {
			return
		}

		if !g.hasKeys && g.placeholder != nil {
			g.keysBox.Remove(g.placeholder)
			g.placeholder = nil
			g.hasKeys = true
		}

		g.clearChildren()
		g.keysBox.Append(g.createKeyWidget(event.Text, false))

		if g.window != nil {
			g.window.QueueResize()
		}
	})
}

func (g *GTKCommon) UpdateHistoryDisplay(events []processor.DisplayEvent) {
	g.mu.Lock()
	defer g.mu.Unlock()

	if g.paused || g.keysBox == nil {
		return
	}

	glib.IdleAdd(func() {
		if g.keysBox == nil {
			return
		}

		g.clearChildren()

		maxKeys := g.cfg.Display.HistoryCount
		start := len(events) - maxKeys
		if start < 0 {
			start = 0
		}

		for i := start; i < len(events); i++ {
			isRecent := i < len(events)-1
			g.keysBox.Append(g.createKeyWidget(events[i].Text, isRecent))
		}

		if g.window != nil {
			g.window.QueueResize()
		}
	})
}

func (g *GTKCommon) SetPausedState(paused bool) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.paused = paused
}

func (g *GTKCommon) ResetDisplay() {
	glib.IdleAdd(func() {
		g.mu.Lock()
		defer g.mu.Unlock()

		if g.keysBox == nil {
			return
		}

		g.clearChildren()
		g.placeholder = gtk.NewLabel("Listening for keystrokes...")
		g.placeholder.AddCSSClass("placeholder")
		g.keysBox.Append(g.placeholder)
		g.hasKeys = false

		if g.window != nil {
			g.window.QueueResize()
		}
	})
}
