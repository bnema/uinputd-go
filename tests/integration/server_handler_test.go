package integration

import (
	"context"
	"encoding/json"
	"net"
	"path/filepath"
	"testing"
	"time"

	"github.com/bnema/uinputd-go/internal/config"
	"github.com/bnema/uinputd-go/internal/protocol"
	"github.com/bnema/uinputd-go/internal/server"
	"github.com/bnema/uinputd-go/internal/uinput"
)

// testServer wraps server and mock device for testing.
type testServer struct {
	server     *server.Server
	mockDevice *MockUinputDevice
	ctx        context.Context
	cancel     context.CancelFunc
	socketPath string
}

// newTestServer creates a new test server with mock device.
func newTestServer(t *testing.T) *testServer {
	t.Helper()

	mockDevice := NewMockUinputDevice()
	ctx, cancel := context.WithCancel(context.Background())

	// Create test socket path
	socketPath := filepath.Join(t.TempDir(), "test.sock")

	// Create minimal config
	cfg := &config.Config{
		Socket: config.SocketConfig{
			Path:        socketPath,
			Permissions: 0600,
		},
		Layout: "us",
	}

	// Create server with mock device
	srv, err := server.New(ctx, cfg, mockDevice)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	// Start server in background
	go func() {
		if err := srv.Start(ctx); err != nil && ctx.Err() == nil {
			t.Logf("Server error: %v", err)
		}
	}()

	// Wait for server to be ready
	time.Sleep(50 * time.Millisecond)

	return &testServer{
		server:     srv,
		mockDevice: mockDevice,
		ctx:        ctx,
		cancel:     cancel,
		socketPath: socketPath,
	}
}

// close shuts down the test server.
func (ts *testServer) close() {
	ts.cancel()
	ts.server.Close()
	ts.mockDevice.Close()
}

// sendCommand sends a command to the server and returns the response.
func (ts *testServer) sendCommand(t *testing.T, cmd *protocol.Command) *protocol.Response {
	t.Helper()

	conn, err := net.Dial("unix", ts.socketPath)
	if err != nil {
		t.Fatalf("Failed to connect to server: %v", err)
	}
	defer conn.Close()

	// Send command
	if err := json.NewEncoder(conn).Encode(cmd); err != nil {
		t.Fatalf("Failed to send command: %v", err)
	}

	// Read response
	var resp protocol.Response
	if err := json.NewDecoder(conn).Decode(&resp); err != nil {
		t.Fatalf("Failed to read response: %v", err)
	}

	return &resp
}

