package layouts

import (
	"context"

	"github.com/bnema/uinputd-go/internal/uinput"
)

// UKLayout implements UK (British) QWERTY keyboard layout.
type UKLayout struct {
	baseMappings    map[rune]KeyMapping
	deadKeyRegistry DeadKeyRegistry
	deadKeys        map[rune]KeyMapping
}

// NewUK creates a new UK QWERTY layout.
func NewUK() *UKLayout {
	// Build the base mappings by merging shared and UK-specific mappings
	base := MergeKeymaps(
		CommonMappings,         // Universal: space, tab, enter (but NOT € - UK has it elsewhere)
		QWERTYBaseMappings,     // Standard QWERTY letter positions
		StandardNumberMappings, // Numbers 0-9 without shift
		ukShiftedSymbols,       // UK-specific shifted symbols (different from US)
		ukPunctuation,          // UK punctuation layout
		ukSymbols,              // UK symbols and brackets
		ukAltGrSymbols,         // UK AltGr combinations
		ukSpecialKeys,          // UK-specific keys like 102ND
	)

	return &UKLayout{
		baseMappings:    base,
		deadKeyRegistry: BuildDeadKeyRegistry(),
		deadKeys:        make(map[rune]KeyMapping), // UK layout has no dead keys
	}
}

// Name returns "uk".
func (l *UKLayout) Name() string {
	return "uk"
}

// CharToKeySequence converts a Unicode character to a sequence of keystrokes.
func (l *UKLayout) CharToKeySequence(ctx context.Context, char rune) ([]KeySequence, error) {
	// First, check if it's a direct mapping
	if mapping, ok := l.baseMappings[char]; ok {
		return []KeySequence{{Keycode: mapping.Keycode, Modifier: mapping.Modifier}}, nil
	}

	// Check if it needs a dead key combination
	if comp, ok := l.deadKeyRegistry[char]; ok {
		deadKeyMapping, hasDead := l.deadKeys[comp.DeadKey]
		if !hasDead {
			return nil, &ErrCharNotSupported{Char: char, Layout: "uk"}
		}

		baseMapping, hasBase := l.baseMappings[comp.BaseChar]
		if !hasBase {
			return nil, &ErrCharNotSupported{Char: char, Layout: "uk"}
		}

		return []KeySequence{
			{Keycode: deadKeyMapping.Keycode, Modifier: deadKeyMapping.Modifier},
			{Keycode: baseMapping.Keycode, Modifier: baseMapping.Modifier},
		}, nil
	}

	return nil, &ErrCharNotSupported{Char: char, Layout: "uk"}
}

// ukShiftedSymbols contains UK-specific shifted symbols on the number row.
// Note: UK uses " instead of @ on Shift+2, and £ instead of # on Shift+3.
var ukShiftedSymbols = map[rune]KeyMapping{
	'!': {Keycode: uinput.Key1, Modifier: ModShift},
	'"': {Keycode: uinput.Key2, Modifier: ModShift}, // UK-specific
	'£': {Keycode: uinput.Key3, Modifier: ModShift}, // UK pound sign
	'$': {Keycode: uinput.Key4, Modifier: ModShift},
	'%': {Keycode: uinput.Key5, Modifier: ModShift},
	'^': {Keycode: uinput.Key6, Modifier: ModShift},
	'&': {Keycode: uinput.Key7, Modifier: ModShift},
	'*': {Keycode: uinput.Key8, Modifier: ModShift},
	'(': {Keycode: uinput.Key9, Modifier: ModShift},
	')': {Keycode: uinput.Key0, Modifier: ModShift},
}

// ukPunctuation contains UK punctuation layout.
var ukPunctuation = map[rune]KeyMapping{
	',': {Keycode: uinput.KeyComma, Modifier: ModNone},
	'<': {Keycode: uinput.KeyComma, Modifier: ModShift},
	'.': {Keycode: uinput.KeyDot, Modifier: ModNone},
	'>': {Keycode: uinput.KeyDot, Modifier: ModShift},
	'/': {Keycode: uinput.KeySlash, Modifier: ModNone},
	'?': {Keycode: uinput.KeySlash, Modifier: ModShift},
}

// ukSymbols contains UK symbols and brackets.
var ukSymbols = map[rune]KeyMapping{
	'-':  {Keycode: uinput.KeyMinus, Modifier: ModNone},
	'_':  {Keycode: uinput.KeyMinus, Modifier: ModShift},
	'=':  {Keycode: uinput.KeyEqual, Modifier: ModNone},
	'+':  {Keycode: uinput.KeyEqual, Modifier: ModShift},
	'[':  {Keycode: uinput.KeyLeftBrace, Modifier: ModNone},
	'{':  {Keycode: uinput.KeyLeftBrace, Modifier: ModShift},
	']':  {Keycode: uinput.KeyRightBrace, Modifier: ModNone},
	'}':  {Keycode: uinput.KeyRightBrace, Modifier: ModShift},
	';':  {Keycode: uinput.KeySemicolon, Modifier: ModNone},
	':':  {Keycode: uinput.KeySemicolon, Modifier: ModShift},
	'\'': {Keycode: uinput.KeyApostrophe, Modifier: ModNone},
	'@':  {Keycode: uinput.KeyApostrophe, Modifier: ModShift}, // UK-specific: @ is Shift+'
	'#':  {Keycode: uinput.KeyBackslash, Modifier: ModNone},   // UK-specific: # on backslash key
	'~':  {Keycode: uinput.KeyBackslash, Modifier: ModShift},
	'`':  {Keycode: uinput.KeyGrave, Modifier: ModNone},
	'¬':  {Keycode: uinput.KeyGrave, Modifier: ModShift}, // UK not sign
}

// ukAltGrSymbols contains UK AltGr combinations.
// Note: UK has € at AltGr+4, not AltGr+E like most other layouts.
var ukAltGrSymbols = map[rune]KeyMapping{
	'€': {Keycode: uinput.Key4, Modifier: ModAltGr}, // UK-specific position
}

// ukSpecialKeys contains UK-specific keys like the 102ND key.
var ukSpecialKeys = map[rune]KeyMapping{
	'\\': {Keycode: uinput.Key102ND, Modifier: ModNone},
	'|':  {Keycode: uinput.Key102ND, Modifier: ModShift},
}
