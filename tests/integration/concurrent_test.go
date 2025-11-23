package integration

import (
	"encoding/json"
	"sync"
	"testing"
	"time"

	"github.com/bnema/uinputd-go/internal/protocol"
)

func TestConcurrent_MultipleClients(t *testing.T) {
	ts := newTestServer(t)
	defer ts.close()

	clientCount := 10
	var wg sync.WaitGroup
	wg.Add(clientCount)

	errors := make([]error, clientCount)

	// Launch multiple clients simultaneously
	for i := 0; i < clientCount; i++ {
		go func(clientID int) {
			defer wg.Done()

			payload := protocol.TypePayload{
				Text:   "hello",
				Layout: "us",
			}

			payloadBytes, err := json.Marshal(payload)
			if err != nil {
				errors[clientID] = err
				return
			}

			cmd := &protocol.Command{
				Type:    protocol.CommandType_Type,
				Payload: payloadBytes,
			}

			resp := ts.sendCommand(t, cmd)

			if !resp.Success {
				t.Errorf("Client %d: command failed: %s", clientID, resp.Error)
			}
		}(i)
	}

	wg.Wait()

	// Check for errors
	for i, err := range errors {
		if err != nil {
			t.Errorf("Client %d: error: %v", i, err)
		}
	}

	// Verify events were generated (should be 10 clients * 5 chars * 4 events)
	eventCount := ts.mockDevice.GetEventCount()
	expectedMin := clientCount * 5 * 4
	if eventCount < expectedMin {
		t.Errorf("Expected at least %d events from %d clients, got %d", expectedMin, clientCount, eventCount)
	}
}

func TestConcurrent_StreamCommands(t *testing.T) {
	ts := newTestServer(t)
	defer ts.close()

	clientCount := 5
	var wg sync.WaitGroup
	wg.Add(clientCount)

	// Each client sends a stream command with delays
	for i := 0; i < clientCount; i++ {
		go func(clientID int) {
			defer wg.Done()

			payload := protocol.StreamPayload{
				Text:      "test",
				Layout:    "us",
				CharDelay: 5,
				DelayMs:   10,
			}

			payloadBytes, err := json.Marshal(payload)
			if err != nil {
				t.Errorf("Client %d: marshal error: %v", clientID, err)
				return
			}

			cmd := &protocol.Command{
				Type:    protocol.CommandType_Stream,
				Payload: payloadBytes,
			}

			resp := ts.sendCommand(t, cmd)

			if !resp.Success {
				t.Errorf("Client %d: command failed: %s", clientID, resp.Error)
			}
		}(i)
	}

	wg.Wait()

	// Verify all events were generated
	if ts.mockDevice.GetEventCount() == 0 {
		t.Error("No events generated from concurrent stream commands")
	}
}

func TestConcurrent_MixedCommands(t *testing.T) {
	ts := newTestServer(t)
	defer ts.close()

	commandCount := 20
	var wg sync.WaitGroup
	wg.Add(commandCount)

	// Send a mix of type, stream, key, and ping commands
	for i := 0; i < commandCount; i++ {
		go func(cmdID int) {
			defer wg.Done()

			var cmd *protocol.Command

			switch cmdID % 4 {
			case 0: // Type command
				payload := protocol.TypePayload{
					Text:   "hello",
					Layout: "us",
				}
				payloadBytes, _ := json.Marshal(payload)
				cmd = &protocol.Command{
					Type:    protocol.CommandType_Type,
					Payload: payloadBytes,
				}

			case 1: // Stream command
				payload := protocol.StreamPayload{
					Text:      "world",
					Layout:    "us",
					CharDelay: 2,
				}
				payloadBytes, _ := json.Marshal(payload)
				cmd = &protocol.Command{
					Type:    protocol.CommandType_Stream,
					Payload: payloadBytes,
				}

			case 2: // Key command
				payload := protocol.KeyPayload{
					Keycode:  30, // KeyA
					Modifier: "",
				}
				payloadBytes, _ := json.Marshal(payload)
				cmd = &protocol.Command{
					Type:    protocol.CommandType_Key,
					Payload: payloadBytes,
				}

			case 3: // Ping command
				payload := protocol.PingPayload{}
				payloadBytes, _ := json.Marshal(payload)
				cmd = &protocol.Command{
					Type:    protocol.CommandType_Ping,
					Payload: payloadBytes,
				}
			}

			resp := ts.sendCommand(t, cmd)

			if !resp.Success {
				t.Errorf("Command %d (type=%s) failed: %s", cmdID, cmd.Type, resp.Error)
			}
		}(i)
	}

	wg.Wait()

	// Verify events were generated (excluding ping commands)
	if ts.mockDevice.GetEventCount() == 0 {
		t.Error("No events generated from mixed concurrent commands")
	}
}

