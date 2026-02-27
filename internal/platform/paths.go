package platform

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// GetHomeDir returns the user's home directory
func GetHomeDir() string {
	if IsWindows() {
		// Git Bash sets HOME to Unix-style path; prefer it over USERPROFILE
		if home := os.Getenv("HOME"); home != "" {
			return home
		}
		if home := os.Getenv("USERPROFILE"); home != "" {
			return home
		}
	}

	home, err := os.UserHomeDir()
	if err != nil {
		// Fallback to HOME environment variable
		return os.Getenv("HOME")
	}
	return home
}

// GetSSHDir returns the SSH directory path
func GetSSHDir() string {
	return filepath.Join(GetHomeDir(), ".ssh")
}

// GetConfigDir returns the configuration directory for an application
func GetConfigDir(appName string) string {
	if IsWindows() {
		// Windows: Use %APPDATA%
		appData := os.Getenv("APPDATA")
		if appData == "" {
			appData = filepath.Join(GetHomeDir(), "AppData", "Roaming")
		}
		return filepath.Join(appData, appName)
	}

	// Linux/macOS: Use XDG_CONFIG_HOME or ~/.config
	configHome := os.Getenv("XDG_CONFIG_HOME")
	if configHome == "" {
		configHome = filepath.Join(GetHomeDir(), ".config")
	}
	return filepath.Join(configHome, appName)
}

// GetGitCredentialsPath returns the path to .git-credentials file
func GetGitCredentialsPath() string {
	return filepath.Join(GetHomeDir(), ".git-credentials")
}

// NormalizePath normalizes a file path for the current platform
func NormalizePath(path string) string {
	if path == "" {
		return path
	}

	// Expand tilde
	path = ExpandPath(path)

	// Clean the path
	return filepath.Clean(path)
}

// ToSSHPath converts a file path to a format compatible with SSH commands on all platforms.
// On Windows with Git Bash/MSYS2, paths like C:/Users/... must be converted to /c/Users/...
// because SSH (OpenSSH running inside Git Bash) requires POSIX-style paths.
// In PowerShell/cmd, forward-slash paths like C:/Users/... are accepted directly.
func ToSSHPath(path string) string {
	// First, normalize backslashes to forward slashes
	path = strings.ReplaceAll(path, "\\", "/")

	// If running in Git Bash / MSYS2 on Windows, convert drive letter to POSIX mount
	// e.g. C:/Users/hendr/.ssh/key -> /c/Users/hendr/.ssh/key
	if IsGitBashEnv() && len(path) >= 3 && path[1] == ':' && path[2] == '/' {
		drive := strings.ToLower(string(path[0]))
		path = "/" + drive + path[2:]
	}

	return path
}

// IsGitBashEnv returns true if the process is running inside Git Bash or MSYS2 on Windows.
// This is detected by the presence of the MSYSTEM environment variable (set by Git Bash/MSYS2).
// Git Bash sets MSYSTEM to "MINGW64" or "MINGW32"; MSYS2 sets it to "MSYS".
// This check is intentionally narrow: only convert paths when on Windows AND MSYSTEM is set,
// to avoid incorrectly converting paths in PowerShell or cmd.exe.
func IsGitBashEnv() bool {
	return runtime.GOOS == "windows" && os.Getenv("MSYSTEM") != ""
}

// ExpandPath expands environment variables and tilde in a path
func ExpandPath(path string) string {
	if path == "" {
		return path
	}

	// Expand tilde
	if strings.HasPrefix(path, "~") {
		home := GetHomeDir()
		if path == "~" {
			return home
		}
		if strings.HasPrefix(path, "~/") || strings.HasPrefix(path, "~\\") {
			return filepath.Join(home, path[2:])
		}
	}

	// Expand environment variables
	if IsWindows() {
		// Handle %VAR% style (Windows-native)
		path = expandWindowsEnvVars(path)
		// Also handle $VAR style (for Git Bash compatibility)
		path = os.ExpandEnv(path)
	} else {
		// Unix: $VAR or ${VAR}
		path = os.ExpandEnv(path)
	}

	return path
}

// expandWindowsEnvVars expands %VAR% style environment variables
func expandWindowsEnvVars(path string) string {
	result := path
	offset := 0
	for {
		start := strings.Index(result[offset:], "%")
		if start == -1 {
			break
		}
		start += offset
		end := strings.Index(result[start+1:], "%")
		if end == -1 {
			break
		}
		end = start + 1 + end
		key := result[start+1 : end]
		if val := os.Getenv(key); val != "" {
			result = result[:start] + val + result[end+1:]
			// Don't advance offset - the replacement may contain more vars
		} else {
			// Skip past this %VAR% to avoid infinite loop on unknown vars
			offset = end + 1
		}
	}
	return result
}

// EnsureDir creates a directory if it doesn't exist
func EnsureDir(path string, perm os.FileMode) error {
	if path == "" || path == "." || path == "./" || path == ".\\" {
		return nil
	}

	return os.MkdirAll(path, perm)
}

// FileExists checks if a file exists
func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// IsDir checks if a path is a directory
func IsDir(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// GetTempDir returns the system temporary directory
func GetTempDir() string {
	return os.TempDir()
}
