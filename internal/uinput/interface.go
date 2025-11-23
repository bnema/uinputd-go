package uinput

import "context"

// DeviceInterface defines the interface for virtual input devices.
// This interface allows for mocking in tests while maintaining
// the same behavior as the concrete Device implementation.
type DeviceInterface interface {
	// SendKey sends a single key press and release event
	SendKey(ctx context.Context, keycode uint16) error

	// SendKeyWithModifier sends a key press with a modifier key (e.g., Shift, Ctrl)
	SendKeyWithModifier(ctx context.Context, modifier, keycode uint16) error

	// WriteEvent writes a raw input event to the device
	WriteEvent(event *InputEvent) error

	// Close closes the device and cleans up resources
	Close() error
}

// Compile-time check to ensure Device implements DeviceInterface
var _ DeviceInterface = (*Device)(nil)
