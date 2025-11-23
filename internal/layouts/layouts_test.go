package layouts

import (
	"context"
	"testing"

	"github.com/bnema/uinputd-go/internal/uinput"
)

func TestFrenchDeadKeyCompositions(t *testing.T) {
	layout := NewFR()
	ctx := context.Background()

	tests := []struct {
		char     rune
		expected []KeySequence
		desc     string
	}{
		// Circumflex combinations (^ is at KeyLeftBrace, ModNone)
		{
			char: 'â',
			expected: []KeySequence{
				{Keycode: uinput.KeyLeftBrace, Modifier: ModNone}, // ^
				{Keycode: uinput.KeyQ, Modifier: ModNone},         // a (on AZERTY KeyQ)
			},
			desc: "â = circumflex + a",
		},
		{
			char: 'ê',
			expected: []KeySequence{
				{Keycode: uinput.KeyLeftBrace, Modifier: ModNone}, // ^
				{Keycode: uinput.KeyE, Modifier: ModNone},         // e
			},
			desc: "ê = circumflex + e",
		},
		{
			char: 'î',
			expected: []KeySequence{
				{Keycode: uinput.KeyLeftBrace, Modifier: ModNone}, // ^
				{Keycode: uinput.KeyI, Modifier: ModNone},         // i
			},
			desc: "î = circumflex + i",
		},
		{
			char: 'ô',
			expected: []KeySequence{
				{Keycode: uinput.KeyLeftBrace, Modifier: ModNone}, // ^
				{Keycode: uinput.KeyO, Modifier: ModNone},         // o
			},
			desc: "ô = circumflex + o",
		},
		{
			char: 'û',
			expected: []KeySequence{
				{Keycode: uinput.KeyLeftBrace, Modifier: ModNone}, // ^
				{Keycode: uinput.KeyU, Modifier: ModNone},         // u
			},
			desc: "û = circumflex + u",
		},
		// Uppercase variants
		{
			char: 'Â',
			expected: []KeySequence{
				{Keycode: uinput.KeyLeftBrace, Modifier: ModNone}, // ^
				{Keycode: uinput.KeyQ, Modifier: ModShift},        // A (Shift+a on AZERTY KeyQ)
			},
			desc: "Â = circumflex + A",
		},
		// Diaeresis combinations (¨ is at KeyLeftBrace, ModShift)
		{
			char: 'ë',
			expected: []KeySequence{
				{Keycode: uinput.KeyLeftBrace, Modifier: ModShift}, // ¨
				{Keycode: uinput.KeyE, Modifier: ModNone},          // e
			},
			desc: "ë = diaeresis + e",
		},
		{
			char: 'ï',
			expected: []KeySequence{
				{Keycode: uinput.KeyLeftBrace, Modifier: ModShift}, // ¨
				{Keycode: uinput.KeyI, Modifier: ModNone},          // i
			},
			desc: "ï = diaeresis + i",
		},
		{
			char: 'ü',
			expected: []KeySequence{
				{Keycode: uinput.KeyLeftBrace, Modifier: ModShift}, // ¨
				{Keycode: uinput.KeyU, Modifier: ModNone},          // u
			},
			desc: "ü = diaeresis + u",
		},
		// Pre-composed characters (should work as direct keys)
		{
			char: 'é',
			expected: []KeySequence{
				{Keycode: uinput.Key2, Modifier: ModNone},
			},
			desc: "é has direct key",
		},
		{
			char: 'è',
			expected: []KeySequence{
				{Keycode: uinput.Key7, Modifier: ModNone},
			},
			desc: "è has direct key",
		},
		{
			char: 'à',
			expected: []KeySequence{
				{Keycode: uinput.Key0, Modifier: ModNone},
			},
			desc: "à has direct key",
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			seq, err := layout.CharToKeySequence(ctx, tt.char)
			if err != nil {
				t.Errorf("char %q not supported: %v", tt.char, err)
				return
			}

			if len(seq) != len(tt.expected) {
				t.Errorf("char %q: got %d keystrokes, want %d", tt.char, len(seq), len(tt.expected))
				return
			}

			for i, key := range seq {
				if key.Keycode != tt.expected[i].Keycode || key.Modifier != tt.expected[i].Modifier {
					t.Errorf("char %q keystroke[%d]: got {Keycode: %d, Modifier: %d}, want {Keycode: %d, Modifier: %d}",
						tt.char, i, key.Keycode, key.Modifier, tt.expected[i].Keycode, tt.expected[i].Modifier)
				}
			}
		})
	}
}

