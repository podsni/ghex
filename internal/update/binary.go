package update

import (
	"fmt"
	"io"
	"os"
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
	var baseDir string

	if runtime.GOOS == "windows" {
		baseDir = os.Getenv("APPDATA")
		if baseDir == "" {
			baseDir = filepath.Join(os.Getenv("USERPROFILE"), "AppData", "Roaming")
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
func (m *BinaryManager) replaceUnix(newBinaryPath string) error {
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
func (m *BinaryManager) replaceWindows(newBinaryPath string) error {
	// Rename current binary to .old
	oldPath := m.BinaryPath + ".old"
	if err := os.Rename(m.BinaryPath, oldPath); err != nil {
		return fmt.Errorf("failed to rename current binary: %w", err)
	}

	// Copy new binary to destination
	if err := copyFile(newBinaryPath, m.BinaryPath); err != nil {
		// Try to restore old binary
		os.Rename(oldPath, m.BinaryPath)
		return fmt.Errorf("failed to copy new binary: %w", err)
	}

	// Remove old binary (may fail if still in use, that's ok)
	os.Remove(oldPath)

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
	tmpFile := filepath.Join(dir, ".ghex_permission_check")
	f, err := os.Create(tmpFile)
	if err != nil {
		return fmt.Errorf("%w: cannot write to %s", ErrPermissionDenied, dir)
	}
	f.Close()
	os.Remove(tmpFile)

	return nil
}

// GetBackupInfo returns information about the backup
func (m *BinaryManager) GetBackupInfo() (os.FileInfo, error) {
	if !m.HasBackup() {
		return nil, ErrNoBackupAvailable
	}
	return os.Stat(m.BackupPath)
}
