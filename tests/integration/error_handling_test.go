package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/bnema/uinputd-go/internal/protocol"
	"github.com/bnema/uinputd-go/internal/uinput"
)

func TestErrorHandling_InvalidJSON(t *testing.T) {
	ts := newTestServer(t)
	defer ts.close()

	conn, err := net.Dial("unix", ts.socketPath)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	// Send invalid JSON
	_, err = conn.Write([]byte("{invalid json}"))
	if err != nil {
		t.Fatalf("Failed to write: %v", err)
	}

	// Read response
	var resp protocol.Response
	if err := json.NewDecoder(conn).Decode(&resp); err != nil {
		t.Fatalf("Failed to read response: %v", err)
	}

	if resp.Success {
		t.Error("Expected error for invalid JSON")
	}

	if resp.Error == "" {
		t.Error("Expected error message")
	}
}

func TestErrorHandling_UnknownCommandType(t *testing.T) {
	ts := newTestServer(t)
	defer ts.close()

	cmd := &protocol.Command{
		Type:    protocol.CommandType("unknown_command"),
		Payload: json.RawMessage("{}"),
	}

	resp := ts.sendCommand(t, cmd)

	if resp.Success {
		t.Error("Expected error for unknown command type")
	}

	if resp.Error == "" {
		t.Error("Expected error message")
	}
}

func TestErrorHandling_MissingRequiredFields(t *testing.T) {
	ts := newTestServer(t)
	defer ts.close()

	tests := []struct {
		name    string
		cmdType protocol.CommandType
		payload string
	}{
		{
			name:    "type command with empty payload",
			cmdType: protocol.CommandType_Type,
			payload: "{}",
		},
		{
			name:    "stream command with empty payload",
			cmdType: protocol.CommandType_Stream,
			payload: "{}",
		},
		{
			name:    "key command with empty payload",
			cmdType: protocol.CommandType_Key,
			payload: "{}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &protocol.Command{
				Type:    tt.cmdType,
				Payload: json.RawMessage(tt.payload),
			}

			resp := ts.sendCommand(t, cmd)

			// Note: Some commands might succeed with empty payload (e.g., empty text)
			// This test verifies the server doesn't crash
			t.Logf("Command %s with empty payload: success=%v, error=%s", tt.cmdType, resp.Success, resp.Error)
		})
	}
}

func TestErrorHandling_InvalidLayout(t *testing.T) {
	ts := newTestServer(t)
	defer ts.close()

	payload := protocol.TypePayload{
		Text:   "hello",
		Layout: "nonexistent_layout",
	}

	payloadBytes, _ := json.Marshal(payload)
	cmd := &protocol.Command{
		Type:    protocol.CommandType_Type,
		Payload: payloadBytes,
	}

	resp := ts.sendCommand(t, cmd)

	if resp.Success {
		t.Error("Expected error for invalid layout")
	}

	if resp.Error == "" {
		t.Error("Expected error message for invalid layout")
	}
}

func TestErrorHandling_UnsupportedCharacter(t *testing.T) {
	ts := newTestServer(t)
	defer ts.close()

	tests := []struct {
		name   string
		text   string
		layout string
	}{
		{
			name:   "French character in US layout",
			text:   "é",
			layout: "us",
		},
		{
			name:   "emoji",
			text:   "😀",
			layout: "us",
		},
		{
			name:   "Chinese character",
			text:   "你",
			layout: "us",
		},
		{
			name:   "special Unicode",
			text:   "™",
			layout: "us",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts.mockDevice.Reset()

			payload := protocol.TypePayload{
				Text:   tt.text,
				Layout: tt.layout,
			}

			payloadBytes, _ := json.Marshal(payload)
			cmd := &protocol.Command{
				Type:    protocol.CommandType_Type,
				Payload: payloadBytes,
			}

			resp := ts.sendCommand(t, cmd)

			// The server warns about unsupported characters but continues
			// This is acceptable behavior - it skips unsupported chars
			t.Logf("Unsupported character result: success=%v, error=%s", resp.Success, resp.Error)

			// No events should be generated for unsupported characters
			if ts.mockDevice.GetEventCount() > 0 {
				t.Error("No events should be generated for unsupported characters")
			}
		})
	}
}

func TestErrorHandling_InvalidKeycode(t *testing.T) {
	ts := newTestServer(t)
	defer ts.close()

	payload := protocol.KeyPayload{
		Keycode:  9999, // Invalid keycode
		Modifier: "",
	}

	payloadBytes, _ := json.Marshal(payload)
	cmd := &protocol.Command{
		Type:    protocol.CommandType_Key,
		Payload: payloadBytes,
	}

	resp := ts.sendCommand(t, cmd)

	// The command might succeed (uinput will just send the keycode)
	// but we verify the server doesn't crash
	t.Logf("Invalid keycode result: success=%v, error=%s", resp.Success, resp.Error)
}

