package input

import "time"

type KeyState int

const (
	KeyReleased KeyState = iota
	KeyPressed
	KeyHeld
)

type KeyEvent struct {
	Code      uint16
	Name      string
	State     KeyState
	Timestamp time.Time
}

type Modifier uint8

const (
	ModNone Modifier = 0
	ModCtrl Modifier = 1 << iota
	ModShift
	ModAlt
	ModSuper
)

func IsModifier(code uint16) bool {
	switch code {
	case KEY_LEFTCTRL, KEY_RIGHTCTRL,
		KEY_LEFTSHIFT, KEY_RIGHTSHIFT,
		KEY_LEFTALT, KEY_RIGHTALT,
		KEY_LEFTMETA, KEY_RIGHTMETA:
		return true
	}
	return false
}

func GetModifier(code uint16) Modifier {
	switch code {
	case KEY_LEFTCTRL, KEY_RIGHTCTRL:
		return ModCtrl
	case KEY_LEFTSHIFT, KEY_RIGHTSHIFT:
		return ModShift
	case KEY_LEFTALT, KEY_RIGHTALT:
		return ModAlt
	case KEY_LEFTMETA, KEY_RIGHTMETA:
		return ModSuper
	}
	return ModNone
}
