package server

import (
	"context"
	"encoding/json"
	"fmt"

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
		keycode, shift, altGr, err := layout.CharToKeycode(ctx, char)
		if err != nil {
			log.Warn("character not supported", "char", string(char), "error", err)
			continue // Skip unsupported characters
		}

		// Send key event with modifiers
		if err := s.sendKeyWithModifiers(ctx, keycode, shift, altGr); err != nil {
			return fmt.Errorf("failed to send key: %w", err)
		}
	}

	return nil
}

// handleStream processes real-time streaming command.
func (s *Server) handleStream(ctx context.Context, payload json.RawMessage) error {
	// For now, handleStream is identical to handleType
	// In Phase 4, we'll add proper streaming with delays
	return s.handleType(ctx, payload)
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
