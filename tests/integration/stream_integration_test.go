package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/bnema/uinputd-go/internal/protocol"
)

// mockDaemon simulates the uinputd daemon for integration testing
type mockDaemon struct {
	listener     net.Listener
	receivedCmds []protocol.Command
	t            *testing.T
}

func newMockDaemon(t *testing.T) *mockDaemon {
	listener, err := net.Listen("unix", filepath.Join(t.TempDir(), "test.sock"))
	if err != nil {
		t.Fatalf("Failed to create mock daemon: %v", err)
	}

	md := &mockDaemon{
		listener:     listener,
		receivedCmds: make([]protocol.Command, 0),
		t:            t,
	}

	go md.serve()

	return md
}

func (md *mockDaemon) serve() {
	for {
		conn, err := md.listener.Accept()
		if err != nil {
			return // Server closed
		}

		go md.handleConnection(conn)
	}
}

func (md *mockDaemon) handleConnection(conn net.Conn) {
	defer conn.Close()

	var cmd protocol.Command
	if err := json.NewDecoder(conn).Decode(&cmd); err != nil {
		md.t.Logf("Failed to decode command: %v", err)
		return
	}

	md.receivedCmds = append(md.receivedCmds, cmd)

	resp := protocol.Response{Success: true, Message: "command executed successfully"}
	if err := json.NewEncoder(conn).Encode(resp); err != nil {
		md.t.Logf("Failed to send response: %v", err)
	}
}

func (md *mockDaemon) close() {
	md.listener.Close()
}

func (md *mockDaemon) addr() string {
	return md.listener.Addr().String()
}

func (md *mockDaemon) getLastCommand() *protocol.Command {
	if len(md.receivedCmds) == 0 {
		return nil
	}
	return &md.receivedCmds[len(md.receivedCmds)-1]
}

func (md *mockDaemon) getCommands() []protocol.Command {
	return md.receivedCmds
}

// getClientBinary returns the path to the uinput-client binary
func getClientBinary(t *testing.T) string {
	// Try to find the binary in the bin directory
	binPath := "../../bin/uinput-client"
	if _, err := os.Stat(binPath); err == nil {
		absPath, _ := filepath.Abs(binPath)
		return absPath
	}

	// Build if not found
	t.Log("Building uinput-client for testing...")
	cmd := exec.Command("make", "build")
	cmd.Dir = "../../"
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to build uinput-client: %v", err)
	}

	absPath, _ := filepath.Abs(binPath)
	return absPath
}

func TestStreamCommand_StdinIntegration(t *testing.T) {
	daemon := newMockDaemon(t)
	defer daemon.close()

	clientBin := getClientBinary(t)

	tests := []struct {
		name         string
		input        string
		expectedText string
	}{
		{
			name:         "single line",
			input:        "Hello world\n",
			expectedText: "Hello world",
		},
		{
			name:         "multiple lines accumulated",
			input:        "Hello\nworld\nfrom\nstdin\n",
			expectedText: "Hello world from stdin",
		},
		{
			name:         "empty lines ignored",
			input:        "Hello\n\n\nworld\n",
			expectedText: "Hello world",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset received commands
			daemon.receivedCmds = make([]protocol.Command, 0)

			// Run uinput-client stream command
			cmd := exec.Command(clientBin, "stream", "--socket", daemon.addr())
			cmd.Stdin = bytes.NewBufferString(tt.input)

			var stdout, stderr bytes.Buffer
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr

			if err := cmd.Run(); err != nil {
				t.Fatalf("Command failed: %v\nStdout: %s\nStderr: %s", err, stdout.String(), stderr.String())
			}

			// Verify command was received
			time.Sleep(10 * time.Millisecond) // Give server time to process
			receivedCmd := daemon.getLastCommand()
			if receivedCmd == nil {
				t.Fatal("No command received by daemon")
			}

			if receivedCmd.Type != protocol.CommandType_Stream {
				t.Errorf("Expected command type %v, got %v", protocol.CommandType_Stream, receivedCmd.Type)
			}

			// Unmarshal payload
			var payload protocol.StreamPayload
			if err := json.Unmarshal(receivedCmd.Payload, &payload); err != nil {
				t.Fatalf("Failed to unmarshal payload: %v", err)
			}

			if payload.Text != tt.expectedText {
				t.Errorf("Expected text %q, got %q", tt.expectedText, payload.Text)
			}
		})
	}
}

