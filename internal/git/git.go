package git

import (
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/dwirx/ghex/internal/platform"
	"github.com/dwirx/ghex/internal/shell"
)

// IsGitRepo checks if the given path is inside a git repository
func IsGitRepo(path string) bool {
	_, err := shell.RunInDir(path, "git", "rev-parse", "--is-inside-work-tree")
	return err == nil
}

// GetGitRoot returns the root directory of the git repository
func GetGitRoot(path string) (string, error) {
	return shell.RunInDir(path, "git", "rev-parse", "--show-toplevel")
}

// GetRemoteURL returns the URL of a remote
func GetRemoteURL(remote, path string) (string, error) {
	if remote == "" {
		remote = "origin"
	}
	if path == "" {
		path = "."
	}

	return shell.RunInDir(path, "git", "remote", "get-url", remote)
}

// SetRemoteURL sets the URL of a remote
func SetRemoteURL(remoteURL, remote, path string) error {
	if remote == "" {
		remote = "origin"
	}
	if path == "" {
		path = "."
	}

	_, err := shell.RunInDir(path, "git", "remote", "set-url", remote, remoteURL)
	return err
}

// SetLocalIdentity sets the local git user.name and user.email
func SetLocalIdentity(name, email, path string) error {
	if path == "" {
		path = "."
	}

	if name != "" {
		if _, err := shell.RunInDir(path, "git", "config", "user.name", name); err != nil {
			return fmt.Errorf("failed to set user.name: %w", err)
		}
	}

	if email != "" {
		if _, err := shell.RunInDir(path, "git", "config", "user.email", email); err != nil {
			return fmt.Errorf("failed to set user.email: %w", err)
		}
	}

	return nil
}

// SetGlobalIdentity sets the global git user.name and user.email
func SetGlobalIdentity(name, email string) error {
	if name != "" {
		if _, err := shell.Run("git", "config", "--global", "user.name", name); err != nil {
			return fmt.Errorf("failed to set global user.name: %w", err)
		}
	}

	if email != "" {
		if _, err := shell.Run("git", "config", "--global", "user.email", email); err != nil {
			return fmt.Errorf("failed to set global user.email: %w", err)
		}
	}

	return nil
}

// GetCurrentUser returns the current git user.name and user.email
func GetCurrentUser(path string) (name, email string, err error) {
	if path == "" {
		path = "."
	}

	name, _ = shell.RunInDir(path, "git", "config", "user.name")
	email, _ = shell.RunInDir(path, "git", "config", "user.email")

	return name, email, nil
}

// GetCurrentBranch returns the current branch name
func GetCurrentBranch(path string) (string, error) {
	if path == "" {
		path = "."
	}

	return shell.RunInDir(path, "git", "branch", "--show-current")
}

// EnsureCredentialStore sets up git credential store
func EnsureCredentialStore() error {
	_, err := shell.Run("git", "config", "credential.helper", "store")
	return err
}

// WriteCredentials writes credentials to ~/.git-credentials
func WriteCredentials(username, token, host string) error {
	if host == "" {
		host = "github.com"
	}

	credPath := platform.GetGitCredentialsPath()

	// Read existing credentials
	var existing string
	data, err := os.ReadFile(credPath)
	if err == nil {
		existing = string(data)
	}

	// Filter out existing credentials for this host
	lines := strings.Split(existing, "\n")
	var filtered []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if !strings.Contains(line, "@"+host) {
			filtered = append(filtered, line)
		}
	}

	// Add new credential
	encodedUser := url.QueryEscape(username)
	encodedToken := url.QueryEscape(token)
	newCred := fmt.Sprintf("https://%s:%s@%s", encodedUser, encodedToken, host)
	filtered = append(filtered, newCred)

	// Write back
	content := strings.Join(filtered, "\n") + "\n"
	return os.WriteFile(credPath, []byte(content), 0600)
}

// TestTokenAuth tests token authentication against GitHub API
func TestTokenAuth(username, token string) (bool, string, error) {
	// Use curl to test authentication
	output, err := shell.Exec("curl",
		"-s",
		"-o", "/dev/null",
		"-w", "%{http_code}",
		"-u", fmt.Sprintf("%s:%s", username, token),
		"https://api.github.com/user",
	)

	code := strings.TrimSpace(output)
	if code == "200" {
		return true, "HTTP 200 OK", nil
	}

	return false, fmt.Sprintf("HTTP %s", code), err
}

// GetConfigList returns all git configuration
func GetConfigList() (string, error) {
	return shell.Run("git", "config", "--list")
}
