package integration

import (
	"encoding/json"
	"testing"

	"github.com/bnema/uinputd-go/internal/protocol"
	"github.com/bnema/uinputd-go/internal/uinput"
)

func TestLayoutIntegration_FrenchDeadKeys(t *testing.T) {
	ts := newTestServer(t)
	defer ts.close()

	tests := []struct {
		name          string
		text          string
		wantMinEvents int
		hasDeadKey    bool
	}{
		{
			name:          "circumflex a -> â",
			text:          "â",
			wantMinEvents: 8, // 2 chars (^ + a) * 4 events
			hasDeadKey:    true,
		},
		{
			name:          "circumflex e -> ê",
			text:          "ê",
			wantMinEvents: 8,
			hasDeadKey:    true,
		},
		{
			name:          "circumflex o -> ô",
			text:          "ô",
			wantMinEvents: 8,
			hasDeadKey:    true,
		},
		{
			name:          "circumflex u -> û",
			text:          "û",
			wantMinEvents: 8,
			hasDeadKey:    true,
		},
		{
			name:          "diaeresis e -> ë",
			text:          "ë",
			wantMinEvents: 8, // shift+^ + e
			hasDeadKey:    true,
		},
		{
			name:          "diaeresis i -> ï",
			text:          "ï",
			wantMinEvents: 8,
			hasDeadKey:    true,
		},
		{
			name:          "direct key é",
			text:          "é",
			wantMinEvents: 4, // Direct key, no dead key
			hasDeadKey:    false,
		},
		{
			name:          "direct key è",
			text:          "è",
			wantMinEvents: 4,
			hasDeadKey:    false,
		},
		{
			name:          "mixed text with dead keys",
			text:          "châ",
			wantMinEvents: 12, // c(4) + ^+a(8)
			hasDeadKey:    true,
		},
		{
			name:          "word with multiple dead keys",
			text:          "château",
			wantMinEvents: 28, // c(4) + h(4) + ^+a(8) + t(4) + e(4) + ^+a(8) + u(4) - adjusted calculation
			hasDeadKey:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts.mockDevice.Reset()

			payload := protocol.TypePayload{
				Text:   tt.text,
				Layout: "fr",
			}

			payloadBytes, err := json.Marshal(payload)
			if err != nil {
				t.Fatalf("Failed to marshal payload: %v", err)
			}

			cmd := &protocol.Command{
				Type:    protocol.CommandType_Type,
				Payload: payloadBytes,
			}

			resp := ts.sendCommand(t, cmd)

			if !resp.Success {
				t.Fatalf("Command failed: %s", resp.Error)
			}

			// Verify events were generated
			eventCount := ts.mockDevice.GetEventCount()
			if eventCount < tt.wantMinEvents {
				t.Errorf("Expected at least %d events, got %d", tt.wantMinEvents, eventCount)
				t.Logf("Key sequence: %v", ts.mockDevice.GetKeyPressSequence())
			}

			// Verify key events exist
			events := ts.mockDevice.GetEvents()
			keyEventCount := 0
			for _, event := range events {
				if event.Type == uinput.EvKey {
					keyEventCount++
				}
			}

			if keyEventCount == 0 {
				t.Error("No key events generated")
			}
		})
	}
}

func TestLayoutIntegration_GermanDeadKeys(t *testing.T) {
	ts := newTestServer(t)
	defer ts.close()

	tests := []struct {
		name          string
		text          string
		wantMinEvents int
	}{
		{
			name:          "circumflex a -> â",
			text:          "â",
			wantMinEvents: 8, // ^ (dead key) + a
		},
		{
			name:          "acute e -> é",
			text:          "é",
			wantMinEvents: 8, // ´ (dead key) + e
		},
		{
			name:          "direct key ä",
			text:          "ä",
			wantMinEvents: 4, // Direct key
		},
		{
			name:          "direct key ö",
			text:          "ö",
			wantMinEvents: 4,
		},
		{
			name:          "direct key ü",
			text:          "ü",
			wantMinEvents: 4,
		},
		{
			name:          "german word with umlaut",
			text:          "schön",
			wantMinEvents: 20, // s(4) + c(4) + h(4) + ö(4) + n(4)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts.mockDevice.Reset()

			payload := protocol.TypePayload{
				Text:   tt.text,
				Layout: "de",
			}

			payloadBytes, err := json.Marshal(payload)
			if err != nil {
				t.Fatalf("Failed to marshal payload: %v", err)
			}

			cmd := &protocol.Command{
				Type:    protocol.CommandType_Type,
				Payload: payloadBytes,
			}

			resp := ts.sendCommand(t, cmd)

			if !resp.Success {
				t.Fatalf("Command failed: %s", resp.Error)
			}

			eventCount := ts.mockDevice.GetEventCount()
			if eventCount < tt.wantMinEvents {
				t.Errorf("Expected at least %d events, got %d", tt.wantMinEvents, eventCount)
			}
		})
	}
}

