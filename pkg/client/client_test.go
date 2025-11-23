package client

import (
	"context"
	"encoding/json"
	"net"
	"testing"
	"time"

	"github.com/bnema/uinputd-go/internal/protocol"
)

// mockServer simulates the uinputd daemon for testing
type mockServer struct {
	listener net.Listener
	handler  func(protocol.Command) protocol.Response
}

func newMockServer(t *testing.T, handler func(protocol.Command) protocol.Response) *mockServer {
	listener, err := net.Listen("unix", t.TempDir()+"/test.sock")
	if err != nil {
		t.Fatalf("Failed to create mock server: %v", err)
	}

	ms := &mockServer{
		listener: listener,
		handler:  handler,
	}

	go ms.serve()

	return ms
}

func (ms *mockServer) serve() {
	for {
		conn, err := ms.listener.Accept()
		if err != nil {
			return // Server closed
		}

		go ms.handleConnection(conn)
	}
}

func (ms *mockServer) handleConnection(conn net.Conn) {
	defer conn.Close()

	var cmd protocol.Command
	if err := json.NewDecoder(conn).Decode(&cmd); err != nil {
		return
	}

	resp := ms.handler(cmd)
	json.NewEncoder(conn).Encode(resp)
}

func (ms *mockServer) close() {
	ms.listener.Close()
}

func (ms *mockServer) addr() string {
	return ms.listener.Addr().String()
}

func TestNew(t *testing.T) {
	tests := []struct {
		name       string
		socketPath string
		opts       *Options
		wantErr    bool
	}{
		{
			name:       "default options",
			socketPath: "/tmp/test.sock",
			opts:       nil,
			wantErr:    false,
		},
		{
			name:       "custom timeout",
			socketPath: "/tmp/test.sock",
			opts:       &Options{Timeout: 10 * time.Second},
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := New(tt.socketPath, tt.opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && client == nil {
				t.Error("New() returned nil client")
			}
			if client != nil {
				client.Close()
			}
		})
	}
}

func TestClient_TypeText(t *testing.T) {
	tests := []struct {
		name    string
		text    string
		opts    *TypeOptions
		wantCmd protocol.CommandType
		wantErr bool
	}{
		{
			name:    "simple text",
			text:    "Hello",
			opts:    nil,
			wantCmd: protocol.CommandType_Type,
			wantErr: false,
		},
		{
			name:    "with layout",
			text:    "Bonjour",
			opts:    &TypeOptions{Layout: "fr"},
			wantCmd: protocol.CommandType_Type,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var receivedCmd protocol.Command

			server := newMockServer(t, func(cmd protocol.Command) protocol.Response {
				receivedCmd = cmd
				return protocol.Response{Success: true}
			})
			defer server.close()

			client, err := New(server.addr(), nil)
			if err != nil {
				t.Fatalf("Failed to create client: %v", err)
			}
			defer client.Close()

			ctx := context.Background()
			err = client.TypeText(ctx, tt.text, tt.opts)

			if (err != nil) != tt.wantErr {
				t.Errorf("TypeText() error = %v, wantErr %v", err, tt.wantErr)
			}

			if receivedCmd.Type != tt.wantCmd {
				t.Errorf("Expected command type %v, got %v", tt.wantCmd, receivedCmd.Type)
			}

			var payload protocol.TypePayload
			if err := json.Unmarshal(receivedCmd.Payload, &payload); err != nil {
				t.Fatalf("Failed to unmarshal payload: %v", err)
			}

			if payload.Text != tt.text {
				t.Errorf("Expected text %q, got %q", tt.text, payload.Text)
			}

			if tt.opts != nil && payload.Layout != tt.opts.Layout {
				t.Errorf("Expected layout %q, got %q", tt.opts.Layout, payload.Layout)
			}
		})
	}
}

func TestClient_StreamText(t *testing.T) {
	server := newMockServer(t, func(cmd protocol.Command) protocol.Response {
		if cmd.Type != protocol.CommandType_Stream {
			return protocol.Response{Success: false, Error: "wrong command type"}
		}
		return protocol.Response{Success: true}
	})
	defer server.close()

	client, err := New(server.addr(), nil)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	ctx := context.Background()
	opts := &StreamOptions{
		Layout:    "us",
		DelayMs:   50,
		CharDelay: 10,
	}

	err = client.StreamText(ctx, "Hello", opts)
	if err != nil {
		t.Errorf("StreamText() error = %v", err)
	}
}

