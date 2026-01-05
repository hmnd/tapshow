package processor

import (
	"strings"
	"sync"
	"time"

	"github.com/tapshow/tapshow/internal/input"
)

type DisplayEvent struct {
	Text      string
	Timestamp time.Time
	IsHeld    bool
}

type Processor struct {
	events    chan DisplayEvent
	done      chan struct{}
	config    Config
	mu        sync.Mutex
	modifiers input.Modifier
	history   []DisplayEvent
	lastKey   *input.KeyEvent
	heldTimer *time.Timer
}

type Config struct {
	CombineModifiers bool
	ShowModifierOnly bool
	ShowHeldKeys     bool
	HeldKeyTimeout   time.Duration
	HistoryCount     int
	ExcludedKeys     []string
}

func DefaultConfig() Config {
	return Config{
		CombineModifiers: true,
		ShowModifierOnly: false,
		ShowHeldKeys:     true,
		HeldKeyTimeout:   500 * time.Millisecond,
		HistoryCount:     4,
		ExcludedKeys:     []string{},
	}
}

func New(cfg Config) *Processor {
	return &Processor{
		events:  make(chan DisplayEvent, 50),
		done:    make(chan struct{}),
		config:  cfg,
		history: make([]DisplayEvent, 0, cfg.HistoryCount),
	}
}

func (p *Processor) Events() <-chan DisplayEvent {
	return p.events
}

func (p *Processor) History() []DisplayEvent {
	p.mu.Lock()
	defer p.mu.Unlock()
	result := make([]DisplayEvent, len(p.history))
	copy(result, p.history)
	return result
}

func (p *Processor) Process(inputEvents <-chan input.KeyEvent) {
	for {
		select {
		case <-p.done:
			return
		case ev, ok := <-inputEvents:
			if !ok {
				return
			}
			p.handleKeyEvent(ev)
		}
	}
}

func (p *Processor) Stop() {
	close(p.done)
	if p.heldTimer != nil {
		p.heldTimer.Stop()
	}
}

func (p *Processor) handleKeyEvent(ev input.KeyEvent) {
	for _, excluded := range p.config.ExcludedKeys {
		if strings.EqualFold(ev.Name, excluded) {
			return
		}
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	if input.IsModifier(ev.Code) {
		mod := input.GetModifier(ev.Code)
		if ev.State == input.KeyPressed {
			p.modifiers |= mod
		} else if ev.State == input.KeyReleased {
			p.modifiers &^= mod
		}

		if p.config.ShowModifierOnly && ev.State == input.KeyPressed {
			p.emitEvent(ev.Name, false)
		}
		return
	}

	switch ev.State {
	case input.KeyPressed:
		text := p.buildKeyText(ev.Name)
		p.emitEvent(text, false)
		p.lastKey = &ev

		if p.config.ShowHeldKeys {
			if p.heldTimer != nil {
				p.heldTimer.Stop()
			}
			p.heldTimer = time.AfterFunc(p.config.HeldKeyTimeout, func() {
				p.mu.Lock()
				defer p.mu.Unlock()
				if p.lastKey != nil && p.lastKey.Code == ev.Code {
					text := p.buildKeyText(ev.Name)
					p.emitEvent(text+" (held)", true)
				}
			})
		}

	case input.KeyReleased:
		if p.heldTimer != nil {
			p.heldTimer.Stop()
		}
		p.lastKey = nil

	case input.KeyHeld:
		if p.config.ShowHeldKeys {
			text := p.buildKeyText(ev.Name)
			p.emitEvent(text, true)
		}
	}
}

func (p *Processor) buildKeyText(keyName string) string {
	if !p.config.CombineModifiers || p.modifiers == 0 {
		return keyName
	}

	var parts []string
	for _, m := range []struct {
		mod  input.Modifier
		name string
	}{
		{input.ModCtrl, "Ctrl"},
		{input.ModAlt, "Alt"},
		{input.ModShift, "Shift"},
		{input.ModSuper, "Super"},
	} {
		if p.modifiers&m.mod != 0 {
			parts = append(parts, m.name)
		}
	}

	parts = append(parts, keyName)
	return strings.Join(parts, "+")
}

func (p *Processor) emitEvent(text string, isHeld bool) {
	event := DisplayEvent{
		Text:      text,
		Timestamp: time.Now(),
		IsHeld:    isHeld,
	}

	if !isHeld {
		if len(p.history) >= p.config.HistoryCount {
			p.history = p.history[1:]
		}
		p.history = append(p.history, event)
	}

	select {
	case p.events <- event:
	default:
	}
}
