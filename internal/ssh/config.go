package ssh

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/dwirx/ghex/internal/platform"
)

// GetSSHConfigPath returns the path to SSH config file
func GetSSHConfigPath() string {
	return filepath.Join(platform.GetSSHDir(), "config")
}

// EnsureConfigBlock ensures an SSH Host block exists in the config file
// If the block already exists, it updates it; otherwise, it appends a new block
func EnsureConfigBlock(alias, keyPath, hostname string) error {
	if hostname == "" {
		hostname = "github.com"
	}

	sshDir := platform.GetSSHDir()
	configPath := GetSSHConfigPath()

	// Ensure SSH directory exists with proper permissions
	if err := platform.EnsureDir(sshDir, 0700); err != nil {
		return fmt.Errorf("failed to create SSH directory: %w", err)
	}

	// Read existing config
	var content string
	data, err := os.ReadFile(configPath)
	if err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("failed to read SSH config: %w", err)
		}
		content = ""
	} else {
		content = string(data)
		// Normalize line endings
		content = strings.ReplaceAll(content, "\r\n", "\n")
		content = strings.ReplaceAll(content, "\r", "\n")
	}

	// Build the new Host block
	block := buildHostBlock(alias, keyPath, hostname)

	// Check if Host block already exists
	if containsHostBlock(content, alias) {
		// Update existing block
		content = updateHostBlock(content, alias, block)
	} else {
		// Append new block
		if content != "" && !strings.HasSuffix(content, "\n") {
			content += "\n"
		}
		if content != "" {
			content += "\n"
		}
		content += block + "\n"
	}

	// Write config file
	if err := os.WriteFile(configPath, []byte(content), 0600); err != nil {
		return fmt.Errorf("failed to write SSH config: %w", err)
	}

	return nil
}

// buildHostBlock creates an SSH Host block string
func buildHostBlock(alias, keyPath, hostname string) string {
	// Normalize path separators for SSH config (always use forward slashes)
	keyPath = strings.ReplaceAll(keyPath, "\\", "/")
	return fmt.Sprintf(`Host %s
  HostName %s
  User git
  IdentityFile %s
  IdentitiesOnly yes`, alias, hostname, keyPath)
}

// containsHostBlock checks if a Host block exists in the config
func containsHostBlock(content, alias string) bool {
	pattern := fmt.Sprintf(`(?m)^Host\s+%s\s*$`, regexp.QuoteMeta(alias))
	matched, _ := regexp.MatchString(pattern, content)
	return matched
}

// updateHostBlock replaces an existing Host block with a new one
func updateHostBlock(content, alias, newBlock string) string {
	// Normalize line endings
	content = strings.ReplaceAll(content, "\r\n", "\n")
	content = strings.ReplaceAll(content, "\r", "\n")
	lines := strings.Split(content, "\n")
	var result []string
	inBlock := false
	hostPattern := regexp.MustCompile(`^Host\s+`)
	targetPattern := regexp.MustCompile(fmt.Sprintf(`^Host\s+%s\s*$`, regexp.QuoteMeta(alias)))

	for i := 0; i < len(lines); i++ {
		line := lines[i]

		if targetPattern.MatchString(line) {
			// Found the target block, replace it
			result = append(result, newBlock)
			inBlock = true
			continue
		}

		if inBlock {
			// Check if we've reached the next Host block or end of block
			if hostPattern.MatchString(line) {
				inBlock = false
				result = append(result, line)
			}
			// Skip lines in the old block
			continue
		}

		result = append(result, line)
	}

	return strings.Join(result, "\n")
}

// RemoveHostBlock removes a Host block from the SSH config
func RemoveHostBlock(alias string) error {
	configPath := GetSSHConfigPath()

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // Nothing to remove
		}
		return fmt.Errorf("failed to read SSH config: %w", err)
	}

	content := string(data)
	// Normalize line endings
	content = strings.ReplaceAll(content, "\r\n", "\n")
	content = strings.ReplaceAll(content, "\r", "\n")
	if !containsHostBlock(content, alias) {
		return nil // Block doesn't exist
	}

	// Remove the block
	lines := strings.Split(content, "\n")
	var result []string
	inBlock := false
	hostPattern := regexp.MustCompile(`^Host\s+`)
	targetPattern := regexp.MustCompile(fmt.Sprintf(`^Host\s+%s\s*$`, regexp.QuoteMeta(alias)))

	for _, line := range lines {
		if targetPattern.MatchString(line) {
			inBlock = true
			continue
		}

		if inBlock {
			if hostPattern.MatchString(line) {
				inBlock = false
				result = append(result, line)
			}
			continue
		}

		result = append(result, line)
	}

	// Clean up extra newlines
	newContent := strings.TrimSpace(strings.Join(result, "\n")) + "\n"

	return os.WriteFile(configPath, []byte(newContent), 0600)
}

// GetHostBlock retrieves a Host block from the SSH config
func GetHostBlock(alias string) (string, error) {
	configPath := GetSSHConfigPath()

	data, err := os.ReadFile(configPath)
	if err != nil {
		return "", err
	}

	content := string(data)
	// Normalize line endings
	content = strings.ReplaceAll(content, "\r\n", "\n")
	content = strings.ReplaceAll(content, "\r", "\n")
	lines := strings.Split(content, "\n")
	var blockLines []string
	inBlock := false
	hostPattern := regexp.MustCompile(`^Host\s+`)
	targetPattern := regexp.MustCompile(fmt.Sprintf(`^Host\s+%s\s*$`, regexp.QuoteMeta(alias)))

	for _, line := range lines {
		if targetPattern.MatchString(line) {
			inBlock = true
			blockLines = append(blockLines, line)
			continue
		}

		if inBlock {
			if hostPattern.MatchString(line) {
				break
			}
			blockLines = append(blockLines, line)
		}
	}

	if len(blockLines) == 0 {
		return "", fmt.Errorf("host block not found: %s", alias)
	}

	return strings.Join(blockLines, "\n"), nil
}
