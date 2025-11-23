package layouts

import (
	"context"

	"github.com/bnema/uinputd-go/internal/uinput"
)

// FRLayout implements French AZERTY keyboard layout.
type FRLayout struct {
	baseMappings    map[rune]KeyMapping
	deadKeyRegistry DeadKeyRegistry
	deadKeys        map[rune]KeyMapping
}

// NewFR creates a new French AZERTY layout.
func NewFR() *FRLayout {
	// Build the base mappings by merging common and French-specific mappings
	base := MergeKeymaps(
		CommonMappings,       // Universal: space, tab, enter, €
		frAZERTYLetters,      // AZERTY letter positions
		frNumberRow,          // French number row (shifted)
		frPrecomposedAccents, // Direct keys for é, è, à, ç, ù
		frPunctuation,        // French punctuation layout
		frAltGrSymbols,       // AltGr combinations
		frRemainingSymbols,   // Other French-specific symbols
	)

	return &FRLayout{
		baseMappings:    base,
		deadKeyRegistry: BuildDeadKeyRegistry(),
		deadKeys:        frDeadKeys,
	}
}

// Name returns "fr".
func (l *FRLayout) Name() string {
	return "fr"
}

// CharToKeySequence converts a Unicode character to a sequence of keystrokes.
func (l *FRLayout) CharToKeySequence(ctx context.Context, char rune) ([]KeySequence, error) {
	// First, check if it's a direct mapping
	if mapping, ok := l.baseMappings[char]; ok {
		return []KeySequence{{Keycode: mapping.Keycode, Modifier: mapping.Modifier}}, nil
	}

	// Check if it needs a dead key combination
	if comp, ok := l.deadKeyRegistry[char]; ok {
		// Get the dead key mapping for this layout
		deadKeyMapping, hasDead := l.deadKeys[comp.DeadKey]
		if !hasDead {
			// This layout doesn't have this dead key
			return nil, &ErrCharNotSupported{Char: char, Layout: "fr"}
		}

		// Get the base character mapping
		baseMapping, hasBase := l.baseMappings[comp.BaseChar]
		if !hasBase {
			return nil, &ErrCharNotSupported{Char: char, Layout: "fr"}
		}

		// Return the sequence: dead key, then base character
		return []KeySequence{
			{Keycode: deadKeyMapping.Keycode, Modifier: deadKeyMapping.Modifier},
			{Keycode: baseMapping.Keycode, Modifier: baseMapping.Modifier},
		}, nil
	}

	return nil, &ErrCharNotSupported{Char: char, Layout: "fr"}
}

// frDeadKeys maps dead key symbols to their physical location on French AZERTY keyboard.
var frDeadKeys = map[rune]KeyMapping{
	'^': {Keycode: uinput.KeyLeftBrace, Modifier: ModNone},  // Circumflex
	'¨': {Keycode: uinput.KeyLeftBrace, Modifier: ModShift}, // Diaeresis
}