func TestStreamCommand_LayoutFlagIntegration(t *testing.T) {
	daemon := newMockDaemon(t)
	defer daemon.close()

	clientBin := getClientBinary(t)

	tests := []struct {
		name           string
		layout         string
		expectedLayout string
	}{
		{
			name:           "french layout",
			layout:         "fr",
			expectedLayout: "fr",
		},
		{
			name:           "german layout",
			layout:         "de",
			expectedLayout: "de",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset received commands
			daemon.receivedCmds = make([]protocol.Command, 0)

			// Run uinput-client stream command with layout flag
			cmd := exec.Command(clientBin, "stream", "--socket", daemon.addr(), "--layout", tt.layout)
			cmd.Stdin = bytes.NewBufferString("Test text\n")

			if err := cmd.Run(); err != nil {
				t.Fatalf("Command failed: %v", err)
			}

			// Verify command
			time.Sleep(10 * time.Millisecond)
			receivedCmd := daemon.getLastCommand()
			if receivedCmd == nil {
				t.Fatal("No command received by daemon")
			}

			// Unmarshal payload
			var payload protocol.StreamPayload
			if err := json.Unmarshal(receivedCmd.Payload, &payload); err != nil {
				t.Fatalf("Failed to unmarshal payload: %v", err)
			}

			if payload.Layout != tt.expectedLayout {
				t.Errorf("Expected layout %q, got %q", tt.expectedLayout, payload.Layout)
			}
		})
	}
}

func TestStreamCommand_DelayFlagsIntegration(t *testing.T) {
	daemon := newMockDaemon(t)
	defer daemon.close()

	clientBin := getClientBinary(t)

	tests := []struct {
		name              string
		charDelay         string
		wordDelay         string
		expectedCharDelay int
		expectedWordDelay int
	}{
		{
			name:              "both delays specified",
			charDelay:         "50",
			wordDelay:         "200",
			expectedCharDelay: 50,
			expectedWordDelay: 200,
		},
		{
			name:              "only char delay",
			charDelay:         "100",
			wordDelay:         "",
			expectedCharDelay: 100,
			expectedWordDelay: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset received commands
			daemon.receivedCmds = make([]protocol.Command, 0)

			// Build command arguments
			args := []string{"stream", "--socket", daemon.addr()}
			if tt.charDelay != "" {
				args = append(args, "--char-delay", tt.charDelay)
			}
			if tt.wordDelay != "" {
				args = append(args, "--word-delay", tt.wordDelay)
			}

			// Run uinput-client stream command
			cmd := exec.Command(clientBin, args...)
			cmd.Stdin = bytes.NewBufferString("Test text\n")

			if err := cmd.Run(); err != nil {
				t.Fatalf("Command failed: %v", err)
			}

			// Verify command
			time.Sleep(10 * time.Millisecond)
			receivedCmd := daemon.getLastCommand()
			if receivedCmd == nil {
				t.Fatal("No command received by daemon")
			}

			// Unmarshal payload
			var payload protocol.StreamPayload
			if err := json.Unmarshal(receivedCmd.Payload, &payload); err != nil {
				t.Fatalf("Failed to unmarshal payload: %v", err)
			}

			if payload.CharDelay != tt.expectedCharDelay {
				t.Errorf("Expected CharDelay %d, got %d", tt.expectedCharDelay, payload.CharDelay)
			}

			if payload.DelayMs != tt.expectedWordDelay {
				t.Errorf("Expected DelayMs %d, got %d", tt.expectedWordDelay, payload.DelayMs)
			}
		})
	}
}

func TestStreamCommand_SimulStreamingIntegration(t *testing.T) {
	daemon := newMockDaemon(t)
	defer daemon.close()

	clientBin := getClientBinary(t)

	// Simulate SimulStreaming output (already filtered through awk)
	input := `my fellow Americans
ask not
what your country
`
	expectedText := "my fellow Americans ask not what your country"

	// Reset received commands
	daemon.receivedCmds = make([]protocol.Command, 0)

	// Run command
	cmd := exec.Command(clientBin, "stream", "--socket", daemon.addr(), "--layout", "fr")
	cmd.Stdin = bytes.NewBufferString(input)

	if err := cmd.Run(); err != nil {
		t.Fatalf("Command failed: %v", err)
	}

	// Verify
	time.Sleep(10 * time.Millisecond)
	receivedCmd := daemon.getLastCommand()
	if receivedCmd == nil {
		t.Fatal("No command received by daemon")
	}

	var payload protocol.StreamPayload
	if err := json.Unmarshal(receivedCmd.Payload, &payload); err != nil {
		t.Fatalf("Failed to unmarshal payload: %v", err)
	}

	if payload.Text != expectedText {
		t.Errorf("Expected text %q, got %q", expectedText, payload.Text)
	}

	if payload.Layout != "fr" {
		t.Errorf("Expected layout 'fr', got %q", payload.Layout)
	}
}

