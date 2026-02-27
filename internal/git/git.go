package git

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/dwirx/ghex/internal/platform"
	"github.com/dwirx/ghex/internal/shell"
)

// IsGitRepo checks if the given path is inside a git repository
func IsGitRepo(path string) bool {
	_, err := shell.RunInDir(path, "git", "rev-parse", "--is-inside-work-tree")
	return err == nil
}

// convertMSYSPath converts MSYS/Git Bash paths like /c/Users/... to C:/Users/...
func convertMSYSPath(path string) string {
	if len(path) >= 2 && path[0] == '/' && path[1] != '/' {
		// Check if it looks like /c/... or /d/... (single letter drive)
		parts := strings.SplitN(path[1:], "/", 2)
		if len(parts[0]) == 1 {
			driveLetter := strings.ToUpper(parts[0])
			if len(parts) > 1 {
				return driveLetter + ":/" + parts[1]
			}
			return driveLetter + ":/"
		}
	}
	return path
}

// GetGitRoot returns the root directory of the git repository
func GetGitRoot(path string) (string, error) {
	result, err := shell.RunInDir(path, "git", "rev-parse", "--show-toplevel")
	if err != nil {
		return "", err
	}
	result = strings.TrimSpace(result)
	if runtime.GOOS == "windows" {
		result = convertMSYSPath(result)
	}
	return result, nil
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
		// Normalize line endings (handle Windows \r\n)
		existing = strings.ReplaceAll(existing, "\r\n", "\n")
		existing = strings.ReplaceAll(existing, "\r", "\n")
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
	return TestTokenAuthForHost(username, token, "github.com")
}

// TestTokenAuthForHost tests token authentication against a specific host's API
func TestTokenAuthForHost(username, token, host string) (bool, string, error) {
	// Build API URL based on host
	var apiURL string
	switch host {
	case "github.com":
		apiURL = "https://api.github.com/user"
	case "gitlab.com":
		apiURL = "https://gitlab.com/api/v4/user"
	default:
		// For self-hosted GitLab, Gitea, Codeberg, etc.
		// Try Gitea/Codeberg style API first (most common for self-hosted)
		apiURL = fmt.Sprintf("https://%s/api/v1/user", host)
	}

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return false, "", fmt.Errorf("failed to create request: %w", err)
	}

	req.SetBasicAuth(username, token)
	req.Header.Set("User-Agent", "ghex-cli")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return false, "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		return true, fmt.Sprintf("HTTP %d OK", resp.StatusCode), nil
	}

	return false, fmt.Sprintf("HTTP %d", resp.StatusCode), nil
}

// GetConfigList returns all git configuration
func GetConfigList() (string, error) {
	return shell.Run("git", "config", "--list")
}