func TestGermanDeadKeyCompositions(t *testing.T) {
	layout := NewDE()
	ctx := context.Background()

	tests := []struct {
		char     rune
		expected []KeySequence
		desc     string
	}{
		// Circumflex combinations (^ is at KeyGrave, ModNone)
		{
			char: 'â',
			expected: []KeySequence{
				{Keycode: uinput.KeyGrave, Modifier: ModNone}, // ^
				{Keycode: uinput.KeyA, Modifier: ModNone},     // a
			},
			desc: "â = circumflex + a",
		},
		{
			char: 'ô',
			expected: []KeySequence{
				{Keycode: uinput.KeyGrave, Modifier: ModNone}, // ^
				{Keycode: uinput.KeyO, Modifier: ModNone},     // o
			},
			desc: "ô = circumflex + o",
		},
		// Acute combinations (´ is at KeyEqual, ModNone)
		{
			char: 'á',
			expected: []KeySequence{
				{Keycode: uinput.KeyEqual, Modifier: ModNone}, // ´
				{Keycode: uinput.KeyA, Modifier: ModNone},     // a
			},
			desc: "á = acute + a",
		},
		{
			char: 'é',
			expected: []KeySequence{
				{Keycode: uinput.KeyEqual, Modifier: ModNone}, // ´
				{Keycode: uinput.KeyE, Modifier: ModNone},     // e
			},
			desc: "é = acute + e",
		},
		// Umlauts (should work as direct keys)
		{
			char: 'ä',
			expected: []KeySequence{
				{Keycode: uinput.KeyApostrophe, Modifier: ModNone},
			},
			desc: "ä has direct key",
		},
		{
			char: 'ö',
			expected: []KeySequence{
				{Keycode: uinput.KeySemicolon, Modifier: ModNone},
			},
			desc: "ö has direct key",
		},
		{
			char: 'ü',
			expected: []KeySequence{
				{Keycode: uinput.KeyLeftBrace, Modifier: ModNone},
			},
			desc: "ü has direct key",
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			seq, err := layout.CharToKeySequence(ctx, tt.char)
			if err != nil {
				t.Errorf("char %q not supported: %v", tt.char, err)
				return
			}

			if len(seq) != len(tt.expected) {
				t.Errorf("char %q: got %d keystrokes, want %d", tt.char, len(seq), len(tt.expected))
				return
			}

			for i, key := range seq {
				if key.Keycode != tt.expected[i].Keycode || key.Modifier != tt.expected[i].Modifier {
					t.Errorf("char %q keystroke[%d]: got {Keycode: %d, Modifier: %d}, want {Keycode: %d, Modifier: %d}",
						tt.char, i, key.Keycode, key.Modifier, tt.expected[i].Keycode, tt.expected[i].Modifier)
				}
			}
		})
	}
}

func TestSpanishDeadKeyCompositions(t *testing.T) {
	layout := NewES()
	ctx := context.Background()

	tests := []struct {
		char     rune
		expected []KeySequence
		desc     string
	}{
		// Grave combinations (` is at KeyLeftBrace, ModNone)
		{
			char: 'à',
			expected: []KeySequence{
				{Keycode: uinput.KeyLeftBrace, Modifier: ModNone}, // `
				{Keycode: uinput.KeyA, Modifier: ModNone},         // a
			},
			desc: "à = grave + a",
		},
		{
			char: 'è',
			expected: []KeySequence{
				{Keycode: uinput.KeyLeftBrace, Modifier: ModNone}, // `
				{Keycode: uinput.KeyE, Modifier: ModNone},         // e
			},
			desc: "è = grave + e",
		},
		// Circumflex combinations (^ is at KeyLeftBrace, ModShift)
		{
			char: 'â',
			expected: []KeySequence{
				{Keycode: uinput.KeyLeftBrace, Modifier: ModShift}, // ^
				{Keycode: uinput.KeyA, Modifier: ModNone},          // a
			},
			desc: "â = circumflex + a",
		},
		// Pre-composed characters
		{
			char: 'ñ',
			expected: []KeySequence{
				{Keycode: uinput.KeySemicolon, Modifier: ModNone},
			},
			desc: "ñ has direct key",
		},
		{
			char: 'á',
			expected: []KeySequence{
				{Keycode: uinput.KeyApostrophe, Modifier: ModNone},
			},
			desc: "á has direct key",
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			seq, err := layout.CharToKeySequence(ctx, tt.char)
			if err != nil {
				t.Errorf("char %q not supported: %v", tt.char, err)
				return
			}

			if len(seq) != len(tt.expected) {
				t.Errorf("char %q: got %d keystrokes, want %d", tt.char, len(seq), len(tt.expected))
				return
			}

			for i, key := range seq {
				if key.Keycode != tt.expected[i].Keycode || key.Modifier != tt.expected[i].Modifier {
					t.Errorf("char %q keystroke[%d]: got {Keycode: %d, Modifier: %d}, want {Keycode: %d, Modifier: %d}",
						tt.char, i, key.Keycode, key.Modifier, tt.expected[i].Keycode, tt.expected[i].Modifier)
				}
			}
		})
	}
}

