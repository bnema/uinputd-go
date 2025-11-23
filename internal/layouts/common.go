package layouts

import "github.com/bnema/uinputd-go/internal/uinput"

// CommonMappings contains truly universal character mappings that work
// identically across ALL keyboard layouts.
var CommonMappings = map[rune]KeyMapping{
	' ':  {Keycode: uinput.KeySpace, Modifier: ModNone},
	'\t': {Keycode: uinput.KeyTab, Modifier: ModNone},
	'\n': {Keycode: uinput.KeyEnter, Modifier: ModNone},
	'€':  {Keycode: uinput.KeyE, Modifier: ModAltGr},
}

// QWERTYBaseMappings contains the standard QWERTY letter positions.
// This is used by US, UK, and other QWERTY-based layouts where letters
// map directly to their corresponding key positions (a→KeyA, b→KeyB, etc.).
//
// Note: AZERTY (French) and QWERTZ (German) layouts should NOT use this
// as their letter positions are different.
var QWERTYBaseMappings = map[rune]KeyMapping{
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
}

// StandardNumberMappings contains the number row for standard QWERTY layouts.
// On QWERTY layouts (US, UK, etc.), numbers 0-9 are typed without shift.
//
// Note: French AZERTY requires shift for numbers, so this should NOT be used
// for French layouts.
var StandardNumberMappings = map[rune]KeyMapping{
	'0': {Keycode: uinput.Key0, Modifier: ModNone},
	'1': {Keycode: uinput.Key1, Modifier: ModNone},
	'2': {Keycode: uinput.Key2, Modifier: ModNone},
	'3': {Keycode: uinput.Key3, Modifier: ModNone},
	'4': {Keycode: uinput.Key4, Modifier: ModNone},
	'5': {Keycode: uinput.Key5, Modifier: ModNone},
	'6': {Keycode: uinput.Key6, Modifier: ModNone},
	'7': {Keycode: uinput.Key7, Modifier: ModNone},
	'8': {Keycode: uinput.Key8, Modifier: ModNone},
	'9': {Keycode: uinput.Key9, Modifier: ModNone},
}

// MergeKeymaps merges multiple keymaps into a single map.
// Later maps override earlier ones in case of conflicts.
// This allows layouts to compose their mappings from shared bases
// plus layout-specific overrides.
func MergeKeymaps(maps ...map[rune]KeyMapping) map[rune]KeyMapping {
	result := make(map[rune]KeyMapping)
	for _, m := range maps {
		for k, v := range m {
			result[k] = v
		}
	}
	return result
}