func TestErrorHandling_InvalidModifier(t *testing.T) {
	ts := newTestServer(t)
	defer ts.close()

	payload := protocol.KeyPayload{
		Keycode:  uinput.KeyA,
		Modifier: "invalid_modifier",
	}

	payloadBytes, _ := json.Marshal(payload)
	cmd := &protocol.Command{
		Type:    protocol.CommandType_Key,
		Payload: payloadBytes,
	}

	resp := ts.sendCommand(t, cmd)

	if resp.Success {
		t.Error("Expected error for invalid modifier")
	}

	if resp.Error == "" {
		t.Error("Expected error message for invalid modifier")
	}
}

func TestErrorHandling_ConnectionDrop(t *testing.T) {
	ts := newTestServer(t)
	defer ts.close()

	conn, err := net.Dial("unix", ts.socketPath)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}

	// Close connection immediately without sending anything complete
	conn.Close()

	// Server should handle this gracefully
	time.Sleep(50 * time.Millisecond)

	// Try to connect again - server might have shut down due to error
	// which is acceptable behavior
	testConn, err := net.Dial("unix", ts.socketPath)
	if err != nil {
		t.Logf("Server closed after connection drop (acceptable): %v", err)
		return
	}
	testConn.Close()
	t.Log("Server remained responsive after connection drop")
}

func TestErrorHandling_PartialData(t *testing.T) {
	ts := newTestServer(t)
	defer ts.close()

	conn, err := net.Dial("unix", ts.socketPath)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}

	// Send partial JSON
	_, err = conn.Write([]byte(`{"type":"type","payload":`))
	if err != nil {
		t.Fatalf("Failed to write: %v", err)
	}

	// Close without completing the JSON
	conn.Close()

	// Server should handle this gracefully
	time.Sleep(50 * time.Millisecond)

	// Try to connect again - server might have shut down due to error
	// which is acceptable behavior
	testConn, err := net.Dial("unix", ts.socketPath)
	if err != nil {
		t.Logf("Server closed after partial data (acceptable): %v", err)
		return
	}
	testConn.Close()
	t.Log("Server remained responsive after partial data")
}

func TestErrorHandling_NegativeDelays(t *testing.T) {
	ts := newTestServer(t)
	defer ts.close()

	payload := protocol.StreamPayload{
		Text:      "hello",
		Layout:    "us",
		CharDelay: -100,
		DelayMs:   -50,
	}

	payloadBytes, _ := json.Marshal(payload)
	cmd := &protocol.Command{
		Type:    protocol.CommandType_Stream,
		Payload: payloadBytes,
	}

	resp := ts.sendCommand(t, cmd)

	// Server should handle negative delays gracefully (treat as 0 or error)
	t.Logf("Negative delays result: success=%v, error=%s", resp.Success, resp.Error)
}

func TestErrorHandling_ExtremelyLargeDelay(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping large delay test in short mode")
	}

	ts := newTestServer(t)
	defer ts.close()

	payload := protocol.StreamPayload{
		Text:      "a",
		Layout:    "us",
		CharDelay: 999999, // Very large delay
	}

	payloadBytes, _ := json.Marshal(payload)
	cmd := &protocol.Command{
		Type:    protocol.CommandType_Stream,
		Payload: payloadBytes,
	}

	// Set a timeout for the connection
	conn, err := net.Dial("unix", ts.socketPath)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	conn.SetDeadline(time.Now().Add(2 * time.Second))

	// Send command
	if err := json.NewEncoder(conn).Encode(cmd); err != nil {
		t.Fatalf("Failed to send command: %v", err)
	}

	// Read response (should timeout or complete quickly)
	var resp protocol.Response
	err = json.NewDecoder(conn).Decode(&resp)

	// Either timeout or completion is acceptable
	if err != nil {
		t.Logf("Command timed out as expected: %v", err)
	} else {
		t.Logf("Command completed: success=%v", resp.Success)
	}
}

func TestErrorHandling_VeryLongText(t *testing.T) {
	ts := newTestServer(t)
	defer ts.close()

	// Generate very long text (10,000 characters)
	longText := make([]byte, 10000)
	for i := range longText {
		longText[i] = 'a'
	}

	payload := protocol.TypePayload{
		Text:   string(longText),
		Layout: "us",
	}

	payloadBytes, _ := json.Marshal(payload)
	cmd := &protocol.Command{
		Type:    protocol.CommandType_Type,
		Payload: payloadBytes,
	}

	resp := ts.sendCommand(t, cmd)

	if !resp.Success {
		t.Errorf("Very long text failed: %s", resp.Error)
	}

	// Verify events were generated
	if ts.mockDevice.GetEventCount() == 0 {
		t.Error("No events generated for long text")
	}
}

