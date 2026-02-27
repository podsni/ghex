package update

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

// BinaryManager handles binary file operations
type BinaryManager struct {
	BinaryPath string
	BackupPath string
}

// NewBinaryManager creates a new BinaryManager
func NewBinaryManager() (*BinaryManager, error) {
	binaryPath, err := GetCurrentBinaryPath()
	if err != nil {
		return nil, err
	}

	backupPath := getBackupPath()

	return &BinaryManager{
		BinaryPath: binaryPath,
		BackupPath: backupPath,
	}, nil
}

// GetCurrentBinaryPath returns the path to the running binary
func GetCurrentBinaryPath() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("failed to get executable path: %w", err)
	}

	// Resolve symlinks
	exe, err = filepath.EvalSymlinks(exe)
	if err != nil {
		return "", fmt.Errorf("failed to resolve symlinks: %w", err)
	}

	return exe, nil
}

// getBackupPath returns the backup file path based on platform
func getBackupPath() string {
	if runtime.GOOS == "windows" {
		baseDir := os.Getenv("APPDATA")
		if baseDir == "" {
			userProfile := os.Getenv("USERPROFILE")
			if userProfile == "" {
				// Last resort fallback
				homeDir, _ := os.UserHomeDir()
				userProfile = homeDir
			}
			baseDir = filepath.Join(userProfile, "AppData", "Roaming")
		}
		return filepath.Join(baseDir, "ghex", "backup", "ghex.exe.backup")
	}

	// Unix-like systems
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, ".ghex", "backup", "ghex.backup")
}


// Backup creates a backup of the current binary
func (m *BinaryManager) Backup() error {
	// Ensure backup directory exists
	backupDir := filepath.Dir(m.BackupPath)
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return fmt.Errorf("%w: %v", ErrBackupFailed, err)
	}

	// Copy current binary to backup location
	if err := copyFile(m.BinaryPath, m.BackupPath); err != nil {
		return fmt.Errorf("%w: %v", ErrBackupFailed, err)
	}

	return nil
}

// Restore restores the binary from backup
func (m *BinaryManager) Restore() error {
	if !m.HasBackup() {
		return ErrNoBackupAvailable
	}

	// Copy backup to binary location
	if err := copyFile(m.BackupPath, m.BinaryPath); err != nil {
		return fmt.Errorf("%w: %v", ErrRestoreFailed, err)
	}

	// Set executable permissions on Unix
	if runtime.GOOS != "windows" {
		if err := SetExecutable(m.BinaryPath); err != nil {
			return fmt.Errorf("%w: failed to set permissions: %v", ErrRestoreFailed, err)
		}
	}

	return nil
}

// HasBackup checks if a backup exists
func (m *BinaryManager) HasBackup() bool {
	_, err := os.Stat(m.BackupPath)
	return err == nil
}

// Replace replaces the current binary with a new one
func (m *BinaryManager) Replace(newBinaryPath string) error {
	// On Windows, we need special handling for running executables
	if runtime.GOOS == "windows" {
		return m.replaceWindows(newBinaryPath)
	}

	return m.replaceUnix(newBinaryPath)
}

// replaceUnix handles binary replacement on Unix systems
// Linux doesn't allow writing to a running executable ("text file busy" error)
// Solution: remove the old file first, then copy the new one
func (m *BinaryManager) replaceUnix(newBinaryPath string) error {
	// Remove the current binary first (this is allowed even while running)
	if err := os.Remove(m.BinaryPath); err != nil {
		return fmt.Errorf("failed to remove current binary: %w", err)
	}

	// Copy new binary to destination
	if err := copyFile(newBinaryPath, m.BinaryPath); err != nil {
		return fmt.Errorf("failed to replace binary: %w", err)
	}

	// Set executable permissions
	if err := SetExecutable(m.BinaryPath); err != nil {
		return fmt.Errorf("failed to set permissions: %w", err)
	}

	return nil
}


// replaceWindows handles binary replacement on Windows
// Windows doesn't allow replacing a running executable directly
// We use a batch script that runs after this process exits
func (m *BinaryManager) replaceWindows(newBinaryPath string) error {
	// Write a batch script to replace the binary after this process exits
	// Use a unique script name to avoid conflicts
	scriptPath := m.BinaryPath + "_update.bat"
	batchContent := fmt.Sprintf(`@echo off
setlocal
set "NEW_BINARY=%s"
set "TARGET=%s"
set "SCRIPT=%%~f0"

:wait_loop
timeout /t 1 /nobreak > nul
move /y "%%NEW_BINARY%%" "%%TARGET%%" > nul 2>&1
if errorlevel 1 goto wait_loop

del "%%SCRIPT%%"
endlocal
`, newBinaryPath, m.BinaryPath)

	if err := os.WriteFile(scriptPath, []byte(batchContent), 0755); err != nil {
		return fmt.Errorf("failed to write update script: %w", err)
	}

	// Launch the script detached so it runs after we exit
	cmd := exec.Command("cmd", "/c", "start", "/b", "", scriptPath)
	if err := cmd.Start(); err != nil {
		os.Remove(scriptPath)
		return fmt.Errorf("failed to launch update script: %w", err)
	}

	return nil
}

// SetExecutable sets executable permissions on Unix systems
func SetExecutable(filePath string) error {
	if runtime.GOOS == "windows" {
		return nil // Windows doesn't need this
	}

	return os.Chmod(filePath, 0755)
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	// Get source file info for permissions
	srcInfo, err := srcFile.Stat()
	if err != nil {
		return err
	}

	// Create destination file
	dstFile, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, srcInfo.Mode())
	if err != nil {
		return err
	}
	defer dstFile.Close()

	// Copy content
	_, err = io.Copy(dstFile, srcFile)
	return err
}

// CheckWritePermission checks if we can write to the binary location
func CheckWritePermission(path string) error {
	dir := filepath.Dir(path)

	// Try to create a temp file in the directory
	f, err := os.CreateTemp(dir, ".ghex_permission_check_*")
	if err != nil {
		return fmt.Errorf("%w: cannot write to %s", ErrPermissionDenied, dir)
	}
	f.Close()
	os.Remove(f.Name())

	return nil
}

// GetBackupInfo returns information about the backup
func (m *BinaryManager) GetBackupInfo() (os.FileInfo, error) {
	if !m.HasBackup() {
		return nil, ErrNoBackupAvailable
	}
	return os.Stat(m.BackupPath)
}
