package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/bnema/uinputd-go/internal/protocol"
)

// Client represents a connection to the uinputd daemon.
// It provides a high-level API for sending input automation commands.
//
// Example usage:
//
//	client, err := client.New("/tmp/.uinputd.sock")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer client.Close()
//
//	err = client.TypeText(ctx, "Hello, World!", nil)
type Client struct {
	socketPath string
	mu         sync.Mutex
	conn       net.Conn
	timeout    time.Duration
}

// Options contains optional configuration for the client.
type Options struct {
	// Timeout for socket operations (default: 5s)
	Timeout time.Duration
}

// TypeOptions contains options for typing text.
type TypeOptions struct {
	// Layout specifies the keyboard layout (us, fr, de, es, uk, it)
	// If empty, uses the daemon's default layout
	Layout string
}

// StreamOptions contains options for streaming text.
type StreamOptions struct {
	// Layout specifies the keyboard layout
	Layout string
	// DelayMs is the delay between words in milliseconds
	DelayMs int
	// CharDelay is the delay between characters in milliseconds
	CharDelay int
}

// KeyModifier represents keyboard modifiers
type KeyModifier string

const (
	ModifierNone  KeyModifier = ""
	ModifierShift KeyModifier = "shift"
	ModifierCtrl  KeyModifier = "ctrl"
	ModifierAlt   KeyModifier = "alt"
	ModifierAltGr KeyModifier = "altgr"
)

// New creates a new client connected to the uinputd daemon.
// The socketPath is typically "/tmp/.uinputd.sock" or the path from config.
func New(socketPath string, opts *Options) (*Client, error) {
	if opts == nil {
		opts = &Options{}
	}

	if opts.Timeout == 0 {
		opts.Timeout = 5 * time.Second
	}

	c := &Client{
		socketPath: socketPath,
		timeout:    opts.Timeout,
	}

	return c, nil
}

// NewDefault creates a client with the default socket path.
func NewDefault() (*Client, error) {
	return New("/tmp/.uinputd.sock", nil)
}

// connect establishes a connection to the daemon.
func (c *Client) connect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn != nil {
		return nil // Already connected
	}

	conn, err := net.DialTimeout("unix", c.socketPath, c.timeout)
	if err != nil {
		return fmt.Errorf("failed to connect to daemon at %s: %w (is uinputd running?)", c.socketPath, err)
	}

	c.conn = conn
	return nil
}

// disconnect closes the connection to the daemon.
func (c *Client) disconnect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn != nil {
		err := c.conn.Close()
		c.conn = nil
		return err
	}

	return nil
}

// sendCommand sends a command to the daemon and returns the response.
func (c *Client) sendCommand(ctx context.Context, cmdType protocol.CommandType, payload interface{}) error {
	// Connect if not already connected
	if err := c.connect(); err != nil {
		return err
	}

	// Set deadline based on context or timeout
	var deadline time.Time
	if d, ok := ctx.Deadline(); ok {
		deadline = d
	} else {
		deadline = time.Now().Add(c.timeout)
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if err := c.conn.SetDeadline(deadline); err != nil {
		return fmt.Errorf("failed to set deadline: %w", err)
	}

	// Marshal payload
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	// Create command
	cmd := protocol.Command{
		Type:    cmdType,
		Payload: payloadBytes,
	}

	// Send command
	if err := json.NewEncoder(c.conn).Encode(&cmd); err != nil {
		c.conn = nil // Connection broken, force reconnect next time
		return fmt.Errorf("failed to send command: %w", err)
	}

	// Read response
	var resp protocol.Response
	if err := json.NewDecoder(c.conn).Decode(&resp); err != nil {
		c.conn = nil // Connection broken
		return fmt.Errorf("failed to read response: %w", err)
	}

	// Check for errors
	if !resp.Success {
		return fmt.Errorf("daemon error: %s", resp.Error)
	}

	return nil
}

// TypeText types the given text using the specified layout.
// This is a batch operation - all text is sent at once.
//
// Example:
//
//	err := client.TypeText(ctx, "Hello, World!", &client.TypeOptions{
//	    Layout: "us",
//	})
func (c *Client) TypeText(ctx context.Context, text string, opts *TypeOptions) error {
	if opts == nil {
		opts = &TypeOptions{}
	}

	payload := protocol.TypePayload{
		Text:   text,
		Layout: opts.Layout,
	}

	return c.sendCommand(ctx, protocol.CommandType_Type, payload)
}

// StreamText streams text with configurable delays.
// This allows for more natural-looking typing with delays between words/characters.
//
// Example:
//
//	err := client.StreamText(ctx, "Hello World", &client.StreamOptions{
//	    Layout:    "fr",
//	    DelayMs:   50,  // 50ms between words
//	    CharDelay: 10,  // 10ms between characters
//	})
func (c *Client) StreamText(ctx context.Context, text string, opts *StreamOptions) error {
	if opts == nil {
		opts = &StreamOptions{}
	}

	payload := protocol.StreamPayload{
		Text:      text,
		Layout:    opts.Layout,
		DelayMs:   opts.DelayMs,
		CharDelay: opts.CharDelay,
	}

	return c.sendCommand(ctx, protocol.CommandType_Stream, payload)
}

// SendKey sends a single keypress with an optional modifier.
//
// Example:
//
//	// Send Enter key (keycode 28)
//	err := client.SendKey(ctx, 28, client.ModifierNone)
//
//	// Send Ctrl+C (keycode 46 for 'C')
//	err := client.SendKey(ctx, 46, client.ModifierCtrl)
func (c *Client) SendKey(ctx context.Context, keycode uint16, modifier KeyModifier) error {
	payload := protocol.KeyPayload{
		Keycode:  keycode,
		Modifier: string(modifier),
	}

	return c.sendCommand(ctx, protocol.CommandType_Key, payload)
}

// Ping checks if the daemon is responsive.
// Returns nil if the daemon responds successfully.
//
// Example:
//
//	if err := client.Ping(ctx); err != nil {
//	    log.Fatal("Daemon not responding:", err)
//	}
func (c *Client) Ping(ctx context.Context) error {
	return c.sendCommand(ctx, protocol.CommandType_Ping, protocol.PingPayload{})
}

// Close closes the connection to the daemon.
// Should be called when the client is no longer needed.
func (c *Client) Close() error {
	return c.disconnect()
}

// IsConnected returns true if the client is currently connected to the daemon.
func (c *Client) IsConnected() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.conn != nil
}
