package main

import (
	"bufio"
	_ "embed"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/bnema/uinputd-go/internal/doctor"
	"github.com/bnema/uinputd-go/internal/installer"
	"github.com/bnema/uinputd-go/internal/protocol"
	"github.com/bnema/uinputd-go/internal/styles"
	"github.com/spf13/cobra"
)

//go:embed embedded/uinputd
var embeddedDaemon []byte

//go:embed embedded/uinputd.yaml
var embeddedConfig []byte

//go:embed embedded/uinputd.service
var embeddedSystemd []byte

var (
	version   = "dev"
	commit    = "unknown"
	buildTime = "unknown"

	socketPath  string
	layout      string
	charDelayMs int
	wordDelayMs int
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, styles.Error(err.Error()))
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use:   "uinput-client",
	Short: "uinput-client - Send input automation commands to uinputd",
	Long: `uinput-client communicates with the uinputd daemon
to send keyboard input automation commands.

Examples:
  uinput-client type "Hello, world!"
  uinput-client type --layout fr "Bonjour!"

  # Stream text from stdin with natural typing delays
  echo "Hello from stdin" | uinput-client stream
  cat document.txt | uinput-client stream --layout fr

  # SimulStreaming integration (filter timestamps, then stream)
  simulstreaming_output | awk '{$1=$2=""; print substr($0,3)}' | uinput-client stream --layout fr

  # Custom delays
  echo "Slow typing" | uinput-client stream --char-delay 100 --word-delay 300

  uinput-client key 28  # Send Enter key (keycode 28)

  # Installation
  uinput-client install daemon         # Install daemon binary
  uinput-client install systemd-service # Install systemd service`,
	Version: version,
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&socketPath, "socket", "s", "/run/uinputd.sock", "socket path")
	rootCmd.PersistentFlags().StringVarP(&layout, "layout", "l", "", "keyboard layout (us, fr, de, es, uk, it)")
}

var typeCmd = &cobra.Command{
	Use:   "type TEXT",
	Short: "Type text (batch mode)",
	Args:  cobra.ExactArgs(1),
	RunE:  runType,
}

var streamCmd = &cobra.Command{
	Use:   "stream",
	Short: "Stream text from stdin with real-time delays",
	Args:  cobra.MaximumNArgs(0),
	RunE:  runStream,
}

var keyCmd = &cobra.Command{
	Use:   "key KEYCODE",
	Short: "Send a single key press",
	Args:  cobra.ExactArgs(1),
	RunE:  runKey,
}

var pingCmd = &cobra.Command{
	Use:   "ping",
	Short: "Check if daemon is running",
	RunE:  runPing,
}

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install daemon or systemd service",
	Long:  `Install the embedded daemon binary or systemd service file.`,
}

var installDaemonCmd = &cobra.Command{
	Use:   "daemon",
	Short: "Install the uinputd daemon binary",
	Long: `Extract and install the embedded uinputd daemon binary to /usr/local/bin.
Requires root privileges.`,
	RunE: runInstallDaemon,
}

var installSystemdCmd = &cobra.Command{
	Use:   "systemd-service",
	Short: "Install systemd service unit",
	Long: `Install the systemd service unit file for uinputd.
The daemon must be installed first. Requires root privileges.`,
	RunE: runInstallSystemd,
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("uinput-client %s\n", version)
		fmt.Printf("  commit:     %s\n", commit)
		fmt.Printf("  build time: %s\n", buildTime)
	},
}

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Check system health and configuration",
	Long:  `Run health checks to verify uinputd installation and configuration.`,
	RunE:  runDoctor,
}

func init() {
	rootCmd.AddCommand(typeCmd)
	rootCmd.AddCommand(streamCmd)
	rootCmd.AddCommand(keyCmd)
	rootCmd.AddCommand(pingCmd)
	rootCmd.AddCommand(installCmd)
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(doctorCmd)

	installCmd.AddCommand(installDaemonCmd)
	installCmd.AddCommand(installSystemdCmd)

	// Stream command flags
	streamCmd.Flags().IntVar(&charDelayMs, "char-delay", 0, "delay between characters in ms (0=use config default)")
	streamCmd.Flags().IntVar(&wordDelayMs, "word-delay", 0, "delay between words in ms (0=use config default)")
}

func runType(cmd *cobra.Command, args []string) error {
	text := args[0]

	payload := protocol.TypePayload{
		Text:   text,
		Layout: layout,
	}

	return sendCommand(protocol.CommandType_Type, payload)
}