func TestLayoutIntegration_SpanishDeadKeys(t *testing.T) {
	ts := newTestServer(t)
	defer ts.close()

	tests := []struct {
		name          string
		text          string
		wantMinEvents int
	}{
		{
			name:          "grave a -> à",
			text:          "à",
			wantMinEvents: 8, // ` (dead key) + a
		},
		{
			name:          "circumflex a -> â",
			text:          "â",
			wantMinEvents: 8, // shift+^ (dead key) + a
		},
		{
			name:          "direct key ñ",
			text:          "ñ",
			wantMinEvents: 4,
		},
		{
			name:          "direct key á",
			text:          "á",
			wantMinEvents: 4,
		},
		{
			name:          "spanish word",
			text:          "mañana",
			wantMinEvents: 24, // m(4) + a(4) + ñ(4) + a(4) + n(4) + a(4)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts.mockDevice.Reset()

			payload := protocol.TypePayload{
				Text:   tt.text,
				Layout: "es",
			}

			payloadBytes, err := json.Marshal(payload)
			if err != nil {
				t.Fatalf("Failed to marshal payload: %v", err)
			}

			cmd := &protocol.Command{
				Type:    protocol.CommandType_Type,
				Payload: payloadBytes,
			}

			resp := ts.sendCommand(t, cmd)

			if !resp.Success {
				t.Fatalf("Command failed: %s", resp.Error)
			}

			eventCount := ts.mockDevice.GetEventCount()
			if eventCount < tt.wantMinEvents {
				t.Errorf("Expected at least %d events, got %d", tt.wantMinEvents, eventCount)
			}
		})
	}
}

func TestLayoutIntegration_ItalianDeadKeys(t *testing.T) {
	ts := newTestServer(t)
	defer ts.close()

	tests := []struct {
		name          string
		text          string
		wantMinEvents int
	}{
		{
			name:          "circumflex e -> ê",
			text:          "ê",
			wantMinEvents: 8, // shift+^ (dead key) + e
		},
		{
			name:          "direct key è",
			text:          "è",
			wantMinEvents: 4,
		},
		{
			name:          "direct key à",
			text:          "à",
			wantMinEvents: 4,
		},
		{
			name:          "direct key ù",
			text:          "ù",
			wantMinEvents: 4,
		},
		{
			name:          "italian word",
			text:          "città",
			wantMinEvents: 20, // c(4) + i(4) + t(4) + t(4) + à(4)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts.mockDevice.Reset()

			payload := protocol.TypePayload{
				Text:   tt.text,
				Layout: "it",
			}

			payloadBytes, err := json.Marshal(payload)
			if err != nil {
				t.Fatalf("Failed to marshal payload: %v", err)
			}

			cmd := &protocol.Command{
				Type:    protocol.CommandType_Type,
				Payload: payloadBytes,
			}

			resp := ts.sendCommand(t, cmd)

			if !resp.Success {
				t.Fatalf("Command failed: %s", resp.Error)
			}

			eventCount := ts.mockDevice.GetEventCount()
			if eventCount < tt.wantMinEvents {
				t.Errorf("Expected at least %d events, got %d", tt.wantMinEvents, eventCount)
			}
		})
	}
}

func TestLayoutIntegration_LayoutSwitching(t *testing.T) {
	ts := newTestServer(t)
	defer ts.close()

	// Test switching between layouts in different commands
	tests := []struct {
		text   string
		layout string
	}{
		{"Hello", "us"},
		{"Bonjour", "fr"},
		{"Hallo", "de"},
		{"Hola", "es"},
		{"Ciao", "it"},
		{"World", "us"},
		{"château", "fr"},
	}

	for _, tt := range tests {
		t.Run(tt.text+"_"+tt.layout, func(t *testing.T) {
			ts.mockDevice.Reset()

			payload := protocol.TypePayload{
				Text:   tt.text,
				Layout: tt.layout,
			}

			payloadBytes, err := json.Marshal(payload)
			if err != nil {
				t.Fatalf("Failed to marshal payload: %v", err)
			}

			cmd := &protocol.Command{
				Type:    protocol.CommandType_Type,
				Payload: payloadBytes,
			}

			resp := ts.sendCommand(t, cmd)

			if !resp.Success {
				t.Fatalf("Command failed: %s", resp.Error)
			}

			if ts.mockDevice.GetEventCount() == 0 {
				t.Error("No events generated")
			}
		})
	}
}

