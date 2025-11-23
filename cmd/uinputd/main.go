package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/bnema/uinputd-go/internal/config"
	"github.com/bnema/uinputd-go/internal/logger"
	"github.com/bnema/uinputd-go/internal/server"
	"github.com/bnema/uinputd-go/internal/uinput"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
)

var (
	version   = "dev"
	commit    = "unknown"
	buildTime = "unknown"

	configPath string
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use:   "uinputd",
	Short: "uinputd - Input automation daemon with multi-layout support",
	Long: `uinputd is a daemon that creates a virtual keyboard device
and listens for input automation commands via Unix socket.

Features:
  - Multi-layout support (US, FR, DE, ES, UK, IT)
  - Real-time streaming input
  - Low resource usage
  - Group-based socket permissions (root:input)`,
	RunE: runDaemon,
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("uinputd %s\n", version)
		fmt.Printf("  commit:     %s\n", commit)
		fmt.Printf("  build time: %s\n", buildTime)
	},
}

func init() {
	rootCmd.Flags().StringVarP(&configPath, "config", "c", "", "config file path")
	rootCmd.Version = version
	rootCmd.AddCommand(versionCmd)
}

func runDaemon(cmd *cobra.Command, args []string) error {
	// Load configuration
	cfg, err := config.Load(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Setup logger with TTY auto-detection
	logLevel := config.ParseLogLevel(cfg.Logging.Level)
	baseLogger := logger.Setup(logLevel)

	// Create root context with signal handling
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// Add logger to context
	ctx = logger.WithLogger(ctx, baseLogger)
	log := logger.LogFromCtx(ctx)

	log.Info("uinputd starting", "version", version)

	// Check if running as root
	if os.Geteuid() != 0 {
		log.Fatal("uinputd must run as root to access /dev/uinput")
	}

	// Create virtual keyboard device
	device, err := uinput.New(ctx)
	if err != nil {
		log.Fatal("failed to create uinput device", "error", err)
	}
	defer device.Close()

	// Create server
	srv, err := server.New(ctx, cfg, device)
	if err != nil {
		log.Fatal("failed to create server", "error", err)
	}
	defer srv.Close()

	// Run server with errgroup for coordinated shutdown
	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		return srv.Start(ctx)
	})

	// Wait for completion or error
	if err := g.Wait(); err != nil {
		log.Error("server error", "error", err)
		return err
	}

	log.Info("uinputd shutdown complete")
	return nil
}