func TestServerHandler_TypeCommand(t *testing.T) {
	ts := newTestServer(t)
	defer ts.close()

	tests := []struct {
		name         string
		text         string
		layout       string
		wantMinEvents int // Minimum expected events (press+syn+release+syn per key)
	}{
		{
			name:         "simple ASCII text",
			text:         "hello",
			layout:       "us",
			wantMinEvents: 20, // 5 chars * 4 events (press+syn+release+syn)
		},
		{
			name:         "text with space",
			text:         "hi world",
			layout:       "us",
			wantMinEvents: 32, // 8 chars * 4 events
		},
		{
			name:         "uppercase letters",
			text:         "HELLO",
			layout:       "us",
			wantMinEvents: 40, // 5 chars * 8 events (shift modifier adds 4 events)
		},
		{
			name:         "numbers",
			text:         "12345",
			layout:       "us",
			wantMinEvents: 20, // 5 numbers * 4 events
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

			// Verify events were generated
			eventCount := ts.mockDevice.GetEventCount()
			if eventCount < tt.wantMinEvents {
				t.Errorf("Expected at least %d events, got %d", tt.wantMinEvents, eventCount)
			}

			// Verify we got key events (not just syn)
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

func TestServerHandler_StreamCommand(t *testing.T) {
	ts := newTestServer(t)
	defer ts.close()

	tests := []struct {
		name      string
		text      string
		layout    string
		charDelay int
		wordDelay int
	}{
		{
			name:      "stream with char delay",
			text:      "hello",
			layout:    "us",
			charDelay: 10,
			wordDelay: 0,
		},
		{
			name:      "stream with word delay",
			text:      "hello world",
			layout:    "us",
			charDelay: 0,
			wordDelay: 20,
		},
		{
			name:      "stream with both delays",
			text:      "hi there",
			layout:    "us",
			charDelay: 5,
			wordDelay: 10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts.mockDevice.Reset()

			payload := protocol.StreamPayload{
				Text:      tt.text,
				Layout:    tt.layout,
				CharDelay: tt.charDelay,
				DelayMs:   tt.wordDelay,
			}

			payloadBytes, err := json.Marshal(payload)
			if err != nil {
				t.Fatalf("Failed to marshal payload: %v", err)
			}

			cmd := &protocol.Command{
				Type:    protocol.CommandType_Stream,
				Payload: payloadBytes,
			}

			start := time.Now()
			resp := ts.sendCommand(t, cmd)
			elapsed := time.Since(start)

			if !resp.Success {
				t.Fatalf("Command failed: %s", resp.Error)
			}

			// Verify events were generated
			eventCount := ts.mockDevice.GetEventCount()
			if eventCount == 0 {
				t.Error("No events generated")
			}

			// Verify timing if delays are specified
			if tt.charDelay > 0 || tt.wordDelay > 0 {
				// Should take at least some time with delays
				minExpected := time.Duration(tt.charDelay) * time.Millisecond
				if elapsed < minExpected/2 { // Allow some tolerance
					t.Logf("Warning: Expected delay of at least %v, command completed in %v", minExpected/2, elapsed)
				}
			}
		})
	}
}

func TestServerHandler_KeyCommand(t *testing.T) {
	ts := newTestServer(t)
	defer ts.close()

	tests := []struct {
		name          string
		keycode       uint16
		modifier      string
		wantEvents    int
		verifySeq     []EventSequence
	}{
		{
			name:       "simple key press",
			keycode:    uinput.KeyA,
			modifier:   "",
			wantEvents: 4, // press+syn+release+syn
			verifySeq: []EventSequence{
				{Keycode: uinput.KeyA, Pressed: true, IsSyn: false},
				{IsSyn: true},
				{Keycode: uinput.KeyA, Pressed: false, IsSyn: false},
				{IsSyn: true},
			},
		},
		{
			name:       "key with shift",
			keycode:    uinput.KeyA,
			modifier:   "shift",
			wantEvents: 8, // modifier_press+syn+key_press+syn+key_release+syn+modifier_release+syn
			verifySeq: []EventSequence{
				{Keycode: uinput.KeyLeftShift, Pressed: true, IsSyn: false, Modifier: true},
				{IsSyn: true},
				{Keycode: uinput.KeyA, Pressed: true, IsSyn: false},
				{IsSyn: true},
				{Keycode: uinput.KeyA, Pressed: false, IsSyn: false},
				{IsSyn: true},
				{Keycode: uinput.KeyLeftShift, Pressed: false, IsSyn: false, Modifier: true},
				{IsSyn: true},
			},
		},
		{
			name:       "key with ctrl",
			keycode:    uinput.KeyC,
			modifier:   "ctrl",
			wantEvents: 8,
			verifySeq: []EventSequence{
				{Keycode: uinput.KeyLeftCtrl, Pressed: true, IsSyn: false, Modifier: true},
				{IsSyn: true},
				{Keycode: uinput.KeyC, Pressed: true, IsSyn: false},
				{IsSyn: true},
				{Keycode: uinput.KeyC, Pressed: false, IsSyn: false},
				{IsSyn: true},
				{Keycode: uinput.KeyLeftCtrl, Pressed: false, IsSyn: false, Modifier: true},
				{IsSyn: true},
			},
		},
		{
			name:       "key with alt",
			keycode:    62, // F4 key
			modifier:   "alt",
			wantEvents: 8,
			verifySeq: []EventSequence{
				{Keycode: uinput.KeyLeftAlt, Pressed: true, IsSyn: false, Modifier: true},
				{IsSyn: true},
				{Keycode: 62, Pressed: true, IsSyn: false}, // F4 key
				{IsSyn: true},
				{Keycode: 62, Pressed: false, IsSyn: false},
				{IsSyn: true},
				{Keycode: uinput.KeyLeftAlt, Pressed: false, IsSyn: false, Modifier: true},
				{IsSyn: true},
			},
		},
		{
			name:       "key with altgr",
			keycode:    uinput.KeyE,
			modifier:   "altgr",
			wantEvents: 8,
			verifySeq: []EventSequence{
				{Keycode: uinput.KeyRightAlt, Pressed: true, IsSyn: false, Modifier: true},
				{IsSyn: true},
				{Keycode: uinput.KeyE, Pressed: true, IsSyn: false},
				{IsSyn: true},
				{Keycode: uinput.KeyE, Pressed: false, IsSyn: false},
				{IsSyn: true},
				{Keycode: uinput.KeyRightAlt, Pressed: false, IsSyn: false, Modifier: true},
				{IsSyn: true},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts.mockDevice.Reset()

			payload := protocol.KeyPayload{
				Keycode:  tt.keycode,
				Modifier: tt.modifier,
			}

			payloadBytes, err := json.Marshal(payload)
			if err != nil {
				t.Fatalf("Failed to marshal payload: %v", err)
			}

			cmd := &protocol.Command{
				Type:    protocol.CommandType_Key,
				Payload: payloadBytes,
			}

			resp := ts.sendCommand(t, cmd)

			if !resp.Success {
				t.Fatalf("Command failed: %s", resp.Error)
			}

			// Verify event count
			eventCount := ts.mockDevice.GetEventCount()
			if eventCount != tt.wantEvents {
				t.Errorf("Expected %d events, got %d", tt.wantEvents, eventCount)
			}

			// Verify exact event sequence
			if err := ts.mockDevice.VerifyEventSequence(tt.verifySeq); err != nil {
				t.Errorf("Event sequence verification failed: %v", err)
				t.Logf("Got sequence: %v", ts.mockDevice.GetKeyPressSequence())
			}
		})
	}
}

func TestServerHandler_PingCommand(t *testing.T) {
	ts := newTestServer(t)
	defer ts.close()

	payload := protocol.PingPayload{}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("Failed to marshal payload: %v", err)
	}

	cmd := &protocol.Command{
		Type:    protocol.CommandType_Ping,
		Payload: payloadBytes,
	}

	resp := ts.sendCommand(t, cmd)

	if !resp.Success {
		t.Fatalf("Ping failed: %s", resp.Error)
	}

	// Ping should not generate any events
	if ts.mockDevice.GetEventCount() != 0 {
		t.Errorf("Ping command should not generate events, got %d", ts.mockDevice.GetEventCount())
	}
}

func TestServerHandler_InvalidCommand(t *testing.T) {
	ts := newTestServer(t)
	defer ts.close()

	cmd := &protocol.Command{
		Type:    protocol.CommandType("invalid"),
		Payload: []byte("{}"),
	}

	resp := ts.sendCommand(t, cmd)

	if resp.Success {
		t.Error("Expected command to fail with invalid type")
	}

	if resp.Error == "" {
		t.Error("Expected error message")
	}
}

func TestServerHandler_MalformedPayload(t *testing.T) {
	ts := newTestServer(t)
	defer ts.close()

	// Send command with valid JSON structure but invalid payload content
	cmd := &protocol.Command{
		Type:    protocol.CommandType_Type,
		Payload: json.RawMessage(`{"invalid": "structure"}`), // Valid JSON but missing required fields
	}

	resp := ts.sendCommand(t, cmd)

	// This might succeed with empty text, which is acceptable
	// The test mainly verifies the server doesn't crash
	t.Logf("Malformed payload result: success=%v, error=%s", resp.Success, resp.Error)
}

func TestServerHandler_EmptyText(t *testing.T) {
	ts := newTestServer(t)
	defer ts.close()

	payload := protocol.TypePayload{
		Text:   "",
		Layout: "us",
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

	// Empty text should succeed but generate no events
	if !resp.Success {
		t.Errorf("Empty text should succeed: %s", resp.Error)
	}

	if ts.mockDevice.GetEventCount() != 0 {
		t.Errorf("Empty text should not generate events, got %d", ts.mockDevice.GetEventCount())
	}
}
