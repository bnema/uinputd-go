package client_test

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/bnema/uinputd-go/pkg/client"
)

// Example demonstrates basic usage of the uinputd client SDK.
func Example() {
	// Create a client with default socket path
	c, err := client.NewDefault()
	if err != nil {
		log.Fatal(err)
	}
	defer c.Close()

	ctx := context.Background()

	// Type some text
	if err := c.TypeText(ctx, "Hello, World!", nil); err != nil {
		log.Fatal(err)
	}
}

// Example_withLayout demonstrates typing with a specific keyboard layout.
func Example_withLayout() {
	c, err := client.NewDefault()
	if err != nil {
		log.Fatal(err)
	}
	defer c.Close()

	ctx := context.Background()

	// Type French text using AZERTY layout
	err = c.TypeText(ctx, "Bonjour le monde!", &client.TypeOptions{
		Layout: "fr",
	})
	if err != nil {
		log.Fatal(err)
	}
}

// Example_streaming demonstrates real-time text streaming with delays.
func Example_streaming() {
	c, err := client.NewDefault()
	if err != nil {
		log.Fatal(err)
	}
	defer c.Close()

	ctx := context.Background()

	// Stream text with delays (useful for voice-to-text integration)
	err = c.StreamText(ctx, "This text appears word by word", &client.StreamOptions{
		Layout:    "us",
		DelayMs:   50, // 50ms delay between words
		CharDelay: 10, // 10ms delay between characters
	})
	if err != nil {
		log.Fatal(err)
	}
}

// Example_sendKey demonstrates sending individual key presses.
func Example_sendKey() {
	c, err := client.NewDefault()
	if err != nil {
		log.Fatal(err)
	}
	defer c.Close()

	ctx := context.Background()

	// Send Enter key (keycode 28)
	if err := c.SendKey(ctx, 28, client.ModifierNone); err != nil {
		log.Fatal(err)
	}

	// Send Ctrl+C (keycode 46 for 'C')
	if err := c.SendKey(ctx, 46, client.ModifierCtrl); err != nil {
		log.Fatal(err)
	}
}

// Example_withContext demonstrates using context for timeout and cancellation.
func Example_withContext() {
	c, err := client.NewDefault()
	if err != nil {
		log.Fatal(err)
	}
	defer c.Close()

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Operation will be cancelled if it takes longer than 2 seconds
	if err := c.TypeText(ctx, "Hello!", nil); err != nil {
		log.Fatal(err)
	}
}

// Example_customSocketPath demonstrates using a custom socket path.
func Example_customSocketPath() {
	// Connect to daemon at custom socket path
	c, err := client.New("/run/user/1000/uinputd.sock", &client.Options{
		Timeout: 10 * time.Second,
	})
	if err != nil {
		log.Fatal(err)
	}
	defer c.Close()

	ctx := context.Background()

	if err := c.TypeText(ctx, "Custom path!", nil); err != nil {
		log.Fatal(err)
	}
}

// Example_healthCheck demonstrates checking if the daemon is running.
func Example_healthCheck() {
	c, err := client.NewDefault()
	if err != nil {
		log.Fatal(err)
	}
	defer c.Close()

	ctx := context.Background()

	// Check if daemon is responsive
	if err := c.Ping(ctx); err != nil {
		log.Fatal("Daemon not responding:", err)
	}

	fmt.Println("Daemon is healthy")
	// Output: Daemon is healthy
}

// Example_integration demonstrates integration with a voice-to-text system.
func Example_integration() {
	c, err := client.NewDefault()
	if err != nil {
		log.Fatal(err)
	}
	defer c.Close()

	ctx := context.Background()

	// Simulate voice transcription arriving in chunks
	transcriptionChannel := make(chan string)

	// Goroutine simulating real-time transcription
	go func() {
		words := []string{"Hello", "this", "is", "a", "test"}
		for _, word := range words {
			transcriptionChannel <- word + " "
			time.Sleep(100 * time.Millisecond)
		}
		close(transcriptionChannel)
	}()

	// Stream each word as it arrives
	for word := range transcriptionChannel {
		if err := c.TypeText(ctx, word, nil); err != nil {
			log.Fatal(err)
		}
	}
}
