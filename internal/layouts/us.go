package layouts

import (
	"context"

	"github.com/bnema/uinputd-go/internal/uinput"
)

// USLayout implements US QWERTY keyboard layout.
type USLayout struct{}

// NewUS creates a new US QWERTY layout.
func NewUS() *USLayout {
	return &USLayout{}
}

// Name returns "us".
func (l *USLayout) Name() string {
	return "us"
}

// CharToKeycode maps a character to its keycode in US QWERTY layout.
func (l *USLayout) CharToKeycode(ctx context.Context, char rune) (uint16, bool, bool, error) {
	mapping, ok := usKeymapData[char]
	if !ok {
		return 0, false, false, &ErrCharNotSupported{Char: char, Layout: "us"}
	}

	shift := (mapping.Modifier & ModShift) != 0
	altGr := (mapping.Modifier & ModAltGr) != 0

	return mapping.Keycode, shift, altGr, nil
}

// usKeymapData contains the complete US QWERTY character-to-keycode mapping.
var usKeymapData = map[rune]KeyMapping{
	// Numbers (no shift)
	'1': {Keycode: uinput.Key1, Modifier: ModNone},
	'2': {Keycode: uinput.Key2, Modifier: ModNone},
	'3': {Keycode: uinput.Key3, Modifier: ModNone},
	'4': {Keycode: uinput.Key4, Modifier: ModNone},
	'5': {Keycode: uinput.Key5, Modifier: ModNone},
	'6': {Keycode: uinput.Key6, Modifier: ModNone},
	'7': {Keycode: uinput.Key7, Modifier: ModNone},
	'8': {Keycode: uinput.Key8, Modifier: ModNone},
	'9': {Keycode: uinput.Key9, Modifier: ModNone},
	'0': {Keycode: uinput.Key0, Modifier: ModNone},

	// Shifted numbers (symbols)
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

	// Lowercase letters
	'a': {Keycode: uinput.KeyA, Modifier: ModNone},
	'b': {Keycode: uinput.KeyB, Modifier: ModNone},
	'c': {Keycode: uinput.KeyC, Modifier: ModNone},
	'd': {Keycode: uinput.KeyD, Modifier: ModNone},
	'e': {Keycode: uinput.KeyE, Modifier: ModNone},
	'f': {Keycode: uinput.KeyF, Modifier: ModNone},
	'g': {Keycode: uinput.KeyG, Modifier: ModNone},
	'h': {Keycode: uinput.KeyH, Modifier: ModNone},
	'i': {Keycode: uinput.KeyI, Modifier: ModNone},
	'j': {Keycode: uinput.KeyJ, Modifier: ModNone},
	'k': {Keycode: uinput.KeyK, Modifier: ModNone},
	'l': {Keycode: uinput.KeyL, Modifier: ModNone},
	'm': {Keycode: uinput.KeyM, Modifier: ModNone},
	'n': {Keycode: uinput.KeyN, Modifier: ModNone},
	'o': {Keycode: uinput.KeyO, Modifier: ModNone},
	'p': {Keycode: uinput.KeyP, Modifier: ModNone},
	'q': {Keycode: uinput.KeyQ, Modifier: ModNone},
	'r': {Keycode: uinput.KeyR, Modifier: ModNone},
	's': {Keycode: uinput.KeyS, Modifier: ModNone},
	't': {Keycode: uinput.KeyT, Modifier: ModNone},
	'u': {Keycode: uinput.KeyU, Modifier: ModNone},
	'v': {Keycode: uinput.KeyV, Modifier: ModNone},
	'w': {Keycode: uinput.KeyW, Modifier: ModNone},
	'x': {Keycode: uinput.KeyX, Modifier: ModNone},
	'y': {Keycode: uinput.KeyY, Modifier: ModNone},
	'z': {Keycode: uinput.KeyZ, Modifier: ModNone},

	// Uppercase letters
	'A': {Keycode: uinput.KeyA, Modifier: ModShift},
	'B': {Keycode: uinput.KeyB, Modifier: ModShift},
	'C': {Keycode: uinput.KeyC, Modifier: ModShift},
	'D': {Keycode: uinput.KeyD, Modifier: ModShift},
	'E': {Keycode: uinput.KeyE, Modifier: ModShift},
	'F': {Keycode: uinput.KeyF, Modifier: ModShift},
	'G': {Keycode: uinput.KeyG, Modifier: ModShift},
	'H': {Keycode: uinput.KeyH, Modifier: ModShift},
	'I': {Keycode: uinput.KeyI, Modifier: ModShift},
	'J': {Keycode: uinput.KeyJ, Modifier: ModShift},
	'K': {Keycode: uinput.KeyK, Modifier: ModShift},
	'L': {Keycode: uinput.KeyL, Modifier: ModShift},
	'M': {Keycode: uinput.KeyM, Modifier: ModShift},
	'N': {Keycode: uinput.KeyN, Modifier: ModShift},
	'O': {Keycode: uinput.KeyO, Modifier: ModShift},
	'P': {Keycode: uinput.KeyP, Modifier: ModShift},
	'Q': {Keycode: uinput.KeyQ, Modifier: ModShift},
	'R': {Keycode: uinput.KeyR, Modifier: ModShift},
	'S': {Keycode: uinput.KeyS, Modifier: ModShift},
	'T': {Keycode: uinput.KeyT, Modifier: ModShift},
	'U': {Keycode: uinput.KeyU, Modifier: ModShift},
	'V': {Keycode: uinput.KeyV, Modifier: ModShift},
	'W': {Keycode: uinput.KeyW, Modifier: ModShift},
	'X': {Keycode: uinput.KeyX, Modifier: ModShift},
	'Y': {Keycode: uinput.KeyY, Modifier: ModShift},
	'Z': {Keycode: uinput.KeyZ, Modifier: ModShift},

	// Special characters
	' ':  {Keycode: uinput.KeySpace, Modifier: ModNone},
	'\t': {Keycode: uinput.KeyTab, Modifier: ModNone},
	'\n': {Keycode: uinput.KeyEnter, Modifier: ModNone},
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
	',':  {Keycode: uinput.KeyComma, Modifier: ModNone},
	'<':  {Keycode: uinput.KeyComma, Modifier: ModShift},
	'.':  {Keycode: uinput.KeyDot, Modifier: ModNone},
	'>':  {Keycode: uinput.KeyDot, Modifier: ModShift},
	'/':  {Keycode: uinput.KeySlash, Modifier: ModNone},
	'?':  {Keycode: uinput.KeySlash, Modifier: ModShift},
}
