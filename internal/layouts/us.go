package layouts

import (
	"context"

	"github.com/bnema/uinputd-go/internal/uinput"
)

// USLayout implements US QWERTY keyboard layout.
type USLayout struct {
	baseMappings    map[rune]KeyMapping
	deadKeyRegistry DeadKeyRegistry
	deadKeys        map[rune]KeyMapping
}

// NewUS creates a new US QWERTY layout.
func NewUS() *USLayout {
	// Build the base mappings by merging shared mappings
	base := MergeKeymaps(
		CommonMappings,         // Universal: space, tab, enter, €
		QWERTYBaseMappings,     // Standard QWERTY letter positions
		StandardNumberMappings, // Numbers 0-9 without shift
		usShiftedSymbols,       // US-specific shifted symbols
		usPunctuation,          // US punctuation layout
		usSymbols,              // US symbols and brackets
	)

	return &USLayout{
		baseMappings:    base,
		deadKeyRegistry: BuildDeadKeyRegistry(),
		deadKeys:        make(map[rune]KeyMapping), // US layout has no dead keys
	}
}

// Name returns "us".
func (l *USLayout) Name() string {
	return "us"
}

// CharToKeySequence converts a Unicode character to a sequence of keystrokes.
func (l *USLayout) CharToKeySequence(ctx context.Context, char rune) ([]KeySequence, error) {
	// First, check if it's a direct mapping
	if mapping, ok := l.baseMappings[char]; ok {
		return []KeySequence{{Keycode: mapping.Keycode, Modifier: mapping.Modifier}}, nil
	}

	// Check if it needs a dead key combination
	// Note: US layout doesn't have dead keys, but we keep this code for consistency
	if comp, ok := l.deadKeyRegistry[char]; ok {
		deadKeyMapping, hasDead := l.deadKeys[comp.DeadKey]
		if !hasDead {
			return nil, &ErrCharNotSupported{Char: char, Layout: "us"}
		}

		baseMapping, hasBase := l.baseMappings[comp.BaseChar]
		if !hasBase {
			return nil, &ErrCharNotSupported{Char: char, Layout: "us"}
		}

		return []KeySequence{
			{Keycode: deadKeyMapping.Keycode, Modifier: deadKeyMapping.Modifier},
			{Keycode: baseMapping.Keycode, Modifier: baseMapping.Modifier},
		}, nil
	}

	return nil, &ErrCharNotSupported{Char: char, Layout: "us"}
}

// usShiftedSymbols contains US-specific shifted symbols on the number row.
var usShiftedSymbols = map[rune]KeyMapping{
	'!': {Keycode: uinput.Key1, Modifier: ModShift},
	'@': {Keycode: uinput.Key2, Modifier: ModShift},
	'#': {Keycode: uinput.Key3, Modifier: ModShift},
	'$': {Keycode: uinput.Key4, Modifier: ModShift},
	'%': {Keycode: uinput.Key5, Modifier: ModShift},
	'^': {Keycode: uinput.Key6, Modifier: ModShift},
	'&': {Keycode: uinput.Key7, Modifier: ModShift},
	'*': {Keycode: uinput.Key8, Modifier: ModShift},
	'(': {Keycode: uinput.Key9, Modifier: ModShift},
	')': {Keycode: uinput.Key0, Modifier: ModShift},
}

// usPunctuation contains US punctuation layout.
var usPunctuation = map[rune]KeyMapping{
	',': {Keycode: uinput.KeyComma, Modifier: ModNone},
	'<': {Keycode: uinput.KeyComma, Modifier: ModShift},
	'.': {Keycode: uinput.KeyDot, Modifier: ModNone},
	'>': {Keycode: uinput.KeyDot, Modifier: ModShift},
	'/': {Keycode: uinput.KeySlash, Modifier: ModNone},
	'?': {Keycode: uinput.KeySlash, Modifier: ModShift},
}

// usSymbols contains US symbols and brackets.
var usSymbols = map[rune]KeyMapping{
	'-':  {Keycode: uinput.KeyMinus, Modifier: ModNone},
	'_':  {Keycode: uinput.KeyMinus, Modifier: ModShift},
	'=':  {Keycode: uinput.KeyEqual, Modifier: ModNone},
	'+':  {Keycode: uinput.KeyEqual, Modifier: ModShift},
	'[':  {Keycode: uinput.KeyLeftBrace, Modifier: ModNone},
	'{':  {Keycode: uinput.KeyLeftBrace, Modifier: ModShift},
	']':  {Keycode: uinput.KeyRightBrace, Modifier: ModNone},
	'}':  {Keycode: uinput.KeyRightBrace, Modifier: ModShift},
	'\\': {Keycode: uinput.KeyBackslash, Modifier: ModNone},
	'|':  {Keycode: uinput.KeyBackslash, Modifier: ModShift},
	';':  {Keycode: uinput.KeySemicolon, Modifier: ModNone},
	':':  {Keycode: uinput.KeySemicolon, Modifier: ModShift},
	'\'': {Keycode: uinput.KeyApostrophe, Modifier: ModNone},
	'"':  {Keycode: uinput.KeyApostrophe, Modifier: ModShift},
	'`':  {Keycode: uinput.KeyGrave, Modifier: ModNone},
	'~':  {Keycode: uinput.KeyGrave, Modifier: ModShift},
}
