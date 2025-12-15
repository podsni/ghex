package account

import (
	"strings"

	"github.com/dwirx/ghex/internal/config"
	"github.com/dwirx/ghex/internal/git"
)

// DetectActive detects the currently active account for a repository
func (m *Manager) DetectActive(repoPath string) (string, error) {
	if repoPath == "" {
		repoPath = "."
	}

	// Check if we're in a git repository
	if !git.IsGitRepo(repoPath) {
		return "", nil
	}

	// Get current git user and remote info
	userName, userEmail, _ := git.GetCurrentUser(repoPath)
	remoteURL, _ := git.GetRemoteURL("origin", repoPath)

	if userName == "" && userEmail == "" && remoteURL == "" {
		return "", nil
	}

	// Determine auth type from remote URL
	isSSH := strings.HasPrefix(remoteURL, "git@") || strings.HasPrefix(remoteURL, "ssh://")

	// Try to match account based on git identity and remote URL
	for _, account := range m.cfg.Accounts {
		matches := 0
		totalChecks := 0

		// Check git identity match
		if account.GitUserName != "" {
			totalChecks++
			if userName == account.GitUserName {
				matches++
			}
		}

		if account.GitEmail != "" {
			totalChecks++
			if userEmail == account.GitEmail {
				matches++
			}
		}

		// Check remote URL type match (SSH vs HTTPS)
		totalChecks++
		if isSSH && account.SSH != nil {
			matches++
		} else if !isSSH && account.Token != nil {
			matches++
		}

		// If we have matches and they represent a significant portion
		if matches > 0 && totalChecks > 0 {
			threshold := (totalChecks + 1) / 2 // Ceiling division
			if matches >= threshold {
				return account.Name, nil
			}
		}
	}

	return "", nil
}

// DetectActiveAccount is a convenience function
func DetectActiveAccount(cfg *config.AppConfig, repoPath string) (string, error) {
	manager := NewManager(cfg)
	return manager.DetectActive(repoPath)
}

// GetRemoteInfo returns information about the current repository's remote
type RemoteInfo struct {
	RemoteURL string
	RepoPath  string
	AuthType  string // "ssh" or "https"
	Platform  string
	Owner     string
	Repo      string
}

// GetRemoteInfo gets information about the current repository's remote
func GetRemoteInfo(repoPath string) (*RemoteInfo, error) {
	if repoPath == "" {
		repoPath = "."
	}

	remoteURL, err := git.GetRemoteURL("origin", repoPath)
	if err != nil {
		return nil, err
	}

	owner, repo, err := git.ParseRepoFromURL(remoteURL)
	if err != nil {
		return nil, err
	}

	isSSH := strings.HasPrefix(remoteURL, "git@") || strings.HasPrefix(remoteURL, "ssh://")
	authType := "https"
	if isSSH {
		authType = "ssh"
	}

	// Detect platform
	urlInfo, _ := git.ParseURL(remoteURL)
	platform := "github"
	if urlInfo != nil {
		platform = urlInfo.Platform
	}

	return &RemoteInfo{
		RemoteURL: remoteURL,
		RepoPath:  owner + "/" + repo,
		AuthType:  authType,
		Platform:  platform,
		Owner:     owner,
		Repo:      repo,
	}, nil
}