// frAZERTYLetters contains the AZERTY letter layout.
// Unlike QWERTY, several letters are in different positions:
// - First row: a→Q, z→W
// - Second row: q→A
// - Third row: w→Z
// - m is at Semicolon position
var frAZERTYLetters = map[rune]KeyMapping{
	// First row - AZERTY specific positions
	'a': {Keycode: uinput.KeyQ, Modifier: ModNone},
	'A': {Keycode: uinput.KeyQ, Modifier: ModShift},
	'z': {Keycode: uinput.KeyW, Modifier: ModNone},
	'Z': {Keycode: uinput.KeyW, Modifier: ModShift},

	// First row - same as QWERTY
	'e': {Keycode: uinput.KeyE, Modifier: ModNone},
	'E': {Keycode: uinput.KeyE, Modifier: ModShift},
	'r': {Keycode: uinput.KeyR, Modifier: ModNone},
	'R': {Keycode: uinput.KeyR, Modifier: ModShift},
	't': {Keycode: uinput.KeyT, Modifier: ModNone},
	'T': {Keycode: uinput.KeyT, Modifier: ModShift},
	'y': {Keycode: uinput.KeyY, Modifier: ModNone},
	'Y': {Keycode: uinput.KeyY, Modifier: ModShift},
	'u': {Keycode: uinput.KeyU, Modifier: ModNone},
	'U': {Keycode: uinput.KeyU, Modifier: ModShift},
	'i': {Keycode: uinput.KeyI, Modifier: ModNone},
	'I': {Keycode: uinput.KeyI, Modifier: ModShift},
	'o': {Keycode: uinput.KeyO, Modifier: ModNone},
	'O': {Keycode: uinput.KeyO, Modifier: ModShift},
	'p': {Keycode: uinput.KeyP, Modifier: ModNone},
	'P': {Keycode: uinput.KeyP, Modifier: ModShift},

	// Second row - q is at KeyA position
	'q': {Keycode: uinput.KeyA, Modifier: ModNone},
	'Q': {Keycode: uinput.KeyA, Modifier: ModShift},

	// Second row - same as QWERTY
	's': {Keycode: uinput.KeyS, Modifier: ModNone},
	'S': {Keycode: uinput.KeyS, Modifier: ModShift},
	'd': {Keycode: uinput.KeyD, Modifier: ModNone},
	'D': {Keycode: uinput.KeyD, Modifier: ModShift},
	'f': {Keycode: uinput.KeyF, Modifier: ModNone},
	'F': {Keycode: uinput.KeyF, Modifier: ModShift},
	'g': {Keycode: uinput.KeyG, Modifier: ModNone},
	'G': {Keycode: uinput.KeyG, Modifier: ModShift},
	'h': {Keycode: uinput.KeyH, Modifier: ModNone},
	'H': {Keycode: uinput.KeyH, Modifier: ModShift},
	'j': {Keycode: uinput.KeyJ, Modifier: ModNone},
	'J': {Keycode: uinput.KeyJ, Modifier: ModShift},
	'k': {Keycode: uinput.KeyK, Modifier: ModNone},
	'K': {Keycode: uinput.KeyK, Modifier: ModShift},
	'l': {Keycode: uinput.KeyL, Modifier: ModNone},
	'L': {Keycode: uinput.KeyL, Modifier: ModShift},

	// Second row - m is at Semicolon position
	'm': {Keycode: uinput.KeySemicolon, Modifier: ModNone},
	'M': {Keycode: uinput.KeySemicolon, Modifier: ModShift},

	// Third row - w is at KeyZ position
	'w': {Keycode: uinput.KeyZ, Modifier: ModNone},
	'W': {Keycode: uinput.KeyZ, Modifier: ModShift},

	// Third row - same as QWERTY
	'x': {Keycode: uinput.KeyX, Modifier: ModNone},
	'X': {Keycode: uinput.KeyX, Modifier: ModShift},
	'c': {Keycode: uinput.KeyC, Modifier: ModNone},
	'C': {Keycode: uinput.KeyC, Modifier: ModShift},
	'v': {Keycode: uinput.KeyV, Modifier: ModNone},
	'V': {Keycode: uinput.KeyV, Modifier: ModShift},
	'b': {Keycode: uinput.KeyB, Modifier: ModNone},
	'B': {Keycode: uinput.KeyB, Modifier: ModShift},
	'n': {Keycode: uinput.KeyN, Modifier: ModNone},
	'N': {Keycode: uinput.KeyN, Modifier: ModShift},
}

// frNumberRow contains the French AZERTY number row.
// In French AZERTY, numbers require SHIFT, and symbols are unshifted.
var frNumberRow = map[rune]KeyMapping{
	// Shifted numbers
	'1': {Keycode: uinput.Key1, Modifier: ModShift},
	'2': {Keycode: uinput.Key2, Modifier: ModShift},
	'3': {Keycode: uinput.Key3, Modifier: ModShift},
	'4': {Keycode: uinput.Key4, Modifier: ModShift},
	'5': {Keycode: uinput.Key5, Modifier: ModShift},
	'6': {Keycode: uinput.Key6, Modifier: ModShift},
	'7': {Keycode: uinput.Key7, Modifier: ModShift},
	'8': {Keycode: uinput.Key8, Modifier: ModShift},
	'9': {Keycode: uinput.Key9, Modifier: ModShift},
	'0': {Keycode: uinput.Key0, Modifier: ModShift},

	// Unshifted symbols on number row
	'&': {Keycode: uinput.Key1, Modifier: ModNone},
	// Note: é is in frPrecomposedAccents (Key2, ModNone)
	'"':  {Keycode: uinput.Key3, Modifier: ModNone},
	'\'': {Keycode: uinput.Key4, Modifier: ModNone},
	'(':  {Keycode: uinput.Key5, Modifier: ModNone},
	'-':  {Keycode: uinput.Key6, Modifier: ModNone},
	// Note: è is in frPrecomposedAccents (Key7, ModNone)
	'_': {Keycode: uinput.Key8, Modifier: ModNone},
	// Note: ç is in frPrecomposedAccents (Key9, ModNone)
	// Note: à is in frPrecomposedAccents (Key0, ModNone)
}

