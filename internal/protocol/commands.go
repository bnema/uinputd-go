package protocol

import "encoding/json"

// CommandType represents the type of command being sent.
type CommandType string

const (
	CommandType_Type   CommandType = "type"   // Type text in batch mode
	CommandType_Stream CommandType = "stream" // Stream text in real-time
	CommandType_Key    CommandType = "key"    // Send a single key press
	CommandType_Ping   CommandType = "ping"   // Health check
)

// Command is the top-level message sent from client to daemon.
type Command struct {
	Type    CommandType     `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

// TypePayload is the payload for the "type" command (batch typing).
type TypePayload struct {
	Text   string `json:"text"`
	Layout string `json:"layout,omitempty"` // Optional, falls back to config default
}

// StreamPayload is the payload for the "stream" command (real-time typing).
type StreamPayload struct {
	Text      string `json:"text"`
	Layout    string `json:"layout,omitempty"`
	DelayMs   int    `json:"delay_ms,omitempty"`   // Delay between words
	CharDelay int    `json:"char_delay,omitempty"` // Delay between chars
}

// KeyPayload is the payload for the "key" command (single keypress).
type KeyPayload struct {
	Keycode  uint16 `json:"keycode"`
	Modifier string `json:"modifier,omitempty"` // "shift", "ctrl", "alt", "altgr"
}

// PingPayload is empty for ping command.
type PingPayload struct{}
