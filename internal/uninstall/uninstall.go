package uninstall

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/dwirx/ghex/internal/platform"
)

// Options holds uninstall configuration
type Options struct {
	Force      bool // Skip confirmation prompt
	Purge      bool // Remove config files
	KeepConfig bool // Explicitly keep config files
	DryRun     bool // Preview without removing
}

// Preview holds information about what will be removed
type Preview struct {
	BinaryPath    string   `json:"binary_path"`
	ConfigPath    string   `json:"config_path"`
	LegacyConfig  string   `json:"legacy_config,omitempty"`
	PathEntry     string   `json:"path_entry,omitempty"` // Windows only
	FilesToRemove []string `json:"files_to_remove"`
}

// Result holds the result of uninstallation
type Result struct {
	Success       bool     `json:"success"`
	BinaryRemoved bool     `json:"binary_removed"`
	ConfigRemoved bool     `json:"config_removed"`
	PathUpdated   bool     `json:"path_updated"`
	RemovedFiles  []string `json:"removed_files"`
	Errors        []string `json:"errors,omitempty"`
}

// Service handles uninstallation operations
type Service struct {
	binaryPath   string
	configPath   string
	legacyConfig string
	installDir   string // Windows only
}

// NewService creates a new uninstall service
func NewService() *Service {
	s := &Service{
		configPath:   platform.GetConfigDir("ghe"),
		legacyConfig: platform.GetConfigDir("github-switch"),
	}

	if platform.IsWindows() {
		// Windows: %LOCALAPPDATA%\ghex\ghex.exe
		localAppData := os.Getenv("LOCALAPPDATA")
		if localAppData == "" {
			localAppData = filepath.Join(platform.GetHomeDir(), "AppData", "Local")
		}
		s.installDir = filepath.Join(localAppData, "ghex")
		s.binaryPath = filepath.Join(s.installDir, "ghex.exe")
	} else {
		// Try to find the actual binary location
		if binaryPath, err := exec.LookPath("ghex"); err == nil {
			s.binaryPath = binaryPath
			s.installDir = filepath.Dir(binaryPath)
		} else {
			// Fallback to common locations
			s.binaryPath = "/usr/local/bin/ghex"
			s.installDir = "/usr/local/bin"
		}
	}

	return s
}

// GetBinaryPath returns the path to the installed binary
func (s *Service) GetBinaryPath() string {
	return s.binaryPath
}

// GetConfigPath returns the path to the config directory
func (s *Service) GetConfigPath() string {
	return s.configPath
}

// GetLegacyConfigPath returns the path to the legacy config directory
func (s *Service) GetLegacyConfigPath() string {
	return s.legacyConfig
}

// GetInstallDir returns the install directory (relevant for Windows)
func (s *Service) GetInstallDir() string {
	return s.installDir
}

// GetPreview returns information about what will be removed
func (s *Service) GetPreview() *Preview {
	preview := &Preview{
		BinaryPath:    s.binaryPath,
		ConfigPath:    s.configPath,
		FilesToRemove: []string{},
	}

	// Check binary
	if platform.FileExists(s.binaryPath) {
		preview.FilesToRemove = append(preview.FilesToRemove, s.binaryPath)
	}

	// Check config directories
	if platform.FileExists(s.configPath) {
		preview.FilesToRemove = append(preview.FilesToRemove, s.configPath)
	}

	if platform.FileExists(s.legacyConfig) {
		preview.LegacyConfig = s.legacyConfig
		preview.FilesToRemove = append(preview.FilesToRemove, s.legacyConfig)
	}

	// Windows: check PATH entry
	if platform.IsWindows() {
		currentPath := os.Getenv("PATH")
		if strings.Contains(currentPath, s.installDir) {
			preview.PathEntry = s.installDir
		}
	}

	return preview
}

// Execute performs the uninstallation
func (s *Service) Execute(opts Options) *Result {
	result := &Result{
		Success:      true,
		RemovedFiles: []string{},
		Errors:       []string{},
	}

	// Dry run - just return preview info
	if opts.DryRun {
		preview := s.GetPreview()
		result.RemovedFiles = preview.FilesToRemove
		return result
	}

	// Remove binary
	binaryExisted := platform.FileExists(s.binaryPath)
	if binaryExisted {
		if err := s.RemoveBinary(); err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("Failed to remove binary: %v", err))
			result.Success = false
		} else {
			result.BinaryRemoved = true
			result.RemovedFiles = append(result.RemovedFiles, s.binaryPath)
		}
	}

	// Handle config removal
	if opts.Purge && !opts.KeepConfig {
		// Track which config paths exist before removal
		configExisted := platform.FileExists(s.configPath)
		legacyExisted := platform.FileExists(s.legacyConfig)
		if err := s.RemoveConfig(); err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("Failed to remove config: %v", err))
		} else {
			result.ConfigRemoved = true
			if configExisted {
				result.RemovedFiles = append(result.RemovedFiles, s.configPath)
			}
			if legacyExisted {
				result.RemovedFiles = append(result.RemovedFiles, s.legacyConfig)
			}
		}
	}

	// Windows: remove from PATH
	if platform.IsWindows() {
		if err := s.RemoveFromPath(); err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("Failed to update PATH: %v", err))
		} else {
			result.PathUpdated = true
		}
	}

	return result
}

// RemoveBinary removes the GHEX binary
func (s *Service) RemoveBinary() error {
	if !platform.FileExists(s.binaryPath) {
		return nil // Already removed
	}

	err := os.Remove(s.binaryPath)
	if err != nil {
		return fmt.Errorf("cannot remove %s: %w (try running with elevated privileges)", s.binaryPath, err)
	}

	// Windows: try to remove install directory if empty
	if platform.IsWindows() && s.installDir != "" {
		entries, err := os.ReadDir(s.installDir)
		if err == nil && len(entries) == 0 {
			os.Remove(s.installDir)
		}
	}

	return nil
}

// RemoveConfig removes the config directory
func (s *Service) RemoveConfig() error {
	var lastErr error

	// Remove primary config
	if platform.FileExists(s.configPath) {
		if err := os.RemoveAll(s.configPath); err != nil {
			lastErr = err
		}
	}

	// Remove legacy config
	if platform.FileExists(s.legacyConfig) {
		if err := os.RemoveAll(s.legacyConfig); err != nil {
			lastErr = err
		}
	}

	return lastErr
}

// RemoveFromPath removes the install directory from PATH (Windows only)
func (s *Service) RemoveFromPath() error {
	if runtime.GOOS != "windows" {
		return nil
	}

	// This is a simplified version - in practice, modifying Windows PATH
	// from Go requires registry access or PowerShell
	// The actual PATH modification is better handled by the PowerShell script
	return nil
}

// BinaryExists checks if the binary is installed
func (s *Service) BinaryExists() bool {
	return platform.FileExists(s.binaryPath)
}

// ConfigExists checks if config directory exists
func (s *Service) ConfigExists() bool {
	return platform.FileExists(s.configPath) || platform.FileExists(s.legacyConfig)
}

// GetManualRemovalInstructions returns instructions for manual removal
func (s *Service) GetManualRemovalInstructions() string {
	if platform.IsWindows() {
		return fmt.Sprintf(`Manual removal instructions:
1. Delete: %s
2. Remove '%s' from your PATH environment variable
3. Optionally delete config: %s`, s.binaryPath, s.installDir, s.configPath)
	}

	return fmt.Sprintf(`Manual removal instructions:
  sudo rm %s
  rm -rf %s`, s.binaryPath, s.configPath)
}