func TestClient_SendKey(t *testing.T) {
	tests := []struct {
		name     string
		keycode  uint16
		modifier KeyModifier
		wantErr  bool
	}{
		{
			name:     "plain key",
			keycode:  28,
			modifier: ModifierNone,
			wantErr:  false,
		},
		{
			name:     "with shift",
			keycode:  30,
			modifier: ModifierShift,
			wantErr:  false,
		},
		{
			name:     "with ctrl",
			keycode:  46,
			modifier: ModifierCtrl,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var receivedPayload protocol.KeyPayload

			server := newMockServer(t, func(cmd protocol.Command) protocol.Response {
				if cmd.Type != protocol.CommandType_Key {
					return protocol.Response{Success: false, Error: "wrong command type"}
				}

				if err := json.Unmarshal(cmd.Payload, &receivedPayload); err != nil {
					return protocol.Response{Success: false, Error: err.Error()}
				}

				return protocol.Response{Success: true}
			})
			defer server.close()

			client, err := New(server.addr(), nil)
			if err != nil {
				t.Fatalf("Failed to create client: %v", err)
			}
			defer client.Close()

			ctx := context.Background()
			err = client.SendKey(ctx, tt.keycode, tt.modifier)

			if (err != nil) != tt.wantErr {
				t.Errorf("SendKey() error = %v, wantErr %v", err, tt.wantErr)
			}

			if receivedPayload.Keycode != tt.keycode {
				t.Errorf("Expected keycode %d, got %d", tt.keycode, receivedPayload.Keycode)
			}

			if receivedPayload.Modifier != string(tt.modifier) {
				t.Errorf("Expected modifier %q, got %q", tt.modifier, receivedPayload.Modifier)
			}
		})
	}
}

func TestClient_Ping(t *testing.T) {
	server := newMockServer(t, func(cmd protocol.Command) protocol.Response {
		if cmd.Type != protocol.CommandType_Ping {
			return protocol.Response{Success: false, Error: "wrong command type"}
		}
		return protocol.Response{Success: true}
	})
	defer server.close()

	client, err := New(server.addr(), nil)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	ctx := context.Background()
	if err := client.Ping(ctx); err != nil {
		t.Errorf("Ping() error = %v", err)
	}
}

func TestClient_DaemonError(t *testing.T) {
	server := newMockServer(t, func(cmd protocol.Command) protocol.Response {
		return protocol.Response{
			Success: false,
			Error:   "test error from daemon",
		}
	})
	defer server.close()

	client, err := New(server.addr(), nil)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	ctx := context.Background()
	err = client.TypeText(ctx, "test", nil)

	if err == nil {
		t.Error("Expected error from daemon, got nil")
	}

	if err != nil && err.Error() != "daemon error: test error from daemon" {
		t.Errorf("Unexpected error message: %v", err)
	}
}

func TestClient_ContextTimeout(t *testing.T) {
	// Server that doesn't respond
	server := newMockServer(t, func(cmd protocol.Command) protocol.Response {
		time.Sleep(2 * time.Second)
		return protocol.Response{Success: true}
	})
	defer server.close()

	client, err := New(server.addr(), &Options{Timeout: 100 * time.Millisecond})
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	err = client.TypeText(ctx, "test", nil)
	if err == nil {
		t.Error("Expected timeout error, got nil")
	}
}

func TestClient_IsConnected(t *testing.T) {
	server := newMockServer(t, func(cmd protocol.Command) protocol.Response {
		return protocol.Response{Success: true}
	})
	defer server.close()

	client, err := New(server.addr(), nil)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	if client.IsConnected() {
		t.Error("Client should not be connected initially")
	}

	ctx := context.Background()
	if err := client.Ping(ctx); err != nil {
		t.Fatalf("Ping failed: %v", err)
	}

	if !client.IsConnected() {
		t.Error("Client should be connected after Ping")
	}

	client.Close()

	if client.IsConnected() {
		t.Error("Client should not be connected after Close")
	}
}
