package server

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/bnema/uinputd-go/internal/layouts"
	"github.com/bnema/uinputd-go/internal/logger"
	"github.com/bnema/uinputd-go/internal/protocol"
	"github.com/bnema/uinputd-go/internal/uinput"
)

// handleCommand routes commands to appropriate handlers.
func (s *Server) handleCommand(ctx context.Context, cmd *protocol.Command) error {
	log := logger.LogFromCtx(ctx)
	log.Info("handling command", "type", cmd.Type)

	switch cmd.Type {
	case protocol.CommandType_Type:
		return s.handleType(ctx, cmd.Payload)
	case protocol.CommandType_Stream:
		return s.handleStream(ctx, cmd.Payload)
	case protocol.CommandType_Key:
		return s.handleKey(ctx, cmd.Payload)
	case protocol.CommandType_Ping:
		return s.handlePing(ctx)
	default:
		return fmt.Errorf("unknown command type: %s", cmd.Type)
	}
}

// handleType processes batch typing command.
func (s *Server) handleType(ctx context.Context, payload json.RawMessage) error {
	log := logger.LogFromCtx(ctx)

	var p protocol.TypePayload
	if err := json.Unmarshal(payload, &p); err != nil {
		return fmt.Errorf("invalid type payload: %w", err)
	}

	// Get layout (use config default if not specified)
	layoutName := p.Layout
	if layoutName == "" {
		layoutName = s.cfg.Layout
	}

	layout, err := s.registry.Get(layoutName)
	if err != nil {
		return fmt.Errorf("layout error: %w", err)
	}

	log.Info("typing text", "length", len(p.Text), "layout", layoutName)

	// Type each character
	for _, char := range p.Text {
		sequence, err := layout.CharToKeySequence(ctx, char)
		if err != nil {
			log.Warn("character not supported", "char", string(char), "error", err)
			continue // Skip unsupported characters
		}

		// Send each keystroke in the sequence
		// For simple characters, sequence has one element
		// For dead key combinations, sequence has multiple elements (e.g., circumflex + vowel)
		for _, key := range sequence {
			shift := (key.Modifier & layouts.ModShift) != 0
			altGr := (key.Modifier & layouts.ModAltGr) != 0

			if err := s.sendKeyWithModifiers(ctx, key.Keycode, shift, altGr); err != nil {
				return fmt.Errorf("failed to send key: %w", err)
			}
		}
	}

	return nil
}

// handleStream processes real-time streaming command with natural typing delays.
func (s *Server) handleStream(ctx context.Context, payload json.RawMessage) error {
	log := logger.LogFromCtx(ctx)

	var p protocol.StreamPayload
	if err := json.Unmarshal(payload, &p); err != nil {
		return fmt.Errorf("invalid stream payload: %w", err)
	}

	// Get layout (use config default if not specified)
	layoutName := p.Layout
	if layoutName == "" {
		layoutName = s.cfg.Layout
	}

	layout, err := s.registry.Get(layoutName)
	if err != nil {
		return fmt.Errorf("layout error: %w", err)
	}

	// Get delays (use config defaults if not specified)
	charDelay := time.Duration(p.CharDelay) * time.Millisecond
	if p.CharDelay == 0 {
		charDelay = time.Duration(s.cfg.Performance.CharDelayMs) * time.Millisecond
	}

	wordDelay := time.Duration(p.DelayMs) * time.Millisecond
	if p.DelayMs == 0 {
		wordDelay = time.Duration(s.cfg.Performance.StreamDelayMs) * time.Millisecond
	}

	log.Info("streaming text", "length", len(p.Text), "layout", layoutName, "char_delay_ms", charDelay.Milliseconds(), "word_delay_ms", wordDelay.Milliseconds())

	// Split text into words for word-level delays
	words := strings.Fields(p.Text)

	for i, word := range words {
		// Type each character in the word
		for _, char := range word {
			sequence, err := layout.CharToKeySequence(ctx, char)
			if err != nil {
				log.Warn("character not supported", "char", string(char), "error", err)
				continue // Skip unsupported characters
			}

			// Send each keystroke in the sequence
			for _, key := range sequence {
				shift := (key.Modifier & layouts.ModShift) != 0
				altGr := (key.Modifier & layouts.ModAltGr) != 0

				if err := s.sendKeyWithModifiers(ctx, key.Keycode, shift, altGr); err != nil {
					return fmt.Errorf("failed to send key: %w", err)
				}
			}

			// Delay between characters
			if charDelay > 0 {
				time.Sleep(charDelay)
			}
		}

		// Add space between words (except after last word)
		if i < len(words)-1 {
			// Type space character
			sequence, err := layout.CharToKeySequence(ctx, ' ')
			if err == nil {
				for _, key := range sequence {
					shift := (key.Modifier & layouts.ModShift) != 0
					altGr := (key.Modifier & layouts.ModAltGr) != 0

					if err := s.sendKeyWithModifiers(ctx, key.Keycode, shift, altGr); err != nil {
						return fmt.Errorf("failed to send space: %w", err)
					}
				}
			}

			// Delay between words
			if wordDelay > 0 {
				time.Sleep(wordDelay)
			}
		}
	}

	return nil
}

