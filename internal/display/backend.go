package display

import (
	"github.com/tapshow/tapshow/internal/config"
	"github.com/tapshow/tapshow/internal/processor"
)

type Backend interface {
	Init(cfg *config.Config) error

	Show(event processor.DisplayEvent)

	UpdateHistory(events []processor.DisplayEvent)

	Reset()

	SetPaused(paused bool)

	Run() error

	Stop()
}

func New() Backend {
	return NewGTKWindowBackend(Detect())
}
