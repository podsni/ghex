package ssh

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/dwirx/ghex/internal/platform"
	"github.com/dwirx/ghex/internal/shell"
)

// GenerateKey generates a new Ed25519 SSH key pair
func GenerateKey(keyPath, comment string) error {
	// Expand path
	keyPath = platform.ExpandPath(keyPath)

	// Ensure directory exists
	dir := filepath.Dir(keyPath)
	if err := platform.EnsureDir(dir, 0700); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Generate key using ssh-keygen
	// Use ToSSHPath to convert Windows backslashes to forward slashes for SSH compatibility
	args := []string{
		"-t", "ed25519",
		"-f", platform.ToSSHPath(keyPath),
		"-C", comment,
		"-N", "", // Empty passphrase
		"-q",    // Quiet mode to prevent interactive prompts
	}

	_, err := shell.Run("ssh-keygen", args...)
	if err != nil {
		return fmt.Errorf("failed to generate SSH key: %w", err)
	}

	// Set permissions
	if err := SetKeyPermissions(keyPath); err != nil {
		return fmt.Errorf("failed to set key permissions: %w", err)
	}

	return nil
}

// ImportKey copies an SSH private key to a new location
func ImportKey(srcPath, destPath string) error {
	srcPath = platform.ExpandPath(srcPath)
	destPath = platform.ExpandPath(destPath)

	// Check source exists
	if !platform.FileExists(srcPath) {
		return fmt.Errorf("source key not found: %s", srcPath)
	}

	// Ensure destination directory exists
	dir := filepath.Dir(destPath)
	if err := platform.EnsureDir(dir, 0700); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Copy file
	src, err := os.Open(srcPath)
	if err != nil {
		return fmt.Errorf("failed to open source key: %w", err)
	}
	defer src.Close()

	dst, err := os.OpenFile(destPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("failed to create destination key: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return fmt.Errorf("failed to copy key: %w", err)
	}

	// Set permissions
	return SetKeyPermissions(destPath)
}

// EnsurePublicKey generates a public key from a private key if it doesn't exist
func EnsurePublicKey(privateKeyPath string) (string, error) {
	privateKeyPath = platform.ExpandPath(privateKeyPath)
	pubPath := privateKeyPath + ".pub"

	// Check if public key already exists
	if platform.FileExists(pubPath) {
		return pubPath, nil
	}

	// Generate public key from private key
	// Use ToSSHPath to convert Windows backslashes to forward slashes for SSH compatibility
	output, err := shell.Run("ssh-keygen", "-y", "-f", platform.ToSSHPath(privateKeyPath))
	if err != nil {
		return "", fmt.Errorf("failed to generate public key: %w", err)
	}

	// Write public key
	pubKey := strings.TrimSpace(output) + "\n"
	if err := os.WriteFile(pubPath, []byte(pubKey), 0644); err != nil {
		return "", fmt.Errorf("failed to write public key: %w", err)
	}

	return pubPath, nil
}

// SetKeyPermissions sets proper permissions on SSH key files
func SetKeyPermissions(keyPath string) error {
	keyPath = platform.ExpandPath(keyPath)

	// Set private key permissions (600)
	if err := os.Chmod(keyPath, 0600); err != nil {
		// On Windows, chmod might not work as expected
		if !platform.IsWindows() {
			return fmt.Errorf("failed to set private key permissions: %w", err)
		}
	}

	// Set public key permissions (644) if exists
	pubPath := keyPath + ".pub"
	if platform.FileExists(pubPath) {
		if err := os.Chmod(pubPath, 0644); err != nil {
			if !platform.IsWindows() {
				return fmt.Errorf("failed to set public key permissions: %w", err)
			}
		}
	}

	return nil
}

// EnsureKeyPermissions checks and fixes SSH key permissions
// Returns true if permissions were fixed, false if already correct
func EnsureKeyPermissions(keyPath string) (bool, error) {
	keyPath = platform.ExpandPath(keyPath)

	// Check if file exists
	info, err := os.Stat(keyPath)
	if err != nil {
		return false, fmt.Errorf("cannot access key: %w", err)
	}

	// Get current permissions
	currentMode := info.Mode().Perm()

	// Check if permissions need fixing (should be 0600 or 0400)
	needsFix := currentMode != 0600 && currentMode != 0400

	if needsFix {
		// Fix permissions
		if err := os.Chmod(keyPath, 0600); err != nil {
			if !platform.IsWindows() {
				return false, fmt.Errorf("failed to fix permissions: %w", err)
			}
		}
		return true, nil
	}

	return false, nil
}

// TestConnection tests SSH connection to a host (uses default SSH key)
func TestConnection(host string) (bool, string, error) {
	return TestConnectionWithKey(host, "")
}

// FixAllKeyPermissions fixes permissions for all SSH keys in ~/.ssh directory
func FixAllKeyPermissions() (int, error) {
	keys, err := ListPrivateKeys()
	if err != nil {
		return 0, err
	}

	fixed := 0
	for _, key := range keys {
		// Use ForceFixKeyPermissions for more reliable permission fixing
		if ForceFixKeyPermissions(key) {
			fixed++
		}
	}

	return fixed, nil
}

// ForceFixKeyPermissions uses chmod command to fix permissions (more reliable)
func ForceFixKeyPermissions(keyPath string) bool {
	keyPath = platform.ExpandPath(keyPath)

	// Check if file exists
	if _, err := os.Stat(keyPath); err != nil {
		return false
	}

	// Use chmod command directly for more reliable permission fixing
	if !platform.IsWindows() {
		// Fix private key permissions (600)
		_, _ = shell.Exec("chmod", "600", keyPath)

		// Fix public key permissions (644) if exists
		pubPath := keyPath + ".pub"
		if platform.FileExists(pubPath) {
			_, _ = shell.Exec("chmod", "644", pubPath)
		}
		return true
	}

	// On Windows, use os.Chmod (may not work perfectly but try anyway)
	_ = os.Chmod(keyPath, 0600)
	pubPath := keyPath + ".pub"
	if platform.FileExists(pubPath) {
		_ = os.Chmod(pubPath, 0644)
	}
	return true
}

// isGitBash returns true if running in Git Bash / MSYS2 on Windows
func isGitBash() bool {
	return os.Getenv("MSYSTEM") != "" || (platform.IsWindows() && os.Getenv("SHELL") != "")
}

// EnsureSSHDirPermissions ensures ~/.ssh directory and config have correct permissions
func EnsureSSHDirPermissions() {
	sshDir := platform.GetSSHDir()

	if !platform.IsWindows() {
		// Fix SSH directory permissions (700)
		_, _ = shell.Exec("chmod", "700", sshDir)

		// Fix SSH config permissions (600) if exists
		configPath := filepath.Join(sshDir, "config")
		if platform.FileExists(configPath) {
			_, _ = shell.Exec("chmod", "600", configPath)
		}
	}
}

// TestConnectionWithKey tests SSH connection to a host using a specific SSH key
func TestConnectionWithKey(host, keyPath string) (bool, string, error) {
	if host == "" {
		host = "github.com"
	}

	// First, fix permissions for ALL SSH keys to avoid "bad permissions" errors
	// This is critical because SSH will scan all keys and fail if any has bad permissions
	FixAllKeyPermissions()

	// Also ensure SSH directory and config have correct permissions
	EnsureSSHDirPermissions()

	args := []string{
		"-T",
		"-o", "StrictHostKeyChecking=no",
		"-o", "ConnectTimeout=10",
		"-o", "BatchMode=yes",
		"-o", "LogLevel=ERROR", // Suppress warnings
	}

	// If keyPath is provided, use it exclusively
	if keyPath != "" {
		keyPath = platform.ExpandPath(keyPath)

		// Force fix permissions using chmod command (more reliable than os.Chmod)
		ForceFixKeyPermissions(keyPath)

		// IMPORTANT: These options ensure ONLY the specified key is used
		// -F /dev/null - Ignore SSH config file completely (Linux/Mac/Git Bash)
		// -F NUL - Ignore SSH config file completely (Windows cmd/PowerShell)
		// IdentitiesOnly=yes - Only use identities specified on command line
		// IdentityAgent=none - Disable ssh-agent to prevent using other keys
		nullDevice := "/dev/null"
		if platform.IsWindows() && !isGitBash() {
			nullDevice = "NUL"
		}
		args = append(args, "-F", nullDevice)
		args = append(args, "-o", "IdentitiesOnly=yes")
		args = append(args, "-o", "IdentityAgent=none") // Disable ssh-agent
		// Use ToSSHPath to convert Windows backslashes to forward slashes for SSH compatibility
		args = append(args, "-i", platform.ToSSHPath(keyPath))
	}

	args = append(args, fmt.Sprintf("git@%s", host))

	output, err := shell.Exec("ssh", args...)

	// SSH -T returns exit code 1 for successful auth on GitHub/GitLab/Gitea
	// Check output for success patterns
	successPatterns := []string{
		// GitHub patterns
		"successfully authenticated",
		"Hi .+! You've successfully authenticated",
		// GitLab patterns
		"Welcome to GitLab",
		// Bitbucket patterns
		"logged in as",
		"authenticated via",
		// Gitea patterns
		"Hi there,",
		"You've successfully authenticated",
		"Welcome to Gitea",
		"You can use git",
		// Codeberg (Gitea-based)
		"Welcome to Codeberg",
		// Generic patterns
		"successfully authenticated",
		"authentication succeeded",
	}

	for _, pattern := range successPatterns {
		matched, _ := regexp.MatchString("(?i)"+pattern, output)
		if matched {
			// Extract username if possible
			// Patterns for different platforms:
			// GitHub: "Hi username! You've successfully authenticated"
			// GitLab: "Welcome to GitLab, @username!"
			// Gitea: "Hi there, username! You've successfully authenticated"
			// Bitbucket: "logged in as username"
			userPatterns := []string{
				`Hi\s+([^!,]+)[!,]`,           // GitHub, Gitea
				`Hi there,?\s+([^!]+)!`,       // Gitea alternative
				`logged in as\s+(\S+)`,        // Bitbucket
				`@(\S+)`,                      // GitLab
				`Welcome to \w+,?\s*@?(\S+)!`, // Generic welcome
			}
			for _, userPattern := range userPatterns {
				userRe := regexp.MustCompile(userPattern)
				if matches := userRe.FindStringSubmatch(output); len(matches) > 1 {
					username := strings.TrimSpace(matches[1])
					if username != "" {
						return true, fmt.Sprintf("Successfully authenticated as %s", username), nil
					}
				}
			}
			return true, "Successfully authenticated", nil
		}
	}

	if err != nil {
		return false, output, err
	}

	return false, output, nil
}

// ListPrivateKeys returns a list of SSH private keys in the SSH directory
func ListPrivateKeys() ([]string, error) {
	sshDir := platform.GetSSHDir()

	entries, err := os.ReadDir(sshDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, err
	}

	excludeFiles := map[string]bool{
		"known_hosts":      true,
		"known_hosts.old":  true,
		"config":           true,
		"authorized_keys":  true,
		"authorized_keys2": true,
	}

	var keys []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()

		// Skip excluded files
		if excludeFiles[name] {
			continue
		}

		// Skip public keys
		if strings.HasSuffix(name, ".pub") {
			continue
		}

		// Skip PuTTY key files
		if strings.HasSuffix(name, ".ppk") {
			continue
		}

		keys = append(keys, filepath.Join(sshDir, name))
	}

	return keys, nil
}

// SuggestKeyFilenames suggests destination filenames for SSH keys
func SuggestKeyFilenames(username, label string) []string {
	base := username
	if base == "" {
		base = label
	}
	if base == "" {
		base = "github"
	}

	// Clean the base name
	base = strings.ToLower(base)
	base = regexp.MustCompile(`[^a-zA-Z0-9_-]+`).ReplaceAllString(base, "")

	candidates := []string{
		fmt.Sprintf("id_ed25519_%s", base),
		fmt.Sprintf("id_ecdsa_%s", base),
		fmt.Sprintf("id_rsa_%s", base),
		"id_ed25519_github",
	}

	// Remove duplicates
	seen := make(map[string]bool)
	var unique []string
	for _, c := range candidates {
		if !seen[c] {
			seen[c] = true
			unique = append(unique, c)
		}
	}

	return unique
}