func runStream(cmd *cobra.Command, args []string) error {
	// Read from stdin and accumulate lines into continuous text
	var buffer strings.Builder
	scanner := bufio.NewScanner(os.Stdin)

	for scanner.Scan() {
		line := scanner.Text()
		if len(line) > 0 {
			// Add space before if buffer not empty to join segments
			if buffer.Len() > 0 {
				buffer.WriteString(" ")
			}
			buffer.WriteString(line)
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("reading stdin: %w", err)
	}

	text := strings.TrimSpace(buffer.String())
	if text == "" {
		return nil // Empty input, nothing to do
	}

	payload := protocol.StreamPayload{
		Text:      text,
		Layout:    layout,
		DelayMs:   wordDelayMs,
		CharDelay: charDelayMs,
	}

	return sendCommand(protocol.CommandType_Stream, payload)
}

func runKey(cmd *cobra.Command, args []string) error {
	var keycode uint16
	if _, err := fmt.Sscanf(args[0], "%d", &keycode); err != nil {
		return fmt.Errorf("invalid keycode: %w", err)
	}

	payload := protocol.KeyPayload{
		Keycode: keycode,
	}

	return sendCommand(protocol.CommandType_Key, payload)
}

func runPing(cmd *cobra.Command, args []string) error {
	start := time.Now()
	if err := sendCommand(protocol.CommandType_Ping, protocol.PingPayload{}); err != nil {
		return err
	}
	elapsed := time.Since(start)
	fmt.Println(styles.Success(fmt.Sprintf("pong (%.2fms)", float64(elapsed.Microseconds())/1000.0)))
	return nil
}

func sendCommand(cmdType protocol.CommandType, payload interface{}) error {
	// Connect to daemon
	conn, err := net.Dial("unix", socketPath)
	if err != nil {
		return fmt.Errorf("failed to connect to daemon (is uinputd running?): %w", err)
	}
	defer conn.Close()

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
	if err := json.NewEncoder(conn).Encode(&cmd); err != nil {
		return fmt.Errorf("failed to send command: %w", err)
	}

	// Read response
	var resp protocol.Response
	if err := json.NewDecoder(conn).Decode(&resp); err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	// Handle response
	if !resp.Success {
		return fmt.Errorf("daemon error: %s", resp.Error)
	}

	if resp.Message != "" {
		fmt.Println(styles.Success(resp.Message))
	}

	return nil
}

// ensureRoot checks if running as root, and if not, re-executes with sudo
func ensureRoot() error {
	if os.Geteuid() == 0 {
		return nil // Already root
	}

	// Re-execute with sudo
	fmt.Println(styles.Info("Root privileges required. Re-executing with sudo..."))

	// Build sudo command with same arguments
	args := append([]string{os.Args[0]}, os.Args[1:]...)
	cmd := exec.Command("sudo", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("sudo execution failed: %w", err)
	}

	// Exit after sudo execution completes
	os.Exit(0)
	return nil
}

func runInstallDaemon(cmd *cobra.Command, args []string) error {
	// Ensure we're running as root (will re-exec with sudo if needed)
	if err := ensureRoot(); err != nil {
		return err
	}

	fmt.Println(styles.Info("Installing uinputd daemon..."))

	// Use installer package for installation logic
	if err := installer.InstallDaemon(embeddedDaemon, embeddedConfig); err != nil {
		return err
	}

	// Get the username that was added to the group
	username, err := installer.GetInstalledUsername()
	if err != nil {
		username = "your-user"
	}

	fmt.Println(styles.Success("Daemon installed: /usr/local/bin/uinputd"))
	fmt.Println(styles.Success("Config installed: /etc/uinputd/uinputd.yaml"))
	fmt.Println(styles.Success(fmt.Sprintf("User '%s' added to 'input' group", username)))

	fmt.Println(styles.Section("Installation complete!"))
	fmt.Println(styles.Bold("Next steps:"))
	fmt.Println(styles.Step(1, "Install systemd service: sudo uinput-client install systemd-service"))
	fmt.Println(styles.Step(2, "Enable service:           sudo systemctl enable uinputd"))
	fmt.Println(styles.Step(3, "Start service:            sudo systemctl start uinputd"))
	fmt.Println(styles.Step(4, fmt.Sprintf("Activate group (no logout): newgrp input")))
	fmt.Println(styles.Dim("  (or logout and login for group changes to take effect)"))

	return nil
}

func runInstallSystemd(cmd *cobra.Command, args []string) error {
	// Ensure we're running as root (will re-exec with sudo if needed)
	if err := ensureRoot(); err != nil {
		return err
	}

	fmt.Println(styles.Info("Installing systemd service..."))

	// Use installer package for installation logic
	if err := installer.InstallSystemdService(embeddedSystemd); err != nil {
		return err
	}

	fmt.Println(styles.Success("Service installed: /etc/systemd/system/uinputd.service"))
	fmt.Println(styles.Success("Systemd reloaded"))

	fmt.Println(styles.Section("Systemd service installed!"))
	fmt.Println(styles.Bold("Next steps:"))
	fmt.Println(styles.ListItem("Enable and start: sudo systemctl enable --now uinputd"))
	fmt.Println(styles.ListItem("Check status:     systemctl status uinputd"))
	fmt.Println(styles.ListItem("Verify setup:     uinput-client doctor"))

	return nil
}

func runDoctor(cmd *cobra.Command, args []string) error {
	fmt.Println(styles.Section("Running health checks..."))
	fmt.Println()

	results := doctor.CheckAll(socketPath)

	// Print results
	for _, result := range results {
		switch result.Status {
		case doctor.StatusOK:
			fmt.Println(styles.Success(result.Name))
			fmt.Printf("  %s\n", styles.Dim(result.Message))
		case doctor.StatusWarning:
			fmt.Println(styles.Warning(result.Name))
			fmt.Printf("  %s\n", result.Message)
			if result.Fix != "" {
				fmt.Printf("  %s %s\n", styles.Dim("Fix:"), result.Fix)
			}
		case doctor.StatusError:
			fmt.Println(styles.Error(result.Name))
			fmt.Printf("  %s\n", result.Message)
			if result.Fix != "" {
				fmt.Printf("  %s %s\n", styles.Dim("Fix:"), result.Fix)
			}
		}
		fmt.Println()
	}

	// Summary
	if doctor.HasErrors(results) {
		fmt.Println(styles.Error("Some checks failed. Please fix the errors above."))
		return fmt.Errorf("health checks failed")
	} else if doctor.HasWarnings(results) {
		fmt.Println(styles.Warning("All critical checks passed, but there are warnings."))
	} else {
		fmt.Println(styles.Success("All checks passed! Your uinputd setup is healthy."))
	}

	return nil
}
