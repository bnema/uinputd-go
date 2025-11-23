package layouts

import (
	"context"
	"fmt"
)

// Layout defines the interface for keyboard layout implementations.
// Each layout maps Unicode characters to Linux keycodes with appropriate modifiers.
type Layout interface {
	// Name returns the layout identifier (e.g., "us", "fr").
	Name() string

	// CharToKeycode converts a Unicode character to a keycode and modifier flags.
	// Returns the keycode, whether Shift is needed, whether AltGr is needed, and any error.
	CharToKeycode(ctx context.Context, char rune) (keycode uint16, shift, altGr bool, err error)
}

// Modifier represents keyboard modifiers.
type Modifier uint8

const (
	ModNone  Modifier = 0
	ModShift Modifier = 1 << 0
	ModAltGr Modifier = 1 << 1
	ModCtrl  Modifier = 1 << 2
	ModAlt   Modifier = 1 << 3
)

// KeyMapping represents a character-to-keycode mapping.
type KeyMapping struct {
	Keycode  uint16
	Modifier Modifier
}

// ErrCharNotSupported is returned when a character has no mapping in the layout.
type ErrCharNotSupported struct {
	Char   rune
	Layout string
}

func (e *ErrCharNotSupported) Error() string {
	return fmt.Sprintf("character %q (U+%04X) not supported in %s layout", e.Char, e.Char, e.Layout)
}
