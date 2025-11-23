package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/exec"
	"time"

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
	version    = "dev"
	socketPath string
	layout     string
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
  uinput-client stream < input.txt
  uinput-client key 28  # Send Enter key (keycode 28)

  # Installation
  uinput-client install daemon         # Install daemon binary
  uinput-client install systemd-service # Install systemd service`,
	Version: version,
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&socketPath, "socket", "s", "/tmp/.uinputd.sock", "socket path")
	rootCmd.PersistentFlags().StringVarP(&layout, "layout", "l", "", "keyboard layout (us, fr, de, es, uk, it)")
}

var typeCmd = &cobra.Command{
	Use:   "type TEXT",
	Short: "Type text (batch mode)",
	Args:  cobra.ExactArgs(1),
	RunE:  runType,
}

var streamCmd = &cobra.Command{
	Use:   "stream TEXT",
	Short: "Stream text in real-time",
	Args:  cobra.ExactArgs(1),
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

func init() {
	rootCmd.AddCommand(typeCmd)
	rootCmd.AddCommand(streamCmd)
	rootCmd.AddCommand(keyCmd)
	rootCmd.AddCommand(pingCmd)
	rootCmd.AddCommand(installCmd)

	installCmd.AddCommand(installDaemonCmd)
	installCmd.AddCommand(installSystemdCmd)
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
	text := args[0]

	payload := protocol.StreamPayload{
		Text:   text,
		Layout: layout,
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

	// Write daemon binary
	daemonPath := "/usr/local/bin/uinputd"
	if err := os.WriteFile(daemonPath, embeddedDaemon, 0755); err != nil {
		return fmt.Errorf("failed to write daemon: %w", err)
	}

	fmt.Println(styles.Success("Daemon installed: " + daemonPath))

	// Create config directory
	configDir := "/etc/uinputd"
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Install default config if it doesn't exist
	configPath := configDir + "/uinputd.yaml"
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		if err := os.WriteFile(configPath, embeddedConfig, 0644); err != nil {
			return fmt.Errorf("failed to write config: %w", err)
		}

		fmt.Println(styles.Success("Config installed: " + configPath))
	} else {
		fmt.Println(styles.Info("Config already exists: " + configPath))
	}

	// Try to set group ownership to 'input'
	if err := setInputGroup(daemonPath); err != nil {
		fmt.Println(styles.Warning("Could not set input group: " + err.Error()))
	}

	fmt.Println(styles.Section("Installation complete!"))
	fmt.Println(styles.Bold("Next steps:"))
	fmt.Println(styles.Step(1, "Install systemd service: sudo uinput-client install systemd-service"))
	fmt.Println(styles.Step(2, "Enable service:           sudo systemctl enable uinputd"))
	fmt.Println(styles.Step(3, "Start service:            sudo systemctl start uinputd"))

	return nil
}

func runInstallSystemd(cmd *cobra.Command, args []string) error {
	// Ensure we're running as root (will re-exec with sudo if needed)
	if err := ensureRoot(); err != nil {
		return err
	}

	fmt.Println(styles.Info("Installing systemd service..."))

	// Check if daemon is installed
	daemonPath := "/usr/local/bin/uinputd"
	if _, err := os.Stat(daemonPath); os.IsNotExist(err) {
		return fmt.Errorf("daemon not found at %s\nInstall it first: uinput-client install daemon", daemonPath)
	}

	// Write service file
	servicePath := "/etc/systemd/system/uinputd.service"
	if err := os.WriteFile(servicePath, embeddedSystemd, 0644); err != nil {
		return fmt.Errorf("failed to write service file: %w", err)
	}

	fmt.Println(styles.Success("Service installed: " + servicePath))

	// Reload systemd
	fmt.Println(styles.Info("Reloading systemd..."))
	reloadCmd := exec.Command("systemctl", "daemon-reload")
	if err := reloadCmd.Run(); err != nil {
		return fmt.Errorf("failed to reload systemd: %w", err)
	}

	fmt.Println(styles.Success("Systemd reloaded"))
	fmt.Println(styles.Section("Systemd service installed!"))
	fmt.Println(styles.Bold("Next steps:"))
	fmt.Println(styles.ListItem("Enable service: sudo systemctl enable uinputd"))
	fmt.Println(styles.ListItem("Start service:  sudo systemctl start uinputd"))
	fmt.Println(styles.ListItem("Check status:   systemctl status uinputd"))
	fmt.Println(styles.Section("User Configuration"))
	fmt.Println(styles.Info("Add your user to the 'input' group to use the client:"))
	fmt.Println(styles.ListItem("sudo usermod -aG input $USER"))
	fmt.Println(styles.Dim("  (logout and login again for group changes to take effect)"))

	return nil
}

// setInputGroup attempts to set the group ownership to 'input' (GID typically 104 or similar)
func setInputGroup(path string) error {
	// This is a best-effort attempt; errors are non-fatal
	cmd := exec.Command("chgrp", "input", path)
	return cmd.Run()
}
