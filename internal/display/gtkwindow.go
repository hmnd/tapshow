package display

import (
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"github.com/tapshow/tapshow/internal/config"
	"github.com/tapshow/tapshow/internal/processor"
)

// GTKWindowBackend implements Backend using a draggable GTK window
type GTKWindowBackend struct {
	GTKCommon
	quit chan struct{}
}

func NewGTKWindowBackend(compositor Compositor) *GTKWindowBackend {
	return &GTKWindowBackend{
		GTKCommon: GTKCommon{compositor: compositor},
		quit:      make(chan struct{}),
	}
}

func (g *GTKWindowBackend) Init(cfg *config.Config) error {
	g.InitGTK(cfg, g.compositor)
	return nil
}

func (g *GTKWindowBackend) Show(event processor.DisplayEvent) {
	g.ShowKey(event)
}

func (g *GTKWindowBackend) UpdateHistory(events []processor.DisplayEvent) {
	g.UpdateHistoryDisplay(events)
}

func (g *GTKWindowBackend) SetPaused(paused bool) {
	g.SetPausedState(paused)
}

func (g *GTKWindowBackend) Run() error {
	g.app = gtk.NewApplication("ca.icewolf.tapshow", 0)

	g.app.ConnectActivate(func() {
		g.setupWindow()
	})

	if code := g.app.Run(nil); code != 0 {
		return nil
	}

	return nil
}

func (g *GTKWindowBackend) setupWindow() {
	g.window = gtk.NewWindow()
	g.window.SetApplication(g.app)
	g.window.SetTitle("tapshow")
	g.window.SetDecorated(false)
	g.window.SetResizable(false)
	g.window.SetFocusOnClick(false)
	g.window.SetCanFocus(false)

	handle := g.BuildUI()
	handle.AddCSSClass("overlay-container")
	g.window.SetChild(handle)

	g.ApplyCSS()
	g.window.SetVisible(true)
}

func (g *GTKWindowBackend) Stop() {
	close(g.quit)
	if g.app != nil {
		g.app.Quit()
	}
}
