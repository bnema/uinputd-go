package layouts

import (
	"context"

	"github.com/bnema/uinputd-go/internal/uinput"
)

// UKLayout implements UK (British) QWERTY keyboard layout.
type UKLayout struct{}

// NewUK creates a new UK QWERTY layout.
func NewUK() *UKLayout {
	return &UKLayout{}
}

// Name returns "uk".
func (l *UKLayout) Name() string {
	return "uk"
}

// CharToKeycode maps a character to its keycode in UK QWERTY layout.
func (l *UKLayout) CharToKeycode(ctx context.Context, char rune) (uint16, bool, bool, error) {
	mapping, ok := ukKeymapData[char]
	if !ok {
		return 0, false, false, &ErrCharNotSupported{Char: char, Layout: "uk"}
	}

	shift := (mapping.Modifier & ModShift) != 0
	altGr := (mapping.Modifier & ModAltGr) != 0

	return mapping.Keycode, shift, altGr, nil
}

// ukKeymapData contains the complete UK QWERTY character-to-keycode mapping.
var ukKeymapData = map[rune]KeyMapping{
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

	// Shifted numbers (symbols - note differences from US)
	'!': {Keycode: uinput.Key1, Modifier: ModShift},
	'"': {Keycode: uinput.Key2, Modifier: ModShift}, // UK uses " not @
	'£': {Keycode: uinput.Key3, Modifier: ModShift}, // UK pound sign
	'$': {Keycode: uinput.Key4, Modifier: ModShift},
	'%': {Keycode: uinput.Key5, Modifier: ModShift},
	'^': {Keycode: uinput.Key6, Modifier: ModShift},
	'&': {Keycode: uinput.Key7, Modifier: ModShift},
	'*': {Keycode: uinput.Key8, Modifier: ModShift},
	'(': {Keycode: uinput.Key9, Modifier: ModShift},
	')': {Keycode: uinput.Key0, Modifier: ModShift},

	// QWERTY letter layout (standard)
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

	// Second row
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

	// Third row
	'z': {Keycode: uinput.KeyZ, Modifier: ModNone},
	'Z': {Keycode: uinput.KeyZ, Modifier: ModShift},
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

	// Special characters
	' ':  {Keycode: uinput.KeySpace, Modifier: ModNone},
	'\t': {Keycode: uinput.KeyTab, Modifier: ModNone},
	'\n': {Keycode: uinput.KeyEnter, Modifier: ModNone},

	// Punctuation
	',': {Keycode: uinput.KeyComma, Modifier: ModNone},
	'<': {Keycode: uinput.KeyComma, Modifier: ModShift},
	'.': {Keycode: uinput.KeyDot, Modifier: ModNone},
	'>': {Keycode: uinput.KeyDot, Modifier: ModShift},
	'/': {Keycode: uinput.KeySlash, Modifier: ModNone},
	'?': {Keycode: uinput.KeySlash, Modifier: ModShift},

	// UK specific symbols
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
	'@':  {Keycode: uinput.KeyApostrophe, Modifier: ModShift}, // UK @ is shifted '
	'#':  {Keycode: uinput.KeyBackslash, Modifier: ModNone},   // UK hash on backslash key
	'~':  {Keycode: uinput.KeyBackslash, Modifier: ModShift},
	'`':  {Keycode: uinput.KeyGrave, Modifier: ModNone},
	'¬':  {Keycode: uinput.KeyGrave, Modifier: ModShift}, // UK not sign
	'\\': {Keycode: uinput.Key102ND, Modifier: ModNone},  // UK backslash on 102nd key
	'|':  {Keycode: uinput.Key102ND, Modifier: ModShift},

	// AltGr combinations (Euro sign)
	'€': {Keycode: uinput.Key4, Modifier: ModAltGr},
}
