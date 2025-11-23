package installer

import (
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"strconv"
	"strings"
)

// ensureInputGroupExists creates the input group if it doesn't exist
func ensureInputGroupExists() error {
	// Check if group exists
	_, err := user.LookupGroup("input")
	if err == nil {
		return nil // Group already exists
	}

	// Create the group
	cmd := exec.Command("groupadd", "--system", "input")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to create input group: %w (output: %s)", err, string(output))
	}

	return nil
}

// addUserToInputGroup adds the specified user to the input group
func addUserToInputGroup(username string) error {
	cmd := exec.Command("usermod", "-aG", "input", username)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to add user to input group: %w (output: %s)", err, string(output))
	}
	return nil
}

// getCurrentNonRootUser gets the username of the user who invoked sudo
// or the current user if not running under sudo
func getCurrentNonRootUser() (string, error) {
	// Check SUDO_USER environment variable first
	if sudoUser := os.Getenv("SUDO_USER"); sudoUser != "" {
		return sudoUser, nil
	}

	// Fall back to current user
	currentUser, err := user.Current()
	if err != nil {
		return "", fmt.Errorf("failed to get current user: %w", err)
	}

	return currentUser.Username, nil
}

// setGroupOwnership sets the group ownership of a file to 'input'
func setGroupOwnership(path string) error {
	group, err := user.LookupGroup("input")
	if err != nil {
		return fmt.Errorf("input group not found: %w", err)
	}

	gid, err := strconv.Atoi(group.Gid)
	if err != nil {
		return fmt.Errorf("invalid group ID: %w", err)
	}

	// Change group ownership (-1 means don't change UID)
	if err := os.Chown(path, -1, gid); err != nil {
		return fmt.Errorf("failed to change group ownership: %w", err)
	}

	return nil
}

// InstallDaemon installs the daemon binary and configuration
func InstallDaemon(daemonBinary, configData []byte) error {
	// Ensure input group exists
	if err := ensureInputGroupExists(); err != nil {
		return fmt.Errorf("failed to ensure input group exists: %w", err)
	}

	// Write daemon binary
	daemonPath := "/usr/local/bin/uinputd"
	if err := os.WriteFile(daemonPath, daemonBinary, 0755); err != nil {
		return fmt.Errorf("failed to write daemon: %w", err)
	}

	// Set group ownership to input
	if err := setGroupOwnership(daemonPath); err != nil {
		// Non-fatal, just warn
		fmt.Printf("Warning: could not set group ownership: %v\n", err)
	}

	// Create config directory
	configDir := "/etc/uinputd"
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Install default config if it doesn't exist
	configPath := configDir + "/uinputd.yaml"
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		if err := os.WriteFile(configPath, configData, 0644); err != nil {
			return fmt.Errorf("failed to write config: %w", err)
		}
	}

	// Add the user who invoked sudo to the input group
	username, err := getCurrentNonRootUser()
	if err != nil {
		return fmt.Errorf("failed to determine user: %w", err)
	}

	if err := addUserToInputGroup(username); err != nil {
		return fmt.Errorf("failed to add user to input group: %w", err)
	}

	return nil
}

// InstallSystemdService installs the systemd service file
func InstallSystemdService(serviceData []byte) error {
	// Check if daemon is installed
	daemonPath := "/usr/local/bin/uinputd"
	if _, err := os.Stat(daemonPath); os.IsNotExist(err) {
		return fmt.Errorf("daemon not found at %s\nInstall it first: uinput-client install daemon", daemonPath)
	}

	// Write service file
	servicePath := "/etc/systemd/system/uinputd.service"
	if err := os.WriteFile(servicePath, serviceData, 0644); err != nil {
		return fmt.Errorf("failed to write service file: %w", err)
	}

	// Reload systemd
	cmd := exec.Command("systemctl", "daemon-reload")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to reload systemd: %w (output: %s)", err, string(output))
	}

	return nil
}

// GetInstalledUsername returns the username that should be added to groups
func GetInstalledUsername() (string, error) {
	return getCurrentNonRootUser()
}

// CheckInputGroupExists checks if the input group exists
func CheckInputGroupExists() bool {
	_, err := user.LookupGroup("input")
	return err == nil
}

// CheckUserInInputGroup checks if a user is in the input group
func CheckUserInInputGroup(username string) (bool, error) {
	u, err := user.Lookup(username)
	if err != nil {
		return false, fmt.Errorf("user not found: %w", err)
	}

	gids, err := u.GroupIds()
	if err != nil {
		return false, fmt.Errorf("failed to get user groups: %w", err)
	}

	inputGroup, err := user.LookupGroup("input")
	if err != nil {
		return false, nil // Group doesn't exist
	}

	for _, gid := range gids {
		if gid == inputGroup.Gid {
			return true, nil
		}
	}

	return false, nil
}

// GetGroupMembers returns the list of users in the input group
func GetGroupMembers() ([]string, error) {
	group, err := user.LookupGroup("input")
	if err != nil {
		return nil, fmt.Errorf("input group not found: %w", err)
	}

	// Read /etc/group to get members
	data, err := os.ReadFile("/etc/group")
	if err != nil {
		return nil, fmt.Errorf("failed to read /etc/group: %w", err)
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		parts := strings.Split(line, ":")
		if len(parts) >= 4 && parts[0] == "input" && parts[2] == group.Gid {
			if parts[3] == "" {
				return []string{}, nil
			}
			return strings.Split(parts[3], ","), nil
		}
	}

	return []string{}, nil
}