func TestItalianDeadKeyCompositions(t *testing.T) {
	layout := NewIT()
	ctx := context.Background()

	tests := []struct {
		char     rune
		expected []KeySequence
		desc     string
	}{
		// Circumflex combinations (^ is at KeyEqual, ModShift)
		{
			char: 'â',
			expected: []KeySequence{
				{Keycode: uinput.KeyEqual, Modifier: ModShift}, // ^
				{Keycode: uinput.KeyA, Modifier: ModNone},      // a
			},
			desc: "â = circumflex + a",
		},
		{
			char: 'ê',
			expected: []KeySequence{
				{Keycode: uinput.KeyEqual, Modifier: ModShift}, // ^
				{Keycode: uinput.KeyE, Modifier: ModNone},      // e
			},
			desc: "ê = circumflex + e",
		},
		// Pre-composed characters
		{
			char: 'è',
			expected: []KeySequence{
				{Keycode: uinput.KeyLeftBrace, Modifier: ModNone},
			},
			desc: "è has direct key",
		},
		{
			char: 'à',
			expected: []KeySequence{
				{Keycode: uinput.KeyApostrophe, Modifier: ModNone},
			},
			desc: "à has direct key",
		},
		{
			char: 'ù',
			expected: []KeySequence{
				{Keycode: uinput.KeyBackslash, Modifier: ModNone},
			},
			desc: "ù has direct key",
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			seq, err := layout.CharToKeySequence(ctx, tt.char)
			if err != nil {
				t.Errorf("char %q not supported: %v", tt.char, err)
				return
			}

			if len(seq) != len(tt.expected) {
				t.Errorf("char %q: got %d keystrokes, want %d", tt.char, len(seq), len(tt.expected))
				return
			}

			for i, key := range seq {
				if key.Keycode != tt.expected[i].Keycode || key.Modifier != tt.expected[i].Modifier {
					t.Errorf("char %q keystroke[%d]: got {Keycode: %d, Modifier: %d}, want {Keycode: %d, Modifier: %d}",
						tt.char, i, key.Keycode, key.Modifier, tt.expected[i].Keycode, tt.expected[i].Modifier)
				}
			}
		})
	}
}

// TestBackwardCompatibility ensures that basic characters still work correctly.
func TestBackwardCompatibility(t *testing.T) {
	layouts := map[string]Layout{
		"us": NewUS(),
		"uk": NewUK(),
		"fr": NewFR(),
		"de": NewDE(),
		"es": NewES(),
		"it": NewIT(),
	}

	// Test that basic ASCII characters work on all layouts
	basicChars := "abc ABC 123 !@#"
	ctx := context.Background()

	for name, layout := range layouts {
		t.Run(name, func(t *testing.T) {
			for _, char := range basicChars {
				seq, err := layout.CharToKeySequence(ctx, char)
				if err != nil && char != '@' && char != '#' && char != '!' {
					// Some symbols may not be supported on all layouts, that's ok
					// But basic letters and numbers should always work
					if (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') ||
						(char >= '0' && char <= '9') || char == ' ' {
						t.Errorf("%s layout: char %q not supported: %v", name, char, err)
					}
					continue
				}

				if len(seq) == 0 {
					t.Errorf("%s layout: char %q returned empty sequence", name, char)
				}
			}
		})
	}
}

// TestUnsupportedCharacters ensures that unsupported characters return proper errors.
func TestUnsupportedCharacters(t *testing.T) {
	layout := NewUS()
	ctx := context.Background()

	// US layout should not support accented characters directly
	unsupportedChars := []rune{'ô', 'â', 'ê', 'ñ', 'ü'}

	for _, char := range unsupportedChars {
		t.Run(string(char), func(t *testing.T) {
			_, err := layout.CharToKeySequence(ctx, char)
			if err == nil {
				t.Errorf("char %q should not be supported in US layout", char)
			}

			// Verify error is of correct type
			if _, ok := err.(*ErrCharNotSupported); !ok {
				t.Errorf("expected ErrCharNotSupported, got %T", err)
			}
		})
	}
}
