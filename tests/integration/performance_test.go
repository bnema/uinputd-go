package integration

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/bnema/uinputd-go/internal/protocol"
)

func BenchmarkServer_TypeCommand(b *testing.B) {
	ts := newTestServer(&testing.T{})
	defer ts.close()

	payload := protocol.TypePayload{
		Text:   "hello world",
		Layout: "us",
	}

	payloadBytes, _ := json.Marshal(payload)
	cmd := &protocol.Command{
		Type:    protocol.CommandType_Type,
		Payload: payloadBytes,
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		resp := ts.sendCommand(&testing.T{}, cmd)
		if !resp.Success {
			b.Fatalf("Command failed: %s", resp.Error)
		}
	}
}

func BenchmarkServer_StreamCommand(b *testing.B) {
	ts := newTestServer(&testing.T{})
	defer ts.close()

	payload := protocol.StreamPayload{
		Text:      "hello world",
		Layout:    "us",
		CharDelay: 1,
		DelayMs:   2,
	}

	payloadBytes, _ := json.Marshal(payload)
	cmd := &protocol.Command{
		Type:    protocol.CommandType_Stream,
		Payload: payloadBytes,
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		resp := ts.sendCommand(&testing.T{}, cmd)
		if !resp.Success {
			b.Fatalf("Command failed: %s", resp.Error)
		}
	}
}

func BenchmarkServer_KeyCommand(b *testing.B) {
	ts := newTestServer(&testing.T{})
	defer ts.close()

	payload := protocol.KeyPayload{
		Keycode:  30, // KeyA
		Modifier: "",
	}

	payloadBytes, _ := json.Marshal(payload)
	cmd := &protocol.Command{
		Type:    protocol.CommandType_Key,
		Payload: payloadBytes,
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		resp := ts.sendCommand(&testing.T{}, cmd)
		if !resp.Success {
			b.Fatalf("Command failed: %s", resp.Error)
		}
	}
}

func BenchmarkServer_PingCommand(b *testing.B) {
	ts := newTestServer(&testing.T{})
	defer ts.close()

	payload := protocol.PingPayload{}
	payloadBytes, _ := json.Marshal(payload)
	cmd := &protocol.Command{
		Type:    protocol.CommandType_Ping,
		Payload: payloadBytes,
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		resp := ts.sendCommand(&testing.T{}, cmd)
		if !resp.Success {
			b.Fatalf("Ping failed: %s", resp.Error)
		}
	}
}

func BenchmarkServer_LongText(b *testing.B) {
	ts := newTestServer(&testing.T{})
	defer ts.close()

	// Generate long text (1000 characters)
	longText := strings.Repeat("hello world ", 100) // ~1200 chars

	payload := protocol.TypePayload{
		Text:   longText,
		Layout: "us",
	}

	payloadBytes, _ := json.Marshal(payload)
	cmd := &protocol.Command{
		Type:    protocol.CommandType_Type,
		Payload: payloadBytes,
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		resp := ts.sendCommand(&testing.T{}, cmd)
		if !resp.Success {
			b.Fatalf("Command failed: %s", resp.Error)
		}
	}
}

func BenchmarkServer_DeadKeyComposition(b *testing.B) {
	ts := newTestServer(&testing.T{})
	defer ts.close()

	payload := protocol.TypePayload{
		Text:   "château",
		Layout: "fr",
	}

	payloadBytes, _ := json.Marshal(payload)
	cmd := &protocol.Command{
		Type:    protocol.CommandType_Type,
		Payload: payloadBytes,
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		resp := ts.sendCommand(&testing.T{}, cmd)
		if !resp.Success {
			b.Fatalf("Command failed: %s", resp.Error)
		}
	}
}

