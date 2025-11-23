package layouts

import (
	"context"

	"github.com/bnema/uinputd-go/internal/uinput"
)

// DELayout implements German QWERTZ keyboard layout.
type DELayout struct {
	baseMappings    map[rune]KeyMapping
	deadKeyRegistry DeadKeyRegistry
	deadKeys        map[rune]KeyMapping
}

// NewDE creates a new German QWERTZ layout.
func NewDE() *DELayout {
	// Build the base mappings by merging shared and German-specific mappings
	base := MergeKeymaps(
		CommonMappings,         // Universal: space, tab, enter, €
		StandardNumberMappings, // Numbers 0-9 without shift
		deQWERTZLetters,        // QWERTZ letter layout (y/z swapped)
		deShiftedSymbols,       // German shifted symbols
		deUmlauts,              // German umlauts (ü, ö, ä, ß)
		dePunctuation,          // German punctuation
		deSymbols,              // German symbols
		deAltGrSymbols,         // German AltGr combinations
		deSpecialKeys,          // German special keys
	)

	return &DELayout{
		baseMappings:    base,
		deadKeyRegistry: BuildDeadKeyRegistry(),
		deadKeys:        deDeadKeys,
	}
}

// Name returns "de".
func (l *DELayout) Name() string {
	return "de"
}

// CharToKeySequence converts a Unicode character to a sequence of keystrokes.
func (l *DELayout) CharToKeySequence(ctx context.Context, char rune) ([]KeySequence, error) {
	// First, check if it's a direct mapping
	if mapping, ok := l.baseMappings[char]; ok {
		return []KeySequence{{Keycode: mapping.Keycode, Modifier: mapping.Modifier}}, nil
	}

	// Check if it needs a dead key combination
	if comp, ok := l.deadKeyRegistry[char]; ok {
		deadKeyMapping, hasDead := l.deadKeys[comp.DeadKey]
		if !hasDead {
			return nil, &ErrCharNotSupported{Char: char, Layout: "de"}
		}

		baseMapping, hasBase := l.baseMappings[comp.BaseChar]
		if !hasBase {
			return nil, &ErrCharNotSupported{Char: char, Layout: "de"}
		}

		return []KeySequence{
			{Keycode: deadKeyMapping.Keycode, Modifier: deadKeyMapping.Modifier},
			{Keycode: baseMapping.Keycode, Modifier: baseMapping.Modifier},
		}, nil
	}

	return nil, &ErrCharNotSupported{Char: char, Layout: "de"}
}

// deDeadKeys maps dead key symbols to their physical location on German QWERTZ keyboard.
var deDeadKeys = map[rune]KeyMapping{
	'^': {Keycode: uinput.KeyGrave, Modifier: ModNone},  // Circumflex
	'´': {Keycode: uinput.KeyEqual, Modifier: ModNone},  // Acute
	'`': {Keycode: uinput.KeyEqual, Modifier: ModShift}, // Grave
}

// deQWERTZLetters contains the QWERTZ letter layout.
// In QWERTZ, Y and Z are swapped compared to QWERTY.
var deQWERTZLetters = map[rune]KeyMapping{
	// First row - standard except Y/Z swap
	'q': {Keycode: uinput.KeyQ, Modifier: ModNone},
	'Q': {Keycode: uinput.KeyQ, Modifier: ModShift},
	'w': {Keycode: uinput.KeyW, Modifier: ModNone},
	'W': {Keycode: uinput.KeyW, Modifier: ModShift},
	'e': {Keycode: uinput.KeyE, Modifier: ModNone},
	'E': {Keycode: uinput.KeyE, Modifier: ModShift},
	'r': {Keycode: uinput.KeyR, Modifier: ModNone},
	'R': {Keycode: uinput.KeyR, Modifier: ModShift},
	't': {Keycode: uinput.KeyT, Modifier: ModNone},
	'T': {Keycode: uinput.KeyT, Modifier: ModShift},
	'z': {Keycode: uinput.KeyY, Modifier: ModNone}, // Z on Y key
	'Z': {Keycode: uinput.KeyY, Modifier: ModShift},
	'u': {Keycode: uinput.KeyU, Modifier: ModNone},
	'U': {Keycode: uinput.KeyU, Modifier: ModShift},
	'i': {Keycode: uinput.KeyI, Modifier: ModNone},
	'I': {Keycode: uinput.KeyI, Modifier: ModShift},
	'o': {Keycode: uinput.KeyO, Modifier: ModNone},
	'O': {Keycode: uinput.KeyO, Modifier: ModShift},
	'p': {Keycode: uinput.KeyP, Modifier: ModNone},
	'P': {Keycode: uinput.KeyP, Modifier: ModShift},

	// Second row - standard
	'a': {Keycode: uinput.KeyA, Modifier: ModNone},
	'A': {Keycode: uinput.KeyA, Modifier: ModShift},
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

	// Third row - Y on Z key
	'y': {Keycode: uinput.KeyZ, Modifier: ModNone}, // Y on Z key
	'Y': {Keycode: uinput.KeyZ, Modifier: ModShift},
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
	'm': {Keycode: uinput.KeyM, Modifier: ModNone},
	'M': {Keycode: uinput.KeyM, Modifier: ModShift},
}

