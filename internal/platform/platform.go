package platform

import (
	"os"
	"runtime"
)

// Platform holds information about the current platform
type Platform struct {
	IsWindows bool
	IsLinux   bool
	IsMacOS   bool
	IsUnix    bool
	OS        string
	Arch      string
}

// Current returns information about the current platform
func Current() Platform {
	return Platform{
		IsWindows: runtime.GOOS == "windows",
		IsLinux:   runtime.GOOS == "linux",
		IsMacOS:   runtime.GOOS == "darwin",
		IsUnix:    runtime.GOOS != "windows",
		OS:        runtime.GOOS,
		Arch:      runtime.GOARCH,
	}
}

// IsWindows returns true if running on Windows
func IsWindows() bool {
	return runtime.GOOS == "windows"
}

// IsLinux returns true if running on Linux
func IsLinux() bool {
	return runtime.GOOS == "linux"
}

// IsMacOS returns true if running on macOS
func IsMacOS() bool {
	return runtime.GOOS == "darwin"
}

// IsUnix returns true if running on a Unix-like system
func IsUnix() bool {
	return runtime.GOOS != "windows"
}

// DetectShell returns the current shell environment
func DetectShell() string {
	if IsWindows() {
		// Check for PowerShell
		if os.Getenv("PSModulePath") != "" {
			return "powershell"
		}
		// Check for Git Bash or WSL
		if shell := os.Getenv("SHELL"); shell != "" {
			return shell
		}
		return "cmd"
	}

	// Unix-like systems
	shell := os.Getenv("SHELL")
	if shell == "" {
		return "bash"
	}
	return shell
}