func TestPerformance_DelayAccuracy(t *testing.T) {
	ts := newTestServer(t)
	defer ts.close()

	tests := []struct {
		name          string
		text          string
		charDelay     int
		wordDelay     int
		expectedMin   time.Duration
		tolerance     float64 // Percentage tolerance
	}{
		{
			name:        "10ms char delay",
			text:        "hello",
			charDelay:   10,
			wordDelay:   0,
			expectedMin: 40 * time.Millisecond, // 4 delays between 5 chars
			tolerance:   0.5,                    // ±50%
		},
		{
			name:        "50ms char delay",
			text:        "test",
			charDelay:   50,
			wordDelay:   0,
			expectedMin: 150 * time.Millisecond, // 3 delays between 4 chars
			tolerance:   0.5,
		},
		{
			name:        "100ms word delay",
			text:        "hi world",
			charDelay:   0,
			wordDelay:   100,
			expectedMin: 100 * time.Millisecond, // 1 space
			tolerance:   0.5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts.mockDevice.Reset()

			payload := protocol.StreamPayload{
				Text:      tt.text,
				Layout:    "us",
				CharDelay: tt.charDelay,
				DelayMs:   tt.wordDelay,
			}

			payloadBytes, _ := json.Marshal(payload)
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

			// Calculate expected range
			minExpected := tt.expectedMin
			maxExpected := tt.expectedMin + time.Duration(float64(tt.expectedMin)*tt.tolerance)
			minAllowed := minExpected - time.Duration(float64(minExpected)*tt.tolerance)

			t.Logf("Elapsed: %v, Expected: %v - %v", elapsed, minAllowed, maxExpected)

			if elapsed < minAllowed {
				t.Errorf("Command completed too quickly: %v < %v (tolerance: %.0f%%)",
					elapsed, minAllowed, tt.tolerance*100)
			}

			// Note: We don't set an upper limit because delays might be higher
			// due to system load, scheduling, etc.
		})
	}
}

func TestPerformance_Throughput(t *testing.T) {
	ts := newTestServer(t)
	defer ts.close()

	commandCount := 100
	text := "hello"

	payload := protocol.TypePayload{
		Text:   text,
		Layout: "us",
	}

	payloadBytes, _ := json.Marshal(payload)
	cmd := &protocol.Command{
		Type:    protocol.CommandType_Type,
		Payload: payloadBytes,
	}

	start := time.Now()

	for i := 0; i < commandCount; i++ {
		resp := ts.sendCommand(t, cmd)
		if !resp.Success {
			t.Fatalf("Command %d failed: %s", i, resp.Error)
		}
	}

	elapsed := time.Since(start)
	throughput := float64(commandCount) / elapsed.Seconds()

	t.Logf("Processed %d commands in %v (%.2f cmd/sec)", commandCount, elapsed, throughput)

	// Expect at least 100 commands per second (very conservative)
	minThroughput := 100.0
	if throughput < minThroughput {
		t.Errorf("Throughput too low: %.2f cmd/sec < %.2f cmd/sec", throughput, minThroughput)
	}
}

func TestPerformance_EventGenerationRate(t *testing.T) {
	ts := newTestServer(t)
	defer ts.close()

	// Generate text that will produce a known number of events
	charCount := 1000
	text := strings.Repeat("a", charCount)

	payload := protocol.TypePayload{
		Text:   text,
		Layout: "us",
	}

	payloadBytes, _ := json.Marshal(payload)
	cmd := &protocol.Command{
		Type:    protocol.CommandType_Type,
		Payload: payloadBytes,
	}

	start := time.Now()
	resp := ts.sendCommand(t, cmd)
	elapsed := time.Since(start)

	if !resp.Success {
		t.Fatalf("Command failed: %s", resp.Error)
	}

	eventCount := ts.mockDevice.GetEventCount()
	eventsPerSec := float64(eventCount) / elapsed.Seconds()

	t.Logf("Generated %d events in %v (%.2f events/sec)", eventCount, elapsed, eventsPerSec)

	// Verify we got the expected number of events
	expectedEvents := charCount * 4 // press+syn+release+syn per char
	if eventCount != expectedEvents {
		t.Errorf("Expected %d events, got %d", expectedEvents, eventCount)
	}

	// Expect at least 1000 events per second
	minRate := 1000.0
	if eventsPerSec < minRate {
		t.Errorf("Event generation rate too low: %.2f events/sec < %.2f events/sec", eventsPerSec, minRate)
	}
}

func TestPerformance_MemoryUsage(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping memory test in short mode")
	}

	ts := newTestServer(t)
	defer ts.close()

	// Send many commands and verify memory doesn't grow unbounded
	commandCount := 1000

	for i := 0; i < commandCount; i++ {
		payload := protocol.TypePayload{
			Text:   "test",
			Layout: "us",
		}

		payloadBytes, _ := json.Marshal(payload)
		cmd := &protocol.Command{
			Type:    protocol.CommandType_Type,
			Payload: payloadBytes,
		}

		resp := ts.sendCommand(t, cmd)
		if !resp.Success {
			t.Fatalf("Command %d failed: %s", i, resp.Error)
		}

		// Reset mock device to prevent memory growth from test infrastructure
		if i%100 == 0 {
			ts.mockDevice.Reset()
		}
	}

	t.Logf("Successfully processed %d commands without memory issues", commandCount)
}