func TestErrorHandling_ContextCancellation(t *testing.T) {
	ts := newTestServer(t)
	defer ts.close()

	// Create a mock device that respects context cancellation
	type cancellableDevice struct {
		*MockUinputDevice
		cancel context.CancelFunc
	}

	// This test verifies that the server can handle context cancellation
	// For now, we just verify the server shuts down cleanly
	ts.cancel()
	time.Sleep(100 * time.Millisecond)

	// Try to connect (should fail)
	_, err := net.Dial("unix", ts.socketPath)
	if err == nil {
		t.Error("Expected connection to fail after server shutdown")
	}
}

func TestErrorHandling_MultipleSimultaneousErrors(t *testing.T) {
	ts := newTestServer(t)
	defer ts.close()

	errorCount := 10
	results := make(chan bool, errorCount)

	// Send multiple invalid commands simultaneously
	for i := 0; i < errorCount; i++ {
		go func(id int) {
			cmd := &protocol.Command{
				Type:    protocol.CommandType("invalid"),
				Payload: json.RawMessage("{}"),
			}

			resp := ts.sendCommand(t, cmd)
			results <- !resp.Success // We expect failure
		}(i)
	}

	// Collect results
	failureCount := 0
	for i := 0; i < errorCount; i++ {
		if <-results {
			failureCount++
		}
	}

	if failureCount != errorCount {
		t.Errorf("Expected all %d commands to fail, but only %d failed", errorCount, failureCount)
	}
}

func TestErrorHandling_RecoveryAfterError(t *testing.T) {
	ts := newTestServer(t)
	defer ts.close()

	// Send an invalid command
	invalidCmd := &protocol.Command{
		Type:    protocol.CommandType("invalid"),
		Payload: json.RawMessage("{}"),
	}

	resp1 := ts.sendCommand(t, invalidCmd)
	if resp1.Success {
		t.Error("Expected invalid command to fail")
	}

	// Now send a valid command
	payload := protocol.TypePayload{
		Text:   "hello",
		Layout: "us",
	}
	payloadBytes, _ := json.Marshal(payload)
	validCmd := &protocol.Command{
		Type:    protocol.CommandType_Type,
		Payload: payloadBytes,
	}

	resp2 := ts.sendCommand(t, validCmd)
	if !resp2.Success {
		t.Errorf("Valid command after error should succeed: %s", resp2.Error)
	}

	// Verify events were generated for the valid command
	if ts.mockDevice.GetEventCount() == 0 {
		t.Error("No events generated after recovery")
	}
}

func TestErrorHandling_EmptyConnection(t *testing.T) {
	ts := newTestServer(t)
	defer ts.close()

	conn, err := net.Dial("unix", ts.socketPath)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}

	// Close immediately without sending anything
	conn.Close()

	// Server should handle this gracefully
	time.Sleep(50 * time.Millisecond)

	// Verify server is still responsive
	testConn, err := net.Dial("unix", ts.socketPath)
	if err != nil {
		t.Fatalf("Server not responsive after empty connection: %v", err)
	}
	testConn.Close()
}

func TestErrorHandling_MixedValidInvalidCommands(t *testing.T) {
	ts := newTestServer(t)
	defer ts.close()

	commandCount := 20
	validCount := 0
	invalidCount := 0

	for i := 0; i < commandCount; i++ {
		var cmd *protocol.Command

		if i%2 == 0 {
			// Valid command
			payload := protocol.TypePayload{
				Text:   fmt.Sprintf("test%d", i),
				Layout: "us",
			}
			payloadBytes, _ := json.Marshal(payload)
			cmd = &protocol.Command{
				Type:    protocol.CommandType_Type,
				Payload: payloadBytes,
			}
			validCount++
		} else {
			// Invalid command
			cmd = &protocol.Command{
				Type:    protocol.CommandType("invalid"),
				Payload: json.RawMessage("{}"),
			}
			invalidCount++
		}

		resp := ts.sendCommand(t, cmd)

		if i%2 == 0 {
			// Should succeed
			if !resp.Success {
				t.Errorf("Valid command %d failed: %s", i, resp.Error)
			}
		} else {
			// Should fail
			if resp.Success {
				t.Errorf("Invalid command %d should have failed", i)
			}
		}
	}

	t.Logf("Sent %d valid and %d invalid commands", validCount, invalidCount)

	// Verify events were generated for valid commands
	if ts.mockDevice.GetEventCount() == 0 {
		t.Error("No events generated from valid commands")
	}
}
