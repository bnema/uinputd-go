package layouts

import (
	"context"

	"github.com/bnema/uinputd-go/internal/uinput"
)

// ESLayout implements Spanish QWERTY keyboard layout.
type ESLayout struct {
	baseMappings    map[rune]KeyMapping
	deadKeyRegistry DeadKeyRegistry
	deadKeys        map[rune]KeyMapping
}

// NewES creates a new Spanish QWERTY layout.
func NewES() *ESLayout {
	// Build the base mappings by merging shared and Spanish-specific mappings
	base := MergeKeymaps(
		CommonMappings,         // Universal: space, tab, enter, €
		QWERTYBaseMappings,     // Standard QWERTY letter positions
		StandardNumberMappings, // Numbers 0-9 without shift
		esShiftedSymbols,       // Spanish shifted symbols
		esPrecomposedAccents,   // Direct keys for ñ, á, ç
		esPunctuation,          // Spanish punctuation
		esSymbols,              // Spanish symbols
		esAltGrSymbols,         // Spanish AltGr combinations
		esSpecialKeys,          // Spanish special keys
	)

	return &ESLayout{
		baseMappings:    base,
		deadKeyRegistry: BuildDeadKeyRegistry(),
		deadKeys:        esDeadKeys,
	}
}

// Name returns "es".
func (l *ESLayout) Name() string {
	return "es"
}

// CharToKeySequence converts a Unicode character to a sequence of keystrokes.
func (l *ESLayout) CharToKeySequence(ctx context.Context, char rune) ([]KeySequence, error) {
	// First, check if it's a direct mapping
	if mapping, ok := l.baseMappings[char]; ok {
		return []KeySequence{{Keycode: mapping.Keycode, Modifier: mapping.Modifier}}, nil
	}

	// Check if it needs a dead key combination
	if comp, ok := l.deadKeyRegistry[char]; ok {
		deadKeyMapping, hasDead := l.deadKeys[comp.DeadKey]
		if !hasDead {
			return nil, &ErrCharNotSupported{Char: char, Layout: "es"}
		}

		baseMapping, hasBase := l.baseMappings[comp.BaseChar]
		if !hasBase {
			return nil, &ErrCharNotSupported{Char: char, Layout: "es"}
		}

		return []KeySequence{
			{Keycode: deadKeyMapping.Keycode, Modifier: deadKeyMapping.Modifier},
			{Keycode: baseMapping.Keycode, Modifier: baseMapping.Modifier},
		}, nil
	}

	return nil, &ErrCharNotSupported{Char: char, Layout: "es"}
}

// esDeadKeys maps dead key symbols to their physical location on Spanish keyboard.
var esDeadKeys = map[rune]KeyMapping{
	'`': {Keycode: uinput.KeyLeftBrace, Modifier: ModNone},  // Grave
	'^': {Keycode: uinput.KeyLeftBrace, Modifier: ModShift}, // Circumflex
}

// esShiftedSymbols contains Spanish shifted symbols on number row.
var esShiftedSymbols = map[rune]KeyMapping{
	'!': {Keycode: uinput.Key1, Modifier: ModShift},
	'"': {Keycode: uinput.Key2, Modifier: ModShift},
	'·': {Keycode: uinput.Key3, Modifier: ModShift}, // Middle dot
	'$': {Keycode: uinput.Key4, Modifier: ModShift},
	'%': {Keycode: uinput.Key5, Modifier: ModShift},
	'&': {Keycode: uinput.Key6, Modifier: ModShift},
	'/': {Keycode: uinput.Key7, Modifier: ModShift},
	'(': {Keycode: uinput.Key8, Modifier: ModShift},
	')': {Keycode: uinput.Key9, Modifier: ModShift},
	'=': {Keycode: uinput.Key0, Modifier: ModShift},
}

// esPrecomposedAccents contains Spanish characters with dedicated keys.
var esPrecomposedAccents = map[rune]KeyMapping{
	'ñ': {Keycode: uinput.KeySemicolon, Modifier: ModNone},
	'Ñ': {Keycode: uinput.KeySemicolon, Modifier: ModShift},
	'á': {Keycode: uinput.KeyApostrophe, Modifier: ModNone},
	'Á': {Keycode: uinput.KeyApostrophe, Modifier: ModShift},
	'ç': {Keycode: uinput.KeyBackslash, Modifier: ModNone},
	'Ç': {Keycode: uinput.KeyBackslash, Modifier: ModShift},
}

// esPunctuation contains Spanish punctuation.
var esPunctuation = map[rune]KeyMapping{
	',': {Keycode: uinput.KeyComma, Modifier: ModNone},
	';': {Keycode: uinput.KeyComma, Modifier: ModShift},
	'.': {Keycode: uinput.KeyDot, Modifier: ModNone},
	':': {Keycode: uinput.KeyDot, Modifier: ModShift},
	'-': {Keycode: uinput.KeySlash, Modifier: ModNone},
	'_': {Keycode: uinput.KeySlash, Modifier: ModShift},
}

// esSymbols contains Spanish symbols.
var esSymbols = map[rune]KeyMapping{
	'\'': {Keycode: uinput.KeyMinus, Modifier: ModNone},
	'?':  {Keycode: uinput.KeyMinus, Modifier: ModShift},
	'¡':  {Keycode: uinput.KeyEqual, Modifier: ModNone},  // Inverted exclamation
	'¿':  {Keycode: uinput.KeyEqual, Modifier: ModShift}, // Inverted question mark
	'+':  {Keycode: uinput.KeyRightBrace, Modifier: ModNone},
	'*':  {Keycode: uinput.KeyRightBrace, Modifier: ModShift},
	'º':  {Keycode: uinput.KeyGrave, Modifier: ModNone},  // Masculine ordinal
	'ª':  {Keycode: uinput.KeyGrave, Modifier: ModShift}, // Feminine ordinal
}

// esAltGrSymbols contains Spanish AltGr combinations.
var esAltGrSymbols = map[rune]KeyMapping{
	'@':  {Keycode: uinput.Key2, Modifier: ModAltGr},
	'#':  {Keycode: uinput.Key3, Modifier: ModAltGr},
	'~':  {Keycode: uinput.Key4, Modifier: ModAltGr},
	'[':  {Keycode: uinput.KeyGrave, Modifier: ModAltGr},
	']':  {Keycode: uinput.KeyRightBrace, Modifier: ModAltGr},
	'{':  {Keycode: uinput.KeyApostrophe, Modifier: ModAltGr},
	'}':  {Keycode: uinput.KeyBackslash, Modifier: ModAltGr},
	'\\': {Keycode: uinput.KeyGrave, Modifier: ModAltGr | ModShift},
	'|':  {Keycode: uinput.Key1, Modifier: ModAltGr},
}

// esSpecialKeys contains Spanish special keys.
var esSpecialKeys = map[rune]KeyMapping{
	'<': {Keycode: uinput.Key102ND, Modifier: ModNone},
	'>': {Keycode: uinput.Key102ND, Modifier: ModShift},
}
