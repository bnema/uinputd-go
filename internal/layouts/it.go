package layouts

import (
	"context"

	"github.com/bnema/uinputd-go/internal/uinput"
)

// ITLayout implements Italian QWERTY keyboard layout.
type ITLayout struct {
	baseMappings    map[rune]KeyMapping
	deadKeyRegistry DeadKeyRegistry
	deadKeys        map[rune]KeyMapping
}

// NewIT creates a new Italian QWERTY layout.
func NewIT() *ITLayout {
	// Build the base mappings by merging shared and Italian-specific mappings
	base := MergeKeymaps(
		CommonMappings,         // Universal: space, tab, enter, €
		QWERTYBaseMappings,     // Standard QWERTY letter positions
		StandardNumberMappings, // Numbers 0-9 without shift
		itShiftedSymbols,       // Italian shifted symbols
		itPrecomposedAccents,   // Direct keys for è, é, ò, à, ù, ì, ç
		itPunctuation,          // Italian punctuation
		itSymbols,              // Italian symbols
		itAltGrSymbols,         // Italian AltGr combinations
		itSpecialKeys,          // Italian special keys
	)

	return &ITLayout{
		baseMappings:    base,
		deadKeyRegistry: BuildDeadKeyRegistry(),
		deadKeys:        itDeadKeys,
	}
}

// Name returns "it".
func (l *ITLayout) Name() string {
	return "it"
}

// CharToKeySequence converts a Unicode character to a sequence of keystrokes.
func (l *ITLayout) CharToKeySequence(ctx context.Context, char rune) ([]KeySequence, error) {
	// First, check if it's a direct mapping
	if mapping, ok := l.baseMappings[char]; ok {
		return []KeySequence{{Keycode: mapping.Keycode, Modifier: mapping.Modifier}}, nil
	}

	// Check if it needs a dead key combination
	if comp, ok := l.deadKeyRegistry[char]; ok {
		deadKeyMapping, hasDead := l.deadKeys[comp.DeadKey]
		if !hasDead {
			return nil, &ErrCharNotSupported{Char: char, Layout: "it"}
		}

		baseMapping, hasBase := l.baseMappings[comp.BaseChar]
		if !hasBase {
			return nil, &ErrCharNotSupported{Char: char, Layout: "it"}
		}

		return []KeySequence{
			{Keycode: deadKeyMapping.Keycode, Modifier: deadKeyMapping.Modifier},
			{Keycode: baseMapping.Keycode, Modifier: baseMapping.Modifier},
		}, nil
	}

	return nil, &ErrCharNotSupported{Char: char, Layout: "it"}
}

// itDeadKeys maps dead key symbols to their physical location on Italian keyboard.
var itDeadKeys = map[rune]KeyMapping{
	'^': {Keycode: uinput.KeyEqual, Modifier: ModShift}, // Circumflex
}

// itShiftedSymbols contains Italian shifted symbols on number row.
var itShiftedSymbols = map[rune]KeyMapping{
	'!': {Keycode: uinput.Key1, Modifier: ModShift},
	'"': {Keycode: uinput.Key2, Modifier: ModShift},
	'£': {Keycode: uinput.Key3, Modifier: ModShift},
	'$': {Keycode: uinput.Key4, Modifier: ModShift},
	'%': {Keycode: uinput.Key5, Modifier: ModShift},
	'&': {Keycode: uinput.Key6, Modifier: ModShift},
	'/': {Keycode: uinput.Key7, Modifier: ModShift},
	'(': {Keycode: uinput.Key8, Modifier: ModShift},
	')': {Keycode: uinput.Key9, Modifier: ModShift},
	'=': {Keycode: uinput.Key0, Modifier: ModShift},
}

// itPrecomposedAccents contains Italian characters with dedicated keys.
var itPrecomposedAccents = map[rune]KeyMapping{
	'è': {Keycode: uinput.KeyLeftBrace, Modifier: ModNone},
	'é': {Keycode: uinput.KeyLeftBrace, Modifier: ModShift},
	'ò': {Keycode: uinput.KeySemicolon, Modifier: ModNone},
	'ç': {Keycode: uinput.KeySemicolon, Modifier: ModShift},
	'à': {Keycode: uinput.KeyApostrophe, Modifier: ModNone},
	'°': {Keycode: uinput.KeyApostrophe, Modifier: ModShift}, // Degree symbol
	'ù': {Keycode: uinput.KeyBackslash, Modifier: ModNone},
	'§': {Keycode: uinput.KeyBackslash, Modifier: ModShift}, // Section symbol
	'ì': {Keycode: uinput.KeyEqual, Modifier: ModNone},
}

// itPunctuation contains Italian punctuation.
var itPunctuation = map[rune]KeyMapping{
	',': {Keycode: uinput.KeyComma, Modifier: ModNone},
	';': {Keycode: uinput.KeyComma, Modifier: ModShift},
	'.': {Keycode: uinput.KeyDot, Modifier: ModNone},
	':': {Keycode: uinput.KeyDot, Modifier: ModShift},
	'-': {Keycode: uinput.KeySlash, Modifier: ModNone},
	'_': {Keycode: uinput.KeySlash, Modifier: ModShift},
}

// itSymbols contains Italian symbols.
var itSymbols = map[rune]KeyMapping{
	'\'': {Keycode: uinput.KeyMinus, Modifier: ModNone},
	'?':  {Keycode: uinput.KeyMinus, Modifier: ModShift},
	'+':  {Keycode: uinput.KeyRightBrace, Modifier: ModNone},
	'*':  {Keycode: uinput.KeyRightBrace, Modifier: ModShift},
	'\\': {Keycode: uinput.KeyGrave, Modifier: ModNone},
	'|':  {Keycode: uinput.KeyGrave, Modifier: ModShift},
}

// itAltGrSymbols contains Italian AltGr combinations.
var itAltGrSymbols = map[rune]KeyMapping{
	'@': {Keycode: uinput.KeyApostrophe, Modifier: ModAltGr},
	'#': {Keycode: uinput.KeyBackslash, Modifier: ModAltGr},
	'[': {Keycode: uinput.KeyLeftBrace, Modifier: ModAltGr},
	']': {Keycode: uinput.KeyRightBrace, Modifier: ModAltGr},
	'{': {Keycode: uinput.KeyLeftBrace, Modifier: ModAltGr | ModShift},
	'}': {Keycode: uinput.KeyRightBrace, Modifier: ModAltGr | ModShift},
	'~': {Keycode: uinput.KeyEqual, Modifier: ModAltGr},
	'`': {Keycode: uinput.KeyMinus, Modifier: ModAltGr},
}

// itSpecialKeys contains Italian special keys.
var itSpecialKeys = map[rune]KeyMapping{
	'<': {Keycode: uinput.Key102ND, Modifier: ModNone},
	'>': {Keycode: uinput.Key102ND, Modifier: ModShift},
}