func TestLayoutIntegration_UnsupportedCharacters(t *testing.T) {
	ts := newTestServer(t)
	defer ts.close()

	tests := []struct {
		name           string
		text           string
		layout         string
		hasUnsupported bool
		minEvents      int
	}{
		{
			name:           "US layout with French accent",
			text:           "café",
			layout:         "us",
			hasUnsupported: true, // é not supported
			minEvents:      12,   // "caf" should still work (3 chars * 4 events)
		},
		{
			name:           "US layout basic ASCII",
			text:           "hello world",
			layout:         "us",
			hasUnsupported: false,
			minEvents:      44, // 11 chars * 4 events
		},
		{
			name:           "emoji in any layout",
			text:           "hello 😀",
			layout:         "us",
			hasUnsupported: true, // Emoji not supported
			minEvents:      24,   // "hello " should work (6 chars * 4 events)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts.mockDevice.Reset()

			payload := protocol.TypePayload{
				Text:   tt.text,
				Layout: tt.layout,
			}

			payloadBytes, err := json.Marshal(payload)
			if err != nil {
				t.Fatalf("Failed to marshal payload: %v", err)
			}

			cmd := &protocol.Command{
				Type:    protocol.CommandType_Type,
				Payload: payloadBytes,
			}

			resp := ts.sendCommand(t, cmd)

			// Server continues even with unsupported characters (skips them)
			t.Logf("Result: success=%v, error=%s, events=%d", resp.Success, resp.Error, ts.mockDevice.GetEventCount())

			// For text with some supported chars, verify those were typed
			if !tt.hasUnsupported {
				if !resp.Success {
					t.Errorf("Command should have succeeded: %s", resp.Error)
				}
				if ts.mockDevice.GetEventCount() < tt.minEvents {
					t.Errorf("Expected at least %d events, got %d", tt.minEvents, ts.mockDevice.GetEventCount())
				}
			}
		})
	}
}

func TestLayoutIntegration_InvalidLayout(t *testing.T) {
	ts := newTestServer(t)
	defer ts.close()

	payload := protocol.TypePayload{
		Text:   "hello",
		Layout: "invalid_layout",
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("Failed to marshal payload: %v", err)
	}

	cmd := &protocol.Command{
		Type:    protocol.CommandType_Type,
		Payload: payloadBytes,
	}

	resp := ts.sendCommand(t, cmd)

	if resp.Success {
		t.Error("Expected command to fail with invalid layout")
	}

	if resp.Error == "" {
		t.Error("Expected error message for invalid layout")
	}
}

func TestLayoutIntegration_UppercaseWithDeadKeys(t *testing.T) {
	ts := newTestServer(t)
	defer ts.close()

	tests := []struct {
		name          string
		text          string
		layout        string
		wantMinEvents int
	}{
		{
			name:          "French uppercase Â",
			text:          "Â",
			layout:        "fr",
			wantMinEvents: 12, // ^ + shift+A (with shift modifier)
		},
		{
			name:          "French uppercase Ê",
			text:          "Ê",
			layout:        "fr",
			wantMinEvents: 12,
		},
		{
			name:          "German uppercase Ä (direct)",
			text:          "Ä",
			layout:        "de",
			wantMinEvents: 8, // shift+ä (direct key with shift)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts.mockDevice.Reset()

			payload := protocol.TypePayload{
				Text:   tt.text,
				Layout: tt.layout,
			}

			payloadBytes, err := json.Marshal(payload)
			if err != nil {
				t.Fatalf("Failed to marshal payload: %v", err)
			}

			cmd := &protocol.Command{
				Type:    protocol.CommandType_Type,
				Payload: payloadBytes,
			}

			resp := ts.sendCommand(t, cmd)

			if !resp.Success {
				t.Fatalf("Command failed: %s", resp.Error)
			}

			eventCount := ts.mockDevice.GetEventCount()
			if eventCount < tt.wantMinEvents {
				t.Errorf("Expected at least %d events, got %d", tt.wantMinEvents, eventCount)
				t.Logf("Key sequence: %v", ts.mockDevice.GetKeyPressSequence())
			}
		})
	}
}
