package doctor

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"os/user"
	"strings"
)

// CheckResult represents the result of a health check
type CheckResult struct {
	Name    string
	Status  Status
	Message string
	Fix     string // Suggestion for fixing the issue
}

// Status represents the health check status
type Status int

const (
	StatusOK Status = iota
	StatusWarning
	StatusError
)

// CheckAll runs all health checks
func CheckAll(socketPath string) []CheckResult {
	results := []CheckResult{
		checkDaemonInstalled(),
		checkDaemonRunning(socketPath),
		checkUserInInputGroup(),
		checkSocketPermissions(socketPath),
		checkUinputDevice(),
	}
	return results
}

// checkDaemonInstalled verifies the daemon binary exists
func checkDaemonInstalled() CheckResult {
	_, err := exec.LookPath("uinputd")
	if err != nil {
		return CheckResult{
			Name:    "Daemon Installation",
			Status:  StatusError,
			Message: "uinputd daemon not found in PATH",
			Fix:     "Run: sudo uinput-client install daemon",
		}
	}

	return CheckResult{
		Name:    "Daemon Installation",
		Status:  StatusOK,
		Message: "uinputd daemon is installed",
	}
}

// checkDaemonRunning verifies the daemon is running by checking socket connectivity
func checkDaemonRunning(socketPath string) CheckResult {
	conn, err := net.Dial("unix", socketPath)
	if err != nil {
		// Check if systemd service exists
		cmd := exec.Command("systemctl", "is-active", "uinputd")
		output, _ := cmd.CombinedOutput()
		status := strings.TrimSpace(string(output))

		if status == "inactive" {
			return CheckResult{
				Name:    "Daemon Status",
				Status:  StatusError,
				Message: "Daemon is installed but not running",
				Fix:     "Run: sudo systemctl start uinputd",
			}
		}

		return CheckResult{
			Name:    "Daemon Status",
			Status:  StatusError,
			Message: fmt.Sprintf("Cannot connect to daemon socket: %v", err),
			Fix:     "Run: sudo systemctl start uinputd (or sudo uinputd for manual start)",
		}
	}
	defer conn.Close()

	return CheckResult{
		Name:    "Daemon Status",
		Status:  StatusOK,
		Message: fmt.Sprintf("Daemon is running (socket: %s)", socketPath),
	}
}

// checkUserInInputGroup verifies current user is in the input group
func checkUserInInputGroup() CheckResult {
	currentUser, err := user.Current()
	if err != nil {
		return CheckResult{
			Name:    "User Permissions",
			Status:  StatusWarning,
			Message: fmt.Sprintf("Cannot determine current user: %v", err),
		}
	}

	// Get user's groups
	groups, err := currentUser.GroupIds()
	if err != nil {
		return CheckResult{
			Name:    "User Permissions",
			Status:  StatusWarning,
			Message: fmt.Sprintf("Cannot get user groups: %v", err),
		}
	}

	// Check if user is in input group
	for _, gid := range groups {
		group, err := user.LookupGroupId(gid)
		if err != nil {
			continue
		}
		if group.Name == "input" {
			return CheckResult{
				Name:    "User Permissions",
				Status:  StatusOK,
				Message: fmt.Sprintf("User '%s' is in the 'input' group", currentUser.Username),
			}
		}
	}

	return CheckResult{
		Name:    "User Permissions",
		Status:  StatusWarning,
		Message: fmt.Sprintf("User '%s' is NOT in the 'input' group", currentUser.Username),
		Fix:     fmt.Sprintf("Run: sudo usermod -aG input %s (then logout and login)", currentUser.Username),
	}
}

// checkSocketPermissions verifies socket file has correct permissions
func checkSocketPermissions(socketPath string) CheckResult {
	info, err := os.Stat(socketPath)
	if err != nil {
		return CheckResult{
			Name:    "Socket Permissions",
			Status:  StatusWarning,
			Message: "Socket file not found (daemon may not be running)",
		}
	}

	mode := info.Mode().Perm()
	expectedMode := os.FileMode(0660)

	if mode == expectedMode {
		return CheckResult{
			Name:    "Socket Permissions",
			Status:  StatusOK,
			Message: fmt.Sprintf("Socket has correct permissions: %o", mode),
		}
	}

	return CheckResult{
		Name:    "Socket Permissions",
		Status:  StatusWarning,
		Message: fmt.Sprintf("Socket has permissions %o (expected %o)", mode, expectedMode),
		Fix:     "This is usually set by the daemon. Try restarting: sudo systemctl restart uinputd",
	}
}

// checkUinputDevice verifies /dev/uinput exists and is accessible
func checkUinputDevice() CheckResult {
	info, err := os.Stat("/dev/uinput")
	if err != nil {
		return CheckResult{
			Name:    "UInput Device",
			Status:  StatusError,
			Message: "/dev/uinput not found (kernel module not loaded?)",
			Fix:     "Run: sudo modprobe uinput",
		}
	}

	// Check if readable (this will fail if not root, which is expected)
	file, err := os.OpenFile("/dev/uinput", os.O_RDONLY, 0)
	if err == nil {
		file.Close()
		return CheckResult{
			Name:    "UInput Device",
			Status:  StatusOK,
			Message: "/dev/uinput exists and is accessible",
		}
	}

	// Not accessible to current user, but that's fine since daemon runs as root
	mode := info.Mode().Perm()
	return CheckResult{
		Name:    "UInput Device",
		Status:  StatusOK,
		Message: fmt.Sprintf("/dev/uinput exists (permissions: %o, accessible by root)", mode),
	}
}

// HasErrors returns true if any check has an error status
func HasErrors(results []CheckResult) bool {
	for _, r := range results {
		if r.Status == StatusError {
			return true
		}
	}
	return false
}

// HasWarnings returns true if any check has a warning status
func HasWarnings(results []CheckResult) bool {
	for _, r := range results {
		if r.Status == StatusWarning {
			return true
		}
	}
	return false
}