// frPrecomposedAccents contains French characters that have dedicated keys
// (not requiring dead key combinations).
var frPrecomposedAccents = map[rune]KeyMapping{
	'é': {Keycode: uinput.Key2, Modifier: ModNone},
	'è': {Keycode: uinput.Key7, Modifier: ModNone},
	'à': {Keycode: uinput.Key0, Modifier: ModNone},
	'ç': {Keycode: uinput.Key9, Modifier: ModNone},
	'ù': {Keycode: uinput.KeyApostrophe, Modifier: ModNone},
}

// frPunctuation contains French punctuation layout.
var frPunctuation = map[rune]KeyMapping{
	',': {Keycode: uinput.KeyM, Modifier: ModNone},
	'?': {Keycode: uinput.KeyM, Modifier: ModShift},
	';': {Keycode: uinput.KeyComma, Modifier: ModNone},
	'.': {Keycode: uinput.KeyComma, Modifier: ModShift},
	':': {Keycode: uinput.KeyDot, Modifier: ModNone},
	'/': {Keycode: uinput.KeyDot, Modifier: ModShift},
	'!': {Keycode: uinput.KeySlash, Modifier: ModNone},
}

// frAltGrSymbols contains symbols accessible with AltGr.
var frAltGrSymbols = map[rune]KeyMapping{
	'[':  {Keycode: uinput.Key5, Modifier: ModAltGr},
	']':  {Keycode: uinput.KeyMinus, Modifier: ModAltGr},
	'{':  {Keycode: uinput.Key4, Modifier: ModAltGr},
	'}':  {Keycode: uinput.KeyEqual, Modifier: ModAltGr},
	'@':  {Keycode: uinput.Key0, Modifier: ModAltGr},
	'#':  {Keycode: uinput.Key3, Modifier: ModAltGr},
	'~':  {Keycode: uinput.Key2, Modifier: ModAltGr},
	'\\': {Keycode: uinput.Key8, Modifier: ModAltGr},
	'|':  {Keycode: uinput.Key6, Modifier: ModAltGr},
	'`':  {Keycode: uinput.Key7, Modifier: ModAltGr},
}

// frRemainingSymbols contains other French-specific symbols.
var frRemainingSymbols = map[rune]KeyMapping{
	'°': {Keycode: uinput.KeyMinus, Modifier: ModShift},
	')': {Keycode: uinput.KeyEqual, Modifier: ModNone},
	'=': {Keycode: uinput.KeyEqual, Modifier: ModShift},
	'^': {Keycode: uinput.KeyLeftBrace, Modifier: ModNone},  // Also a dead key
	'¨': {Keycode: uinput.KeyLeftBrace, Modifier: ModShift}, // Also a dead key
	'$': {Keycode: uinput.KeyRightBrace, Modifier: ModNone},
	'£': {Keycode: uinput.KeyRightBrace, Modifier: ModShift},
	'*': {Keycode: uinput.KeyBackslash, Modifier: ModNone},
	'µ': {Keycode: uinput.KeyBackslash, Modifier: ModShift},
	'%': {Keycode: uinput.KeyApostrophe, Modifier: ModShift},
	'<': {Keycode: uinput.KeyGrave, Modifier: ModNone},
	'>': {Keycode: uinput.KeyGrave, Modifier: ModShift},
}