func TestStreamCommand_EmptyInputIntegration(t *testing.T) {
	daemon := newMockDaemon(t)
	defer daemon.close()

	clientBin := getClientBinary(t)

	// Run with empty input
	cmd := exec.Command(clientBin, "stream", "--socket", daemon.addr())
	cmd.Stdin = bytes.NewBufferString("")

	if err := cmd.Run(); err != nil {
		t.Fatalf("Command failed: %v", err)
	}

	// Verify no command was sent
	time.Sleep(10 * time.Millisecond)
	if len(daemon.getCommands()) > 0 {
		t.Error("Expected no command for empty input, but command was sent")
	}
}

// Benchmark for end-to-end streaming performance
func BenchmarkStreamCommand_EndToEnd(b *testing.B) {
	daemon := newMockDaemon(&testing.T{})
	defer daemon.close()

	clientBin := getClientBinary(&testing.T{})

	// Prepare test data
	testData := bytes.Repeat([]byte("Test line\n"), 100)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		cmd := exec.Command(clientBin, "stream", "--socket", daemon.addr())
		cmd.Stdin = bytes.NewReader(testData)

		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		if err := cmd.Run(); err != nil {
			b.Fatalf("Command failed: %v\nStderr: %s", err, stderr.String())
		}
	}
}

// TestStreamCommand_LargeInput tests handling of large text inputs
func TestStreamCommand_LargeInput(t *testing.T) {
	daemon := newMockDaemon(t)
	defer daemon.close()

	clientBin := getClientBinary(t)

	// Generate large input (1000 lines)
	var input bytes.Buffer
	for i := 0; i < 1000; i++ {
		fmt.Fprintf(&input, "Line %d with some text\n", i)
	}

	// Run command
	cmd := exec.Command(clientBin, "stream", "--socket", daemon.addr())
	cmd.Stdin = &input

	if err := cmd.Run(); err != nil {
		t.Fatalf("Command failed: %v", err)
	}

	// Verify command was received
	time.Sleep(20 * time.Millisecond)
	receivedCmd := daemon.getLastCommand()
	if receivedCmd == nil {
		t.Fatal("No command received by daemon")
	}

	var payload protocol.StreamPayload
	if err := json.Unmarshal(receivedCmd.Payload, &payload); err != nil {
		t.Fatalf("Failed to unmarshal payload: %v", err)
	}

	// Verify text is accumulated correctly
	if len(payload.Text) == 0 {
		t.Error("Expected non-empty text for large input")
	}
}

// TestStreamCommand_PipeIntegration tests piping from another command
func TestStreamCommand_PipeIntegration(t *testing.T) {
	daemon := newMockDaemon(t)
	defer daemon.close()

	clientBin := getClientBinary(t)

	// Simulate piping from echo command
	echoCmd := exec.Command("echo", "Hello from pipe")
	streamCmd := exec.Command(clientBin, "stream", "--socket", daemon.addr())

	// Connect pipe
	pipe, err := echoCmd.StdoutPipe()
	if err != nil {
		t.Fatalf("Failed to create pipe: %v", err)
	}
	streamCmd.Stdin = pipe

	// Start both commands
	if err := echoCmd.Start(); err != nil {
		t.Fatalf("Failed to start echo: %v", err)
	}
	if err := streamCmd.Start(); err != nil {
		t.Fatalf("Failed to start stream: %v", err)
	}

	// Wait for completion
	if err := echoCmd.Wait(); err != nil {
		t.Fatalf("Echo failed: %v", err)
	}
	if err := streamCmd.Wait(); err != nil {
		t.Fatalf("Stream failed: %v", err)
	}

	// Verify
	time.Sleep(10 * time.Millisecond)
	receivedCmd := daemon.getLastCommand()
	if receivedCmd == nil {
		t.Fatal("No command received by daemon")
	}

	var payload protocol.StreamPayload
	if err := json.Unmarshal(receivedCmd.Payload, &payload); err != nil {
		t.Fatalf("Failed to unmarshal payload: %v", err)
	}

	expectedText := "Hello from pipe"
	if payload.Text != expectedText {
		t.Errorf("Expected text %q, got %q", expectedText, payload.Text)
	}
}