// handleKey processes single key press command.
func (s *Server) handleKey(ctx context.Context, payload json.RawMessage) error {
	log := logger.LogFromCtx(ctx)

	var p protocol.KeyPayload
	if err := json.Unmarshal(payload, &p); err != nil {
		return fmt.Errorf("invalid key payload: %w", err)
	}

	log.Info("sending key", "keycode", p.Keycode, "modifier", p.Modifier)

	// Parse modifier
	var modKeycode uint16
	switch p.Modifier {
	case "shift":
		modKeycode = uinput.KeyLeftShift
	case "ctrl":
		modKeycode = uinput.KeyLeftCtrl
	case "alt":
		modKeycode = uinput.KeyLeftAlt
	case "altgr":
		modKeycode = uinput.KeyRightAlt
	case "":
		// No modifier, send key directly
		return s.device.SendKey(ctx, p.Keycode)
	default:
		return fmt.Errorf("unknown modifier: %s", p.Modifier)
	}

	// Send key with modifier
	return s.device.SendKeyWithModifier(ctx, modKeycode, p.Keycode)
}

// handlePing responds to health check.
func (s *Server) handlePing(ctx context.Context) error {
	log := logger.LogFromCtx(ctx)
	log.Debug("ping received")
	return nil
}

// sendKeyWithModifiers sends a key press with shift and/or altgr modifiers.
func (s *Server) sendKeyWithModifiers(ctx context.Context, keycode uint16, shift, altGr bool) error {
	if !shift && !altGr {
		// No modifiers, simple key press
		return s.device.SendKey(ctx, keycode)
	}

	if shift && !altGr {
		// Shift only
		return s.device.SendKeyWithModifier(ctx, uinput.KeyLeftShift, keycode)
	}

	if altGr && !shift {
		// AltGr only
		return s.device.SendKeyWithModifier(ctx, uinput.KeyRightAlt, keycode)
	}

	// Both Shift + AltGr (e.g., for some special characters)
	return s.sendKeyWithBothModifiers(ctx, keycode)
}

// sendKeyWithBothModifiers sends a key with both Shift and AltGr pressed.
func (s *Server) sendKeyWithBothModifiers(ctx context.Context, keycode uint16) error {
	// Press Shift
	if err := s.device.WriteEvent(uinput.NewKeyEvent(uinput.KeyLeftShift, true)); err != nil {
		return err
	}
	if err := s.device.WriteEvent(uinput.NewSynEvent()); err != nil {
		return err
	}

	// Press AltGr
	if err := s.device.WriteEvent(uinput.NewKeyEvent(uinput.KeyRightAlt, true)); err != nil {
		return err
	}
	if err := s.device.WriteEvent(uinput.NewSynEvent()); err != nil {
		return err
	}

	// Press key
	if err := s.device.WriteEvent(uinput.NewKeyEvent(keycode, true)); err != nil {
		return err
	}
	if err := s.device.WriteEvent(uinput.NewSynEvent()); err != nil {
		return err
	}

	// Release key
	if err := s.device.WriteEvent(uinput.NewKeyEvent(keycode, false)); err != nil {
		return err
	}
	if err := s.device.WriteEvent(uinput.NewSynEvent()); err != nil {
		return err
	}

	// Release AltGr
	if err := s.device.WriteEvent(uinput.NewKeyEvent(uinput.KeyRightAlt, false)); err != nil {
		return err
	}
	if err := s.device.WriteEvent(uinput.NewSynEvent()); err != nil {
		return err
	}

	// Release Shift
	if err := s.device.WriteEvent(uinput.NewKeyEvent(uinput.KeyLeftShift, false)); err != nil {
		return err
	}
	if err := s.device.WriteEvent(uinput.NewSynEvent()); err != nil {
		return err
	}

	return nil
}