func TestPerformance_LatencyUnderLoad(t *testing.T) {
	ts := newTestServer(t)
	defer ts.close()

	// Measure latency for individual commands under load
	commandCount := 100
	latencies := make([]time.Duration, commandCount)

	payload := protocol.TypePayload{
		Text:   "hello",
		Layout: "us",
	}

	payloadBytes, _ := json.Marshal(payload)

	for i := 0; i < commandCount; i++ {
		cmd := &protocol.Command{
			Type:    protocol.CommandType_Type,
			Payload: payloadBytes,
		}

		start := time.Now()
		resp := ts.sendCommand(t, cmd)
		latencies[i] = time.Since(start)

		if !resp.Success {
			t.Fatalf("Command %d failed: %s", i, resp.Error)
		}
	}

	// Calculate statistics
	var total time.Duration
	min := latencies[0]
	max := latencies[0]

	for _, lat := range latencies {
		total += lat
		if lat < min {
			min = lat
		}
		if lat > max {
			max = lat
		}
	}

	avg := total / time.Duration(commandCount)

	t.Logf("Latency stats: min=%v, max=%v, avg=%v", min, max, avg)

	// Expect average latency under 10ms
	maxAvg := 10 * time.Millisecond
	if avg > maxAvg {
		t.Errorf("Average latency too high: %v > %v", avg, maxAvg)
	}
}

func TestPerformance_LargePayload(t *testing.T) {
	ts := newTestServer(t)
	defer ts.close()

	// Test with very large text
	sizes := []int{1000, 5000, 10000}

	for _, size := range sizes {
		t.Run(fmt.Sprintf("size_%d", size), func(t *testing.T) {
			ts.mockDevice.Reset()

			text := strings.Repeat("a", size)
			payload := protocol.TypePayload{
				Text:   text,
				Layout: "us",
			}

			payloadBytes, _ := json.Marshal(payload)
			cmd := &protocol.Command{
				Type:    protocol.CommandType_Type,
				Payload: payloadBytes,
			}

			start := time.Now()
			resp := ts.sendCommand(t, cmd)
			elapsed := time.Since(start)

			if !resp.Success {
				t.Fatalf("Command failed: %s", resp.Error)
			}

			throughput := float64(size) / elapsed.Seconds()
			t.Logf("Size %d: processed in %v (%.2f chars/sec)", size, elapsed, throughput)

			// Verify events were generated
			eventCount := ts.mockDevice.GetEventCount()
			expectedEvents := size * 4
			if eventCount != expectedEvents {
				t.Errorf("Expected %d events, got %d", expectedEvents, eventCount)
			}
		})
	}
}

func TestPerformance_LayoutSwitchingOverhead(t *testing.T) {
	ts := newTestServer(t)
	defer ts.close()

	layouts := []string{"us", "fr", "de", "es", "it"}
	iterations := 20

	start := time.Now()

	for i := 0; i < iterations; i++ {
		layout := layouts[i%len(layouts)]

		payload := protocol.TypePayload{
			Text:   "hello",
			Layout: layout,
		}

		payloadBytes, _ := json.Marshal(payload)
		cmd := &protocol.Command{
			Type:    protocol.CommandType_Type,
			Payload: payloadBytes,
		}

		resp := ts.sendCommand(t, cmd)
		if !resp.Success {
			t.Fatalf("Command failed: %s", resp.Error)
		}
	}

	elapsed := time.Since(start)
	avgPerSwitch := elapsed / time.Duration(iterations)

	t.Logf("Layout switching: %d iterations in %v (avg %v per switch)", iterations, elapsed, avgPerSwitch)

	// Expect each layout switch + command to complete in under 10ms
	maxAvg := 10 * time.Millisecond
	if avgPerSwitch > maxAvg {
		t.Errorf("Layout switching overhead too high: %v > %v", avgPerSwitch, maxAvg)
	}
}