// deShiftedSymbols contains German shifted symbols on number row.
var deShiftedSymbols = map[rune]KeyMapping{
	'!': {Keycode: uinput.Key1, Modifier: ModShift},
	'"': {Keycode: uinput.Key2, Modifier: ModShift},
	'§': {Keycode: uinput.Key3, Modifier: ModShift},
	'$': {Keycode: uinput.Key4, Modifier: ModShift},
	'%': {Keycode: uinput.Key5, Modifier: ModShift},
	'&': {Keycode: uinput.Key6, Modifier: ModShift},
	'/': {Keycode: uinput.Key7, Modifier: ModShift},
	'(': {Keycode: uinput.Key8, Modifier: ModShift},
	')': {Keycode: uinput.Key9, Modifier: ModShift},
	'=': {Keycode: uinput.Key0, Modifier: ModShift},
}

// deUmlauts contains German umlauts and special characters with dedicated keys.
var deUmlauts = map[rune]KeyMapping{
	'ü': {Keycode: uinput.KeyLeftBrace, Modifier: ModNone},
	'Ü': {Keycode: uinput.KeyLeftBrace, Modifier: ModShift},
	'ö': {Keycode: uinput.KeySemicolon, Modifier: ModNone},
	'Ö': {Keycode: uinput.KeySemicolon, Modifier: ModShift},
	'ä': {Keycode: uinput.KeyApostrophe, Modifier: ModNone},
	'Ä': {Keycode: uinput.KeyApostrophe, Modifier: ModShift},
	'ß': {Keycode: uinput.KeyMinus, Modifier: ModNone},
	'?': {Keycode: uinput.KeyMinus, Modifier: ModShift},
}

// dePunctuation contains German punctuation.
var dePunctuation = map[rune]KeyMapping{
	',': {Keycode: uinput.KeyComma, Modifier: ModNone},
	';': {Keycode: uinput.KeyComma, Modifier: ModShift},
	'.': {Keycode: uinput.KeyDot, Modifier: ModNone},
	':': {Keycode: uinput.KeyDot, Modifier: ModShift},
	'-': {Keycode: uinput.KeySlash, Modifier: ModNone},
	'_': {Keycode: uinput.KeySlash, Modifier: ModShift},
}

// deSymbols contains German symbols.
var deSymbols = map[rune]KeyMapping{
	'+':  {Keycode: uinput.KeyRightBrace, Modifier: ModNone},
	'*':  {Keycode: uinput.KeyRightBrace, Modifier: ModShift},
	'#':  {Keycode: uinput.KeyBackslash, Modifier: ModNone},
	'\'': {Keycode: uinput.KeyBackslash, Modifier: ModShift},
	'°':  {Keycode: uinput.KeyGrave, Modifier: ModShift}, // Also position of ^ dead key
}

// deAltGrSymbols contains German AltGr combinations.
var deAltGrSymbols = map[rune]KeyMapping{
	'@':  {Keycode: uinput.KeyQ, Modifier: ModAltGr},
	'~':  {Keycode: uinput.KeyRightBrace, Modifier: ModAltGr},
	'|':  {Keycode: uinput.KeyGrave, Modifier: ModAltGr},
	'{':  {Keycode: uinput.Key7, Modifier: ModAltGr},
	'[':  {Keycode: uinput.Key8, Modifier: ModAltGr},
	']':  {Keycode: uinput.Key9, Modifier: ModAltGr},
	'}':  {Keycode: uinput.Key0, Modifier: ModAltGr},
	'\\': {Keycode: uinput.KeyMinus, Modifier: ModAltGr},
}

// deSpecialKeys contains German special keys.
var deSpecialKeys = map[rune]KeyMapping{
	'<': {Keycode: uinput.Key102ND, Modifier: ModNone},
	'>': {Keycode: uinput.Key102ND, Modifier: ModShift},
}
