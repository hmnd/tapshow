package input

import (
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"
	"unsafe"
)

const (
	EV_SYN         = 0x00
	EV_KEY         = 0x01
	inputEventSize = 24
)

type inputEvent struct {
	Time  syscall.Timeval
	Type  uint16
	Code  uint16
	Value int32
}

type Reader struct {
	devices []*os.File
	events  chan KeyEvent
	done    chan struct{}
}

func NewReader() *Reader {
	return &Reader{
		events: make(chan KeyEvent, 100),
		done:   make(chan struct{}),
	}
}

func (r *Reader) Events() <-chan KeyEvent {
	return r.events
}

func (r *Reader) Start() error {
	keyboards, err := findKeyboards()
	if err != nil {
		return fmt.Errorf("failed to find keyboards: %w", err)
	}

	if len(keyboards) == 0 {
		return fmt.Errorf("no keyboards found - ensure you're in the 'input' group")
	}

	for _, path := range keyboards {
		f, err := os.Open(path)
		if err != nil {
			continue
		}
		r.devices = append(r.devices, f)
		go r.readDevice(f)
	}

	if len(r.devices) == 0 {
		return fmt.Errorf("could not open any keyboard devices - check 'input' group membership")
	}

	return nil
}

func (r *Reader) Stop() {
	close(r.done)
	for _, f := range r.devices {
		f.Close()
	}
}

func (r *Reader) readDevice(f *os.File) {
	buf := make([]byte, inputEventSize)

	for {
		select {
		case <-r.done:
			return
		default:
		}

		n, err := f.Read(buf)
		if err != nil {
			return
		}
		if n != inputEventSize {
			continue
		}

		ev := parseInputEvent(buf)
		if ev.Type != EV_KEY {
			continue
		}

		var state KeyState
		switch ev.Value {
		case 0:
			state = KeyReleased
		case 1:
			state = KeyPressed
		case 2:
			state = KeyHeld
		default:
			continue
		}

		name := GetKeyName(ev.Code)
		if name == "" {
			continue
		}

		keyEvent := KeyEvent{
			Code:      ev.Code,
			Name:      name,
			State:     state,
			Timestamp: time.Now(),
		}

		select {
		case r.events <- keyEvent:
		case <-r.done:
			return
		default:
		}
	}
}

func parseInputEvent(buf []byte) inputEvent {
	return inputEvent{
		Time: syscall.Timeval{
			Sec:  int64(binary.LittleEndian.Uint64(buf[0:8])),
			Usec: int64(binary.LittleEndian.Uint64(buf[8:16])),
		},
		Type:  binary.LittleEndian.Uint16(buf[16:18]),
		Code:  binary.LittleEndian.Uint16(buf[18:20]),
		Value: int32(binary.LittleEndian.Uint32(buf[20:24])),
	}
}

func findKeyboards() ([]string, error) {
	var keyboards []string

	matches, err := filepath.Glob("/dev/input/event*")
	if err != nil {
		return nil, err
	}

	for _, path := range matches {
		if isKeyboard(path) {
			keyboards = append(keyboards, path)
		}
	}

	return keyboards, nil
}

func isKeyboard(path string) bool {
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	defer f.Close()

	name := make([]byte, 256)
	_, _, errno := syscall.Syscall(
		syscall.SYS_IOCTL,
		f.Fd(),
		uintptr(0x80ff4506), // EVIOCGNAME(256)
		uintptr(unsafe.Pointer(&name[0])),
	)
	if errno != 0 {
		return false
	}

	nameStr := strings.ToLower(string(name))
	if strings.Contains(nameStr, "mouse") ||
		strings.Contains(nameStr, "touchpad") ||
		strings.Contains(nameStr, "trackpad") ||
		strings.Contains(nameStr, "trackpoint") {
		return false
	}

	evBits := make([]byte, (EV_KEY+7)/8+1)
	_, _, errno = syscall.Syscall(
		syscall.SYS_IOCTL,
		f.Fd(),
		uintptr(0x80044520), // EVIOCGBIT(0, size)
		uintptr(unsafe.Pointer(&evBits[0])),
	)
	if errno != 0 {
		return false
	}

	if evBits[EV_KEY/8]&(1<<(EV_KEY%8)) == 0 {
		return false
	}

	keyBits := make([]byte, (KEY_Z+7)/8+1)
	_, _, errno = syscall.Syscall(
		syscall.SYS_IOCTL,
		f.Fd(),
		uintptr(0x80044521), // EVIOCGBIT(EV_KEY, size)
		uintptr(unsafe.Pointer(&keyBits[0])),
	)
	if errno != 0 {
		return false
	}

	for key := uint16(KEY_Q); key <= KEY_P; key++ {
		if keyBits[key/8]&(1<<(key%8)) != 0 {
			return true
		}
	}
	return false
}
