package git

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/dwirx/ghex/internal/shell"
)

// Clone clones a git repository
func Clone(repoURL string, targetDir string) (string, error) {
	normalized, _, err := NormalizeURL(repoURL)
	if err != nil {
		return "", fmt.Errorf("invalid git URL: %w", err)
	}

	args := []string{"clone", normalized}
	if targetDir != "" {
		args = append(args, targetDir)
	}

	_, err = shell.Run("git", args...)
	if err != nil {
		return "", fmt.Errorf("failed to clone repository: %w", err)
	}

	// Determine the cloned directory name
	if targetDir != "" {
		return targetDir, nil
	}

	// Extract directory name from URL
	_, repo, err := ParseRepoFromURL(normalized)
	if err != nil {
		return "", err
	}

	return strings.TrimSuffix(repo, ".git"), nil
}

// CloneWithIdentity clones a repository and sets up git identity
func CloneWithIdentity(repoURL, targetDir, userName, email string) (string, error) {
	clonedDir, err := Clone(repoURL, targetDir)
	if err != nil {
		return "", err
	}

	// Set local identity if provided
	if userName != "" || email != "" {
		if err := SetLocalIdentity(userName, email, clonedDir); err != nil {
			// Don't fail the clone, just warn
			fmt.Printf("Warning: failed to set git identity: %v\n", err)
		}
	}

	return clonedDir, nil
}

// CloneToPath clones a repository to a specific path
func CloneToPath(repoURL, basePath, targetName string) (string, error) {
	var targetDir string
	if targetName != "" {
		targetDir = filepath.Join(basePath, targetName)
	} else {
		_, repo, err := ParseRepoFromURL(repoURL)
		if err != nil {
			return "", err
		}
		targetDir = filepath.Join(basePath, strings.TrimSuffix(repo, ".git"))
	}

	return Clone(repoURL, targetDir)
}
