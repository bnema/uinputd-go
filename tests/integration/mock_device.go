package integration

import (
	"context"
	"fmt"
	"sync"

	"github.com/bnema/uinputd-go/internal/uinput"
)

// MockUinputDevice is a mock implementation of uinput.DeviceInterface
// that records events instead of writing to /dev/uinput.
// This allows testing the full server stack without requiring actual
// uinput device permissions.
type MockUinputDevice struct {
	mu     sync.Mutex
	events []*uinput.InputEvent
	closed bool
}

// NewMockUinputDevice creates a new mock uinput device.
func NewMockUinputDevice() *MockUinputDevice {
	return &MockUinputDevice{
		events: make([]*uinput.InputEvent, 0),
	}
}

// SendKey implements uinput.DeviceInterface.
func (m *MockUinputDevice) SendKey(ctx context.Context, keycode uint16) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if m.closed {
		return fmt.Errorf("device closed")
	}

	// Press key
	m.events = append(m.events, uinput.NewKeyEvent(keycode, true))
	m.events = append(m.events, uinput.NewSynEvent())

	// Release key
	m.events = append(m.events, uinput.NewKeyEvent(keycode, false))
	m.events = append(m.events, uinput.NewSynEvent())

	return nil
}

// SendKeyWithModifier implements uinput.DeviceInterface.
func (m *MockUinputDevice) SendKeyWithModifier(ctx context.Context, modifier, keycode uint16) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if m.closed {
		return fmt.Errorf("device closed")
	}

	// Press modifier
	m.events = append(m.events, uinput.NewKeyEvent(modifier, true))
	m.events = append(m.events, uinput.NewSynEvent())

	// Press key
	m.events = append(m.events, uinput.NewKeyEvent(keycode, true))
	m.events = append(m.events, uinput.NewSynEvent())

	// Release key
	m.events = append(m.events, uinput.NewKeyEvent(keycode, false))
	m.events = append(m.events, uinput.NewSynEvent())

	// Release modifier
	m.events = append(m.events, uinput.NewKeyEvent(modifier, false))
	m.events = append(m.events, uinput.NewSynEvent())

	return nil
}

// WriteEvent implements uinput.DeviceInterface.
func (m *MockUinputDevice) WriteEvent(event *uinput.InputEvent) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.closed {
		return fmt.Errorf("device closed")
	}

	// Create a copy of the event to avoid mutation
	eventCopy := *event
	m.events = append(m.events, &eventCopy)
	return nil
}

// Close implements uinput.DeviceInterface.
func (m *MockUinputDevice) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.closed = true
	return nil
}

// GetEvents returns a copy of all recorded events.
func (m *MockUinputDevice) GetEvents() []*uinput.InputEvent {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Return a copy to avoid race conditions
	events := make([]*uinput.InputEvent, len(m.events))
	copy(events, m.events)
	return events
}

// Reset clears all recorded events.
func (m *MockUinputDevice) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.events = make([]*uinput.InputEvent, 0)
}

// GetEventCount returns the number of recorded events.
func (m *MockUinputDevice) GetEventCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()

	return len(m.events)
}

// EventSequence represents an expected key event sequence.
type EventSequence struct {
	Keycode  uint16
	Pressed  bool
	IsSyn    bool
	Modifier bool // True if this is a modifier key
}

// VerifyEventSequence checks if the recorded events match the expected sequence.
// Returns an error describing the mismatch if verification fails.
func (m *MockUinputDevice) VerifyEventSequence(expected []EventSequence) error {
	events := m.GetEvents()

	if len(events) != len(expected) {
		return fmt.Errorf("event count mismatch: got %d events, want %d", len(events), len(expected))
	}

	for i, exp := range expected {
		event := events[i]

		if exp.IsSyn {
			// Expect a SYN event
			if event.Type != uinput.EvSyn {
				return fmt.Errorf("event[%d]: expected SYN event, got type=%d code=%d", i, event.Type, event.Code)
			}
		} else {
			// Expect a KEY event
			if event.Type != uinput.EvKey {
				return fmt.Errorf("event[%d]: expected KEY event, got type=%d", i, event.Type)
			}

			if event.Code != exp.Keycode {
				return fmt.Errorf("event[%d]: expected keycode %d, got %d", i, exp.Keycode, event.Code)
			}

			expectedValue := int32(uinput.KeyRelease)
			if exp.Pressed {
				expectedValue = int32(uinput.KeyPress)
			}

			if event.Value != expectedValue {
				return fmt.Errorf("event[%d]: expected value %d (pressed=%v), got %d", i, expectedValue, exp.Pressed, event.Value)
			}
		}
	}

	return nil
}

// GetKeyPressSequence returns a simplified view of key presses (ignoring SYN events).
// This is useful for debugging test failures.
func (m *MockUinputDevice) GetKeyPressSequence() []string {
	events := m.GetEvents()
	sequence := make([]string, 0)

	for _, event := range events {
		if event.Type == uinput.EvKey {
			state := "release"
			if event.Value == int32(uinput.KeyPress) {
				state = "press"
			}
			sequence = append(sequence, fmt.Sprintf("%s(%d)", state, event.Code))
		}
	}

	return sequence
}
