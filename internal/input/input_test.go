package input

import (
	"testing"
)

func TestGetKeyName(t *testing.T) {
	tests := []struct {
		code     uint16
		expected string
	}{
		{KEY_A, "A"},
		{KEY_B, "B"},
		{KEY_Z, "Z"},
		{KEY_1, "1"},
		{KEY_0, "0"},
		{KEY_SPACE, "Space"},
		{KEY_ENTER, "Enter"},
		{KEY_TAB, "Tab"},
		{KEY_BACKSPACE, "Backspace"},
		{KEY_ESC, "Esc"},
		{KEY_LEFTCTRL, "Ctrl"},
		{KEY_RIGHTCTRL, "Ctrl"},
		{KEY_LEFTSHIFT, "Shift"},
		{KEY_RIGHTSHIFT, "Shift"},
		{KEY_LEFTALT, "Alt"},
		{KEY_RIGHTALT, "Alt"},
		{KEY_LEFTMETA, "Super"},
		{KEY_RIGHTMETA, "Super"},
		{KEY_F1, "F1"},
		{KEY_F12, "F12"},
		{0xFFFF, ""}, // Unknown key
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			got := GetKeyName(tt.code)
			if got != tt.expected {
				t.Errorf("GetKeyName(%d) = %q, want %q", tt.code, got, tt.expected)
			}
		})
	}
}

func TestIsModifier(t *testing.T) {
	modifiers := []uint16{
		KEY_LEFTCTRL, KEY_RIGHTCTRL,
		KEY_LEFTSHIFT, KEY_RIGHTSHIFT,
		KEY_LEFTALT, KEY_RIGHTALT,
		KEY_LEFTMETA, KEY_RIGHTMETA,
	}

	for _, code := range modifiers {
		if !IsModifier(code) {
			t.Errorf("IsModifier(%d) = false, want true", code)
		}
	}

	nonModifiers := []uint16{KEY_A, KEY_SPACE, KEY_ENTER, KEY_1}
	for _, code := range nonModifiers {
		if IsModifier(code) {
			t.Errorf("IsModifier(%d) = true, want false", code)
		}
	}
}

func TestGetModifier(t *testing.T) {
	tests := []struct {
		code     uint16
		expected Modifier
	}{
		{KEY_LEFTCTRL, ModCtrl},
		{KEY_RIGHTCTRL, ModCtrl},
		{KEY_LEFTSHIFT, ModShift},
		{KEY_RIGHTSHIFT, ModShift},
		{KEY_LEFTALT, ModAlt},
		{KEY_RIGHTALT, ModAlt},
		{KEY_LEFTMETA, ModSuper},
		{KEY_RIGHTMETA, ModSuper},
		{KEY_A, ModNone},
		{KEY_SPACE, ModNone},
	}

	for _, tt := range tests {
		t.Run(GetKeyName(tt.code), func(t *testing.T) {
			got := GetModifier(tt.code)
			if got != tt.expected {
				t.Errorf("GetModifier(%d) = %d, want %d", tt.code, got, tt.expected)
			}
		})
	}
}

func TestModifierFlags(t *testing.T) {
	// Test that modifier flags can be combined
	var mods Modifier = ModCtrl | ModShift

	if mods&ModCtrl == 0 {
		t.Error("Combined modifiers should include ModCtrl")
	}
	if mods&ModShift == 0 {
		t.Error("Combined modifiers should include ModShift")
	}
	if mods&ModAlt != 0 {
		t.Error("Combined modifiers should not include ModAlt")
	}
}