func TestConcurrent_DifferentLayouts(t *testing.T) {
	ts := newTestServer(t)
	defer ts.close()

	layouts := []string{"us", "fr", "de", "es", "it"}
	var wg sync.WaitGroup
	wg.Add(len(layouts))

	// Each goroutine uses a different layout
	for _, layout := range layouts {
		go func(l string) {
			defer wg.Done()

			payload := protocol.TypePayload{
				Text:   "hello",
				Layout: l,
			}

			payloadBytes, err := json.Marshal(payload)
			if err != nil {
				t.Errorf("Layout %s: marshal error: %v", l, err)
				return
			}

			cmd := &protocol.Command{
				Type:    protocol.CommandType_Type,
				Payload: payloadBytes,
			}

			resp := ts.sendCommand(t, cmd)

			if !resp.Success {
				t.Errorf("Layout %s: command failed: %s", l, resp.Error)
			}
		}(layout)
	}

	wg.Wait()

	// Verify events for all layouts
	eventCount := ts.mockDevice.GetEventCount()
	expectedMin := len(layouts) * 5 * 4 // 5 layouts * 5 chars * 4 events
	if eventCount < expectedMin {
		t.Errorf("Expected at least %d events from %d layouts, got %d", expectedMin, len(layouts), eventCount)
	}
}

func TestConcurrent_RapidFireCommands(t *testing.T) {
	ts := newTestServer(t)
	defer ts.close()

	commandCount := 50
	var wg sync.WaitGroup
	wg.Add(commandCount)

	// Send many commands as fast as possible
	start := time.Now()
	for i := 0; i < commandCount; i++ {
		go func(cmdID int) {
			defer wg.Done()

			payload := protocol.TypePayload{
				Text:   "a",
				Layout: "us",
			}

			payloadBytes, _ := json.Marshal(payload)
			cmd := &protocol.Command{
				Type:    protocol.CommandType_Type,
				Payload: payloadBytes,
			}

			resp := ts.sendCommand(t, cmd)

			if !resp.Success {
				t.Errorf("Command %d failed: %s", cmdID, resp.Error)
			}
		}(i)
	}

	wg.Wait()
	elapsed := time.Since(start)

	t.Logf("Processed %d commands in %v (%.2f cmd/sec)", commandCount, elapsed, float64(commandCount)/elapsed.Seconds())

	// Verify all events were generated
	eventCount := ts.mockDevice.GetEventCount()
	expectedMin := commandCount * 4 // 1 char * 4 events per command
	if eventCount < expectedMin {
		t.Errorf("Expected at least %d events, got %d", expectedMin, eventCount)
	}
}

