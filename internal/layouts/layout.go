package layouts

import (
	"context"
	"fmt"
)

// KeySequence represents a single keystroke with its modifiers.
// For simple characters, a sequence contains one keystroke.
// For dead key combinations, a sequence contains multiple keystrokes
// (e.g., circumflex key, then vowel key).
type KeySequence struct {
	Keycode  uint16
	Modifier Modifier
}

// Layout defines the interface for keyboard layout implementations.
// Each layout maps Unicode characters to Linux keycodes with appropriate modifiers.
type Layout interface {
	// Name returns the layout identifier (e.g., "us", "fr").
	Name() string

	// CharToKeySequence converts a Unicode character to a sequence of keystrokes.
	// For simple characters, returns a single-element slice.
	// For dead key combinations (like ô = ^ + o), returns multiple keystrokes.
	CharToKeySequence(ctx context.Context, char rune) ([]KeySequence, error)
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
