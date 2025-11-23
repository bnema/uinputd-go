package layouts

import (
	"context"

	"github.com/bnema/uinputd-go/internal/uinput"
)

// FRLayout implements French AZERTY keyboard layout.
type FRLayout struct{}

// NewFR creates a new French AZERTY layout.
func NewFR() *FRLayout {
	return &FRLayout{}
}

// Name returns "fr".
func (l *FRLayout) Name() string {
	return "fr"
}

// CharToKeycode maps a character to its keycode in French AZERTY layout.
func (l *FRLayout) CharToKeycode(ctx context.Context, char rune) (uint16, bool, bool, error) {
	mapping, ok := frKeymapData[char]
	if !ok {
		return 0, false, false, &ErrCharNotSupported{Char: char, Layout: "fr"}
	}

	shift := (mapping.Modifier & ModShift) != 0
	altGr := (mapping.Modifier & ModAltGr) != 0

	return mapping.Keycode, shift, altGr, nil
}

// frKeymapData contains the complete French AZERTY character-to-keycode mapping.
var frKeymapData = map[rune]KeyMapping{
	// Numbers (shifted in AZERTY!)
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

	// Unshifted symbols (AZERTY specific)
	'&':  {Keycode: uinput.Key1, Modifier: ModNone},
	'é':  {Keycode: uinput.Key2, Modifier: ModNone},
	'"':  {Keycode: uinput.Key3, Modifier: ModNone},
	'\'': {Keycode: uinput.Key4, Modifier: ModNone},
	'(':  {Keycode: uinput.Key5, Modifier: ModNone},
	'-':  {Keycode: uinput.Key6, Modifier: ModNone},
	'è':  {Keycode: uinput.Key7, Modifier: ModNone},
	'_':  {Keycode: uinput.Key8, Modifier: ModNone},
	'ç':  {Keycode: uinput.Key9, Modifier: ModNone},
	'à':  {Keycode: uinput.Key0, Modifier: ModNone},

	// AZERTY letter layout (first row)
	'a': {Keycode: uinput.KeyQ, Modifier: ModNone},
	'A': {Keycode: uinput.KeyQ, Modifier: ModShift},
	'z': {Keycode: uinput.KeyW, Modifier: ModNone},
	'Z': {Keycode: uinput.KeyW, Modifier: ModShift},
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

	// AZERTY letter layout (second row)
	'q': {Keycode: uinput.KeyA, Modifier: ModNone},
	'Q': {Keycode: uinput.KeyA, Modifier: ModShift},
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
	'm': {Keycode: uinput.KeySemicolon, Modifier: ModNone},
	'M': {Keycode: uinput.KeySemicolon, Modifier: ModShift},

	// AZERTY letter layout (third row)
	'w': {Keycode: uinput.KeyZ, Modifier: ModNone},
	'W': {Keycode: uinput.KeyZ, Modifier: ModShift},
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

	// Special characters
	' ':  {Keycode: uinput.KeySpace, Modifier: ModNone},
	'\t': {Keycode: uinput.KeyTab, Modifier: ModNone},
	'\n': {Keycode: uinput.KeyEnter, Modifier: ModNone},

	// Punctuation and symbols
	',': {Keycode: uinput.KeyM, Modifier: ModNone},
	'?': {Keycode: uinput.KeyM, Modifier: ModShift},
	';': {Keycode: uinput.KeyComma, Modifier: ModNone},
	'.': {Keycode: uinput.KeyComma, Modifier: ModShift},
	':': {Keycode: uinput.KeyDot, Modifier: ModNone},
	'/': {Keycode: uinput.KeyDot, Modifier: ModShift},
	'!': {Keycode: uinput.KeySlash, Modifier: ModNone},

	// Brackets and special (with AltGr)
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

	// Remaining symbols
	'°': {Keycode: uinput.KeyMinus, Modifier: ModShift},
	')': {Keycode: uinput.KeyEqual, Modifier: ModNone},
	'=': {Keycode: uinput.KeyEqual, Modifier: ModShift},
	'^': {Keycode: uinput.KeyLeftBrace, Modifier: ModNone},
	'¨': {Keycode: uinput.KeyLeftBrace, Modifier: ModShift},
	'$': {Keycode: uinput.KeyRightBrace, Modifier: ModNone},
	'£': {Keycode: uinput.KeyRightBrace, Modifier: ModShift},
	'*': {Keycode: uinput.KeyBackslash, Modifier: ModNone},
	'µ': {Keycode: uinput.KeyBackslash, Modifier: ModShift},
	'ù': {Keycode: uinput.KeyApostrophe, Modifier: ModNone},
	'%': {Keycode: uinput.KeyApostrophe, Modifier: ModShift},
	'<': {Keycode: uinput.KeyGrave, Modifier: ModNone},
	'>': {Keycode: uinput.KeyGrave, Modifier: ModShift},
}
