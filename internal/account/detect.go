package account

import (
	"strings"

	"github.com/dwirx/ghex/internal/config"
	"github.com/dwirx/ghex/internal/git"
)

// Scoring weights for active account detection
const (
	ScoreUserName     = 30 // Git user.name match
	ScoreUserEmail    = 30 // Git user.email match
	ScoreSSHKey       = 20 // SSH key in use
	ScorePlatform     = 20 // Platform match
	MinConfidenceScore = 30 // Minimum score to consider a match
)

// MatchScore represents how well an account matches current context
type MatchScore struct {
	AccountName   string
	Score         int
	MatchedFields []string
	IsActive      bool
}

// DetectActiveWithScore returns the best matching account with confidence score
func (m *Manager) DetectActiveWithScore(repoPath string) (*MatchScore, error) {
	if repoPath == "" {
		repoPath = "."
	}

	// Check if we're in a git repository
	if !git.IsGitRepo(repoPath) {
		return nil, nil
	}

	// Get current git user and remote info
	userName, userEmail, _ := git.GetCurrentUser(repoPath)
	remoteURL, _ := git.GetRemoteURL("origin", repoPath)

	if userName == "" && userEmail == "" && remoteURL == "" {
		return nil, nil
	}

	// Determine auth type and platform from remote URL
	isSSH := strings.HasPrefix(remoteURL, "git@") || strings.HasPrefix(remoteURL, "ssh://")
	detectedPlatform := DetectPlatformFromURL(remoteURL)

	var bestMatch *MatchScore

	for _, account := range m.cfg.Accounts {
		score := 0
		matchedFields := []string{}

		// Check git user.name match (30 points)
		if account.GitUserName != "" && userName != "" {
			if strings.EqualFold(account.GitUserName, userName) {
				score += ScoreUserName
				matchedFields = append(matchedFields, "user.name")
			}
		}

		// Check git user.email match (30 points)
		if account.GitEmail != "" && userEmail != "" {
			if strings.EqualFold(account.GitEmail, userEmail) {
				score += ScoreUserEmail
				matchedFields = append(matchedFields, "user.email")
			}
		}

		// Check SSH key in use (20 points)
		if isSSH && account.SSH != nil {
			score += ScoreSSHKey
			matchedFields = append(matchedFields, "ssh")
		} else if !isSSH && account.Token != nil {
			score += ScoreSSHKey
			matchedFields = append(matchedFields, "token")
		}

		// Check platform match (20 points)
		accPlatform := PlatformGitHub
		if account.Platform != nil && account.Platform.Type != "" {
			accPlatform = account.Platform.Type
		}
		if strings.EqualFold(accPlatform, detectedPlatform) {
			score += ScorePlatform
			matchedFields = append(matchedFields, "platform")
		}

		// Update best match if this score is higher
		if score >= MinConfidenceScore && (bestMatch == nil || score > bestMatch.Score) {
			bestMatch = &MatchScore{
				AccountName:   account.Name,
				Score:         score,
				MatchedFields: matchedFields,
				IsActive:      true,
			}
		}
	}

	return bestMatch, nil
}

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
			if strings.EqualFold(userName, account.GitUserName) {
				matches++
			}
		}

		if account.GitEmail != "" {
			totalChecks++
			if strings.EqualFold(userEmail, account.GitEmail) {
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
