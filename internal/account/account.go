package account

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/dwirx/ghex/internal/config"
	"github.com/dwirx/ghex/internal/git"
	"github.com/dwirx/ghex/internal/platform"
	"github.com/dwirx/ghex/internal/ssh"
)

// Manager handles account operations
type Manager struct {
	cfg *config.AppConfig
}

// NewManager creates a new account manager
func NewManager(cfg *config.AppConfig) *Manager {
	return &Manager{cfg: cfg}
}

// Add adds a new account to the configuration
func (m *Manager) Add(account config.Account) error {
	// Check for duplicate name
	for _, a := range m.cfg.Accounts {
		if strings.EqualFold(a.Name, account.Name) {
			return fmt.Errorf("account with name '%s' already exists", account.Name)
		}
	}

	m.cfg.Accounts = append(m.cfg.Accounts, account)
	return nil
}

// Remove removes an account by name
func (m *Manager) Remove(name string) error {
	for i, a := range m.cfg.Accounts {
		if strings.EqualFold(a.Name, name) {
			m.cfg.Accounts = append(m.cfg.Accounts[:i], m.cfg.Accounts[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("account '%s' not found", name)
}

// Find finds an account by name
func (m *Manager) Find(name string) *config.Account {
	for i, a := range m.cfg.Accounts {
		if strings.EqualFold(a.Name, name) {
			return &m.cfg.Accounts[i]
		}
	}
	return nil
}

// List returns all accounts
func (m *Manager) List() []config.Account {
	return m.cfg.Accounts
}

// Update updates an existing account
func (m *Manager) Update(name string, updates config.Account) error {
	for i, a := range m.cfg.Accounts {
		if strings.EqualFold(a.Name, name) {
			m.cfg.Accounts[i] = updates
			return nil
		}
	}
	return fmt.Errorf("account '%s' not found", name)
}

// SwitchMethod represents the authentication method to use
type SwitchMethod string

const (
	MethodSSH   SwitchMethod = "ssh"
	MethodToken SwitchMethod = "token"
)

// Switch switches the current repository to use a specific account
func (m *Manager) Switch(accountName string, method SwitchMethod, repoPath string) error {
	account := m.Find(accountName)
	if account == nil {
		return fmt.Errorf("account '%s' not found", accountName)
	}

	if repoPath == "" {
		repoPath = "."
	}

	// Get current remote URL to extract owner/repo
	remoteURL, err := git.GetRemoteURL("origin", repoPath)
	if err != nil {
		return fmt.Errorf("failed to get remote URL: %w", err)
	}

	owner, repo, err := git.ParseRepoFromURL(remoteURL)
	if err != nil {
		return fmt.Errorf("failed to parse remote URL: %w", err)
	}

	repoFullPath := fmt.Sprintf("%s/%s", owner, repo)

	// Get platform info
	platformType := "github"
	domain := ""
	if account.Platform != nil {
		platformType = account.Platform.Type
		domain = account.Platform.Domain
	}

	switch method {
	case MethodSSH:
		if account.SSH == nil {
			return fmt.Errorf("account '%s' has no SSH configuration", accountName)
		}

		// Ensure SSH key permissions
		keyPath := platform.ExpandPath(account.SSH.KeyPath)
		if _, err := os.Stat(keyPath); os.IsNotExist(err) {
			return fmt.Errorf("SSH key not found at path: %s", keyPath)
		}
		if err := ssh.SetKeyPermissions(keyPath); err != nil {
			return fmt.Errorf("failed to set SSH key permissions: %w", err)
		}

		// Configure SSH host
		sshHost := git.GetPlatformSSHHost(platformType, domain)
		if err := ssh.EnsureConfigBlock(sshHost, keyPath, sshHost); err != nil {
			return fmt.Errorf("failed to configure SSH: %w", err)
		}

		// Set remote URL to SSH format
		newURL := git.BuildRemoteURL(platformType, domain, repoFullPath, true)
		if err := git.SetRemoteURL(newURL, "origin", repoPath); err != nil {
			return fmt.Errorf("failed to set remote URL: %w", err)
		}

	case MethodToken:
		if account.Token == nil {
			return fmt.Errorf("account '%s' has no token configuration", accountName)
		}

		// Set up credential store
		if err := git.EnsureCredentialStore(); err != nil {
			return fmt.Errorf("failed to set up credential store: %w", err)
		}

		// Write credentials
		host := git.GetPlatformSSHHost(platformType, domain)
		if err := git.WriteCredentials(account.Token.Username, account.Token.Token, host); err != nil {
			return fmt.Errorf("failed to write credentials: %w", err)
		}

		// Set remote URL to HTTPS format
		newURL := git.BuildRemoteURL(platformType, domain, repoFullPath, false)
		if err := git.SetRemoteURL(newURL, "origin", repoPath); err != nil {
			return fmt.Errorf("failed to set remote URL: %w", err)
		}

	default:
		return fmt.Errorf("unknown method: %s", method)
	}

	// Set local git identity
	if err := git.SetLocalIdentity(account.GitUserName, account.GitEmail, repoPath); err != nil {
		return fmt.Errorf("failed to set git identity: %w", err)
	}

	// Log activity
	m.LogActivity(config.ActivityLogEntry{
		Timestamp:   time.Now().UTC().Format(time.RFC3339),
		Action:      "switch",
		AccountName: accountName,
		RepoPath:    repoFullPath,
		Method:      string(method),
		Platform:    platformType,
		Success:     true,
	})

	return nil
}

// LogActivity adds an activity log entry
func (m *Manager) LogActivity(entry config.ActivityLogEntry) {
	if entry.Timestamp == "" {
		entry.Timestamp = time.Now().UTC().Format(time.RFC3339)
	}
	m.cfg.ActivityLog = append(m.cfg.ActivityLog, entry)
}

// GetRecentActivity returns the most recent activity entries
func (m *Manager) GetRecentActivity(limit int) []config.ActivityLogEntry {
	if limit <= 0 || limit > len(m.cfg.ActivityLog) {
		limit = len(m.cfg.ActivityLog)
	}

	// Return most recent entries (from the end)
	start := len(m.cfg.ActivityLog) - limit
	if start < 0 {
		start = 0
	}

	result := make([]config.ActivityLogEntry, limit)
	copy(result, m.cfg.ActivityLog[start:])

	// Reverse to get most recent first
	for i, j := 0, len(result)-1; i < j; i, j = i+1, j-1 {
		result[i], result[j] = result[j], result[i]
	}

	return result
}
