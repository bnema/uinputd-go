package uinput

import (
	"context"
	"encoding/binary"
	"fmt"
	"time"

	"golang.org/x/sys/unix"
)

// InputEvent represents a Linux input event structure.
// See: <linux/input.h> struct input_event
type InputEvent struct {
	Time  unix.Timeval
	Type  uint16
	Code  uint16
	Value int32
}

// Marshal converts InputEvent to bytes for writing to /dev/uinput.
func (e *InputEvent) Marshal() []byte {
	buf := make([]byte, 24) // sizeof(struct input_event)

	// Timeval (8 + 8 = 16 bytes on 64-bit)
	binary.LittleEndian.PutUint64(buf[0:8], uint64(e.Time.Sec))
	binary.LittleEndian.PutUint64(buf[8:16], uint64(e.Time.Usec))

	// Type, Code, Value
	binary.LittleEndian.PutUint16(buf[16:18], e.Type)
	binary.LittleEndian.PutUint16(buf[18:20], e.Code)
	binary.LittleEndian.PutUint32(buf[20:24], uint32(e.Value))

	return buf
}

// NewEvent creates a new InputEvent with current timestamp.
func NewEvent(typ, code uint16, value int32) *InputEvent {
	now := time.Now()
	return &InputEvent{
		Time: unix.Timeval{
			Sec:  now.Unix(),
			Usec: int64(now.Nanosecond() / 1000),
		},
		Type:  typ,
		Code:  code,
		Value: value,
	}
}

// NewKeyEvent creates a key press/release event.
func NewKeyEvent(keycode uint16, pressed bool) *InputEvent {
	value := int32(KeyRelease)
	if pressed {
		value = int32(KeyPress)
	}
	return NewEvent(EvKey, keycode, value)
}

// NewSynEvent creates a synchronization event (SYN_REPORT).
func NewSynEvent() *InputEvent {
	return NewEvent(EvSyn, SynReport, 0)
}

// SendKey sends a key press and release sequence.
// This is the most common operation: press key -> sync -> release key -> sync.
func (d *Device) SendKey(ctx context.Context, keycode uint16) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Press key
	if err := d.WriteEvent(NewKeyEvent(keycode, true)); err != nil {
		return fmt.Errorf("key press: %w", err)
	}
	if err := d.WriteEvent(NewSynEvent()); err != nil {
		return fmt.Errorf("syn after press: %w", err)
	}

	// Release key
	if err := d.WriteEvent(NewKeyEvent(keycode, false)); err != nil {
		return fmt.Errorf("key release: %w", err)
	}
	if err := d.WriteEvent(NewSynEvent()); err != nil {
		return fmt.Errorf("syn after release: %w", err)
	}

	return nil
}

// SendKeyWithModifier sends a key with a modifier (e.g., Shift+A).
func (d *Device) SendKeyWithModifier(ctx context.Context, modifier, keycode uint16) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Press modifier
	if err := d.WriteEvent(NewKeyEvent(modifier, true)); err != nil {
		return fmt.Errorf("modifier press: %w", err)
	}
	if err := d.WriteEvent(NewSynEvent()); err != nil {
		return fmt.Errorf("syn after modifier: %w", err)
	}

	// Press key
	if err := d.WriteEvent(NewKeyEvent(keycode, true)); err != nil {
		return fmt.Errorf("key press: %w", err)
	}
	if err := d.WriteEvent(NewSynEvent()); err != nil {
		return fmt.Errorf("syn after press: %w", err)
	}

	// Release key
	if err := d.WriteEvent(NewKeyEvent(keycode, false)); err != nil {
		return fmt.Errorf("key release: %w", err)
	}
	if err := d.WriteEvent(NewSynEvent()); err != nil {
		return fmt.Errorf("syn after release: %w", err)
	}

	// Release modifier
	if err := d.WriteEvent(NewKeyEvent(modifier, false)); err != nil {
		return fmt.Errorf("modifier release: %w", err)
	}
	if err := d.WriteEvent(NewSynEvent()); err != nil {
		return fmt.Errorf("syn after modifier release: %w", err)
	}

	return nil
}

// WriteEvent writes a single InputEvent to the uinput device.
func (d *Device) WriteEvent(event *InputEvent) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.fd == nil {
		return fmt.Errorf("device not open")
	}

	data := event.Marshal()
	n, err := d.fd.Write(data)
	if err != nil {
		return fmt.Errorf("write event: %w", err)
	}
	if n != len(data) {
		return fmt.Errorf("incomplete write: %d/%d bytes", n, len(data))
	}

	return nil
}
