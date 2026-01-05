package processor

import (
	"testing"
	"time"

	"github.com/tapshow/tapshow/internal/input"
)

func TestProcessor_CombineModifiers(t *testing.T) {
	cfg := DefaultConfig()
	cfg.ShowHeldKeys = false

	proc := New(cfg)
	events := make(chan input.KeyEvent, 10)

	go proc.Process(events)
	defer proc.Stop()

	events <- input.KeyEvent{
		Code:  input.KEY_LEFTCTRL,
		Name:  "Ctrl",
		State: input.KeyPressed,
	}
	events <- input.KeyEvent{
		Code:  input.KEY_A,
		Name:  "A",
		State: input.KeyPressed,
	}

	time.Sleep(50 * time.Millisecond)

	select {
	case event := <-proc.Events():
		if event.Text != "Ctrl+A" {
			t.Errorf("Expected 'Ctrl+A', got %q", event.Text)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Timeout waiting for display event")
	}
}

func TestProcessor_History(t *testing.T) {
	cfg := DefaultConfig()
	cfg.CombineModifiers = false
	cfg.ShowHeldKeys = false
	cfg.HistoryCount = 3

	proc := New(cfg)
	events := make(chan input.KeyEvent, 10)

	go proc.Process(events)
	defer proc.Stop()

	for _, key := range []string{"A", "B", "C", "D"} {
		events <- input.KeyEvent{
			Code:  input.KEY_A,
			Name:  key,
			State: input.KeyPressed,
		}
		time.Sleep(10 * time.Millisecond)
	}

	time.Sleep(50 * time.Millisecond)

	for len(proc.Events()) > 0 {
		<-proc.Events()
	}

	history := proc.History()
	if len(history) != 3 {
		t.Errorf("Expected history length 3, got %d", len(history))
	}

	expected := []string{"B", "C", "D"}
	for i, h := range history {
		if h.Text != expected[i] {
			t.Errorf("History[%d] = %q, want %q", i, h.Text, expected[i])
		}
	}
}

func TestProcessor_ExcludedKeys(t *testing.T) {
	cfg := DefaultConfig()
	cfg.CombineModifiers = false
	cfg.ShowHeldKeys = false
	cfg.ExcludedKeys = []string{"CapsLock"}

	proc := New(cfg)
	events := make(chan input.KeyEvent, 10)

	go proc.Process(events)
	defer proc.Stop()

	events <- input.KeyEvent{
		Code:  input.KEY_CAPSLOCK,
		Name:  "CapsLock",
		State: input.KeyPressed,
	}

	time.Sleep(50 * time.Millisecond)

	select {
	case event := <-proc.Events():
		t.Errorf("Should not receive excluded key, got %q", event.Text)
	case <-time.After(50 * time.Millisecond):
	}
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if !cfg.CombineModifiers {
		t.Error("Default CombineModifiers should be true")
	}
	if cfg.ShowModifierOnly {
		t.Error("Default ShowModifierOnly should be false")
	}
	if cfg.HistoryCount != 4 {
		t.Errorf("Default HistoryCount = %d, want 4", cfg.HistoryCount)
	}
}
