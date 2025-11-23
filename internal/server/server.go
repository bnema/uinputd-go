package server

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"os/user"
	"strconv"

	"github.com/bnema/uinputd-go/internal/config"
	"github.com/bnema/uinputd-go/internal/layouts"
	"github.com/bnema/uinputd-go/internal/logger"
	"github.com/bnema/uinputd-go/internal/protocol"
	"github.com/bnema/uinputd-go/internal/uinput"
	"golang.org/x/sync/errgroup"
)

// Server manages the Unix socket server and handles client connections.
type Server struct {
	cfg      *config.Config
	device   uinput.DeviceInterface
	registry layouts.RegistryInterface
	listener net.Listener
}

// New creates a new server instance.
func New(ctx context.Context, cfg *config.Config, device uinput.DeviceInterface) (*Server, error) {
	log := logger.LogFromCtx(ctx)

	// Remove existing socket if it exists
	if err := os.RemoveAll(cfg.Socket.Path); err != nil {
		return nil, fmt.Errorf("failed to remove existing socket: %w", err)
	}

	// Create Unix socket listener
	listener, err := net.Listen("unix", cfg.Socket.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to create socket: %w", err)
	}

	// Set socket permissions (group-based: root:input 0660)
	if err := os.Chmod(cfg.Socket.Path, os.FileMode(cfg.Socket.Permissions)); err != nil {
		listener.Close()
		return nil, fmt.Errorf("failed to set socket permissions: %w", err)
	}

	// Try to set group ownership to 'input' (GID typically 104 or similar)
	// This allows users in the input group to connect
	if err := setSocketGroup(cfg.Socket.Path); err != nil {
		log.Warn("failed to set socket group ownership", "error", err, "hint", "run 'chgrp input "+cfg.Socket.Path+"' manually if needed")
	}

	log.Info("unix socket created", "path", cfg.Socket.Path, "permissions", fmt.Sprintf("%o", cfg.Socket.Permissions))

	return &Server{
		cfg:      cfg,
		device:   device,
		registry: layouts.NewRegistry(),
		listener: listener,
	}, nil
}

// Start begins accepting client connections.
// This blocks until ctx is cancelled or an error occurs.
func (s *Server) Start(ctx context.Context) error {
	log := logger.LogFromCtx(ctx)
	log.Info("server starting", "socket", s.cfg.Socket.Path)

	g, ctx := errgroup.WithContext(ctx)

	// Goroutine to accept connections
	g.Go(func() error {
		for {
			conn, err := s.listener.Accept()
			if err != nil {
				select {
				case <-ctx.Done():
					return nil // Graceful shutdown
				default:
					return fmt.Errorf("accept error: %w", err)
				}
			}

			// Handle connection in separate goroutine
			g.Go(func() error {
				return s.handleConnection(ctx, conn)
			})
		}
	})

	// Goroutine to handle shutdown signal
	g.Go(func() error {
		<-ctx.Done()
		log.Info("shutting down server")
		return s.listener.Close()
	})

	return g.Wait()
}

// handleConnection processes a single client connection.
func (s *Server) handleConnection(ctx context.Context, conn net.Conn) error {
	defer conn.Close()

	log := logger.LogFromCtx(ctx)
	log.Debug("client connected", "remote", conn.RemoteAddr())

	// Read command from client
	decoder := json.NewDecoder(conn)
	var cmd protocol.Command
	if err := decoder.Decode(&cmd); err != nil {
		if err == io.EOF {
			return nil // Client disconnected
		}
		return s.sendError(conn, fmt.Errorf("failed to decode command: %w", err))
	}

	// Enrich context with command info
	cmdLogger := log.With("cmd_type", cmd.Type)
	ctx = logger.WithLogger(ctx, cmdLogger)

	// Handle command
	if err := s.handleCommand(ctx, &cmd); err != nil {
		return s.sendError(conn, err)
	}

	// Send success response
	return s.sendSuccess(conn, "command executed successfully")
}

// sendSuccess sends a success response to the client.
func (s *Server) sendSuccess(conn net.Conn, message string) error {
	resp := protocol.NewSuccessResponse(message)
	return json.NewEncoder(conn).Encode(resp)
}

// sendError sends an error response to the client.
func (s *Server) sendError(conn net.Conn, err error) error {
	resp := protocol.NewErrorResponse(err)
	return json.NewEncoder(conn).Encode(resp)
}

// setSocketGroup attempts to set the socket's group to 'input'.
func setSocketGroup(path string) error {
	// Look up 'input' group using standard library
	group, err := user.LookupGroup("input")
	if err != nil {
		return fmt.Errorf("input group not found")
	}

	// Convert GID string to int
	gid, err := strconv.Atoi(group.Gid)
	if err != nil {
		return fmt.Errorf("invalid group ID: %w", err)
	}

	// Change group ownership: -1 means don't change UID
	if err := os.Chown(path, -1, gid); err != nil {
		return fmt.Errorf("failed to set group ownership: %w", err)
	}

	return nil
}

// Close cleanly shuts down the server.
func (s *Server) Close() error {
	if s.listener != nil {
		return s.listener.Close()
	}
	return nil
}