func TestConcurrent_LongRunningStreams(t *testing.T) {
	ts := newTestServer(t)
	defer ts.close()

	streamCount := 3
	var wg sync.WaitGroup
	wg.Add(streamCount)

	// Launch several long-running stream commands
	for i := 0; i < streamCount; i++ {
		go func(streamID int) {
			defer wg.Done()

			// Each stream sends a longer text with delays
			payload := protocol.StreamPayload{
				Text:      "this is a longer text message",
				Layout:    "us",
				CharDelay: 3,
				DelayMs:   5,
			}

			payloadBytes, _ := json.Marshal(payload)
			cmd := &protocol.Command{
				Type:    protocol.CommandType_Stream,
				Payload: payloadBytes,
			}

			resp := ts.sendCommand(t, cmd)

			if !resp.Success {
				t.Errorf("Stream %d failed: %s", streamID, resp.Error)
			}
		}(i)
	}

	wg.Wait()

	// Verify events were generated for all streams
	if ts.mockDevice.GetEventCount() == 0 {
		t.Error("No events generated from concurrent long streams")
	}
}

func TestConcurrent_StressTest(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	ts := newTestServer(t)
	defer ts.close()

	totalCommands := 100
	var wg sync.WaitGroup
	wg.Add(totalCommands)

	start := time.Now()
	successCount := 0
	var mu sync.Mutex

	for i := 0; i < totalCommands; i++ {
		go func(cmdID int) {
			defer wg.Done()

			// Randomize command types and payloads
			var cmd *protocol.Command
			if cmdID%2 == 0 {
				payload := protocol.TypePayload{
					Text:   "stress test",
					Layout: "us",
				}
				payloadBytes, _ := json.Marshal(payload)
				cmd = &protocol.Command{
					Type:    protocol.CommandType_Type,
					Payload: payloadBytes,
				}
			} else {
				payload := protocol.StreamPayload{
					Text:      "test",
					Layout:    "us",
					CharDelay: 1,
				}
				payloadBytes, _ := json.Marshal(payload)
				cmd = &protocol.Command{
					Type:    protocol.CommandType_Stream,
					Payload: payloadBytes,
				}
			}

			resp := ts.sendCommand(t, cmd)

			if resp.Success {
				mu.Lock()
				successCount++
				mu.Unlock()
			} else {
				t.Errorf("Command %d failed: %s", cmdID, resp.Error)
			}
		}(i)
	}

	wg.Wait()
	elapsed := time.Since(start)

	t.Logf("Stress test: %d/%d commands succeeded in %v", successCount, totalCommands, elapsed)

	if successCount < totalCommands {
		t.Errorf("Some commands failed: %d/%d succeeded", successCount, totalCommands)
	}

	// Verify events were generated
	if ts.mockDevice.GetEventCount() == 0 {
		t.Error("No events generated during stress test")
	}
}

func TestConcurrent_NoEventInterleaving(t *testing.T) {
	ts := newTestServer(t)
	defer ts.close()

	// This test verifies that events for different keys don't interleave improperly.
	// Each client sends a unique single character, and we verify all events are complete.

	clientCount := 10
	var wg sync.WaitGroup
	wg.Add(clientCount)

	for i := 0; i < clientCount; i++ {
		go func(clientID int) {
			defer wg.Done()

			// Each client sends a unique character
			char := string(rune('a' + clientID))
			payload := protocol.TypePayload{
				Text:   char,
				Layout: "us",
			}

			payloadBytes, _ := json.Marshal(payload)
			cmd := &protocol.Command{
				Type:    protocol.CommandType_Type,
				Payload: payloadBytes,
			}

			resp := ts.sendCommand(t, cmd)

			if !resp.Success {
				t.Errorf("Client %d: command failed: %s", clientID, resp.Error)
			}
		}(i)
	}

	wg.Wait()

	// Verify all events are properly formed (every key press has press+syn+release+syn)
	events := ts.mockDevice.GetEvents()

	// Count key presses and releases
	presses := 0
	releases := 0
	for _, event := range events {
		if event.Type == 0x01 { // EvKey
			if event.Value == 1 { // KeyPress
				presses++
			} else if event.Value == 0 { // KeyRelease
				releases++
			}
		}
	}

	if presses != releases {
		t.Errorf("Event interleaving detected: %d presses but %d releases", presses, releases)
	}

	if presses != clientCount {
		t.Errorf("Expected %d key presses, got %d", clientCount, presses)
	}
}
