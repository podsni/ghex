package update

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// PermissionError contains details about permission issues
type PermissionError struct {
	Path        string
	NeedsSudo   bool
	Instruction string
}

func (e *PermissionError) Error() string {
	return e.Instruction
}

// CheckUpdatePermissions checks if we have permission to update the binary
func CheckUpdatePermissions() (*PermissionError, error) {
	binaryPath, err := GetCurrentBinaryPath()
	if err != nil {
		return nil, err
	}

	// Check if we can write to the binary location
	if err := CheckWritePermission(binaryPath); err != nil {
		return createPermissionError(binaryPath), nil
	}

	return nil, nil
}

// createPermissionError creates a helpful error message based on the platform
func createPermissionError(binaryPath string) *PermissionError {
	dir := filepath.Dir(binaryPath)

	if runtime.GOOS == "windows" {
		return &PermissionError{
			Path:      binaryPath,
			NeedsSudo: false,
			Instruction: fmt.Sprintf(
				"Cannot write to %s\n\n"+
					"Please try one of the following:\n"+
					"1. Run the command prompt as Administrator\n"+
					"2. Move ghex to a user-writable location\n"+
					"3. Manually download the update from GitHub releases",
				dir,
			),
		}
	}

	// Unix-like systems
	needsSudo := isSystemPath(dir)

	if needsSudo {
		return &PermissionError{
			Path:      binaryPath,
			NeedsSudo: true,
			Instruction: fmt.Sprintf(
				"Cannot write to %s (requires elevated permissions)\n\n"+
					"Please try one of the following:\n"+
					"1. Run with sudo: sudo ghex update\n"+
					"2. Move ghex to a user-writable location (e.g., ~/.local/bin)\n"+
					"3. Manually download the update from GitHub releases",
				dir,
			),
		}
	}

	return &PermissionError{
		Path:      binaryPath,
		NeedsSudo: false,
		Instruction: fmt.Sprintf(
			"Cannot write to %s\n\n"+
				"Please check the directory permissions or try:\n"+
				"chmod u+w %s",
			dir, dir,
		),
	}
}

// isSystemPath checks if a path is a system directory that typically requires sudo
func isSystemPath(path string) bool {
	systemPaths := []string{
		"/usr/local/bin",
		"/usr/bin",
		"/bin",
		"/opt",
		"/usr/local",
	}

	for _, sp := range systemPaths {
		if strings.HasPrefix(path, sp) {
			return true
		}
	}

	return false
}

// IsRunningAsRoot checks if the current process is running as root/admin
func IsRunningAsRoot() bool {
	if runtime.GOOS == "windows" {
		// On Windows, check if running as administrator
		// This is a simplified check
		return false
	}

	return os.Geteuid() == 0
}

// GetSuggestedInstallPath returns a suggested path for user installation
func GetSuggestedInstallPath() string {
	if runtime.GOOS == "windows" {
		return filepath.Join(os.Getenv("LOCALAPPDATA"), "Programs", "ghex")
	}

	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, ".local", "bin")
}
