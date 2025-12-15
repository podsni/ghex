package git

import (
	"fmt"
	"regexp"
	"strings"
)

// URLInfo contains parsed information from a git URL
type URLInfo struct {
	URL      string
	IsSSH    bool
	Host     string
	Owner    string
	Repo     string
	Platform string // github, gitlab, bitbucket, gitea, other
}

// ParseRepoFromURL extracts owner/repo from a git URL
func ParseRepoFromURL(rawURL string) (owner, repo string, err error) {
	if rawURL == "" {
		return "", "", fmt.Errorf("empty URL")
	}

	rawURL = strings.TrimSpace(rawURL)

	// SSH format: git@host:owner/repo.git
	sshPattern := regexp.MustCompile(`^git@([^:]+):(.+?)(?:\.git)?$`)
	if matches := sshPattern.FindStringSubmatch(rawURL); len(matches) == 3 {
		parts := strings.Split(matches[2], "/")
		if len(parts) >= 2 {
			return parts[0], strings.TrimSuffix(parts[len(parts)-1], ".git"), nil
		}
	}

	// SSH format: ssh://git@host/owner/repo.git
	sshURLPattern := regexp.MustCompile(`^ssh://git@([^/]+)/(.+?)(?:\.git)?$`)
	if matches := sshURLPattern.FindStringSubmatch(rawURL); len(matches) == 3 {
		parts := strings.Split(matches[2], "/")
		if len(parts) >= 2 {
			return parts[0], strings.TrimSuffix(parts[len(parts)-1], ".git"), nil
		}
	}

	// HTTPS format: https://host/owner/repo.git
	httpsPattern := regexp.MustCompile(`^https?://([^/]+)/(.+?)(?:\.git)?$`)
	if matches := httpsPattern.FindStringSubmatch(rawURL); len(matches) == 3 {
		parts := strings.Split(matches[2], "/")
		if len(parts) >= 2 {
			return parts[0], strings.TrimSuffix(parts[len(parts)-1], ".git"), nil
		}
	}

	return "", "", fmt.Errorf("unable to parse URL: %s", rawURL)
}

// NormalizeURL normalizes a git URL and adds .git suffix if missing
func NormalizeURL(rawURL string) (normalized string, isSSH bool, err error) {
	if rawURL == "" {
		return "", false, fmt.Errorf("empty URL")
	}

	rawURL = strings.TrimSpace(rawURL)
	rawURL = strings.TrimSuffix(rawURL, "#") // Remove trailing #

	// SSH format: git@host:path or ssh://git@host/path
	if strings.HasPrefix(rawURL, "git@") || strings.HasPrefix(rawURL, "ssh://") {
		if !strings.HasSuffix(rawURL, ".git") {
			rawURL += ".git"
		}
		return rawURL, true, nil
	}

	// HTTPS format
	if strings.HasPrefix(rawURL, "http://") || strings.HasPrefix(rawURL, "https://") {
		if !strings.HasSuffix(rawURL, ".git") {
			rawURL += ".git"
		}
		return rawURL, false, nil
	}

	return "", false, fmt.Errorf("invalid git URL format: %s", rawURL)
}

// ParseURL parses a git URL and returns detailed information
func ParseURL(rawURL string) (*URLInfo, error) {
	normalized, isSSH, err := NormalizeURL(rawURL)
	if err != nil {
		return nil, err
	}

	owner, repo, err := ParseRepoFromURL(normalized)
	if err != nil {
		return nil, err
	}

	host := detectHost(normalized)
	platform := detectPlatform(host)

	return &URLInfo{
		URL:      normalized,
		IsSSH:    isSSH,
		Host:     host,
		Owner:    owner,
		Repo:     repo,
		Platform: platform,
	}, nil
}

// detectHost extracts the host from a URL
func detectHost(rawURL string) string {
	// SSH format: git@host:path
	if strings.HasPrefix(rawURL, "git@") {
		parts := strings.SplitN(rawURL[4:], ":", 2)
		if len(parts) > 0 {
			return parts[0]
		}
	}

	// SSH URL format: ssh://git@host/path
	if strings.HasPrefix(rawURL, "ssh://git@") {
		parts := strings.SplitN(rawURL[10:], "/", 2)
		if len(parts) > 0 {
			return parts[0]
		}
	}

	// HTTPS format
	httpsPattern := regexp.MustCompile(`^https?://([^/]+)`)
	if matches := httpsPattern.FindStringSubmatch(rawURL); len(matches) == 2 {
		return matches[1]
	}

	return ""
}

// detectPlatform detects the git platform from the host
func detectPlatform(host string) string {
	host = strings.ToLower(host)

	if strings.Contains(host, "github") {
		return "github"
	}
	if strings.Contains(host, "gitlab") {
		return "gitlab"
	}
	if strings.Contains(host, "bitbucket") {
		return "bitbucket"
	}
	if strings.Contains(host, "gitea") {
		return "gitea"
	}

	return "other"
}

// BuildRemoteURL builds a remote URL for a given platform
func BuildRemoteURL(platform, domain, repoPath string, useSSH bool) string {
	if domain == "" {
		switch platform {
		case "github":
			domain = "github.com"
		case "gitlab":
			domain = "gitlab.com"
		case "bitbucket":
			domain = "bitbucket.org"
		default:
			domain = "github.com"
		}
	}

	// Ensure repo path has .git suffix
	if !strings.HasSuffix(repoPath, ".git") {
		repoPath += ".git"
	}

	if useSSH {
		return fmt.Sprintf("git@%s:%s", domain, repoPath)
	}

	return fmt.Sprintf("https://%s/%s", domain, repoPath)
}

// WithGitSuffix ensures a repo path has .git suffix
func WithGitSuffix(repoPath string) string {
	if strings.HasSuffix(repoPath, ".git") {
		return repoPath
	}
	return repoPath + ".git"
}

// GetPlatformSSHHost returns the SSH host for a platform
func GetPlatformSSHHost(platform, domain string) string {
	if domain != "" {
		return domain
	}

	switch platform {
	case "github":
		return "github.com"
	case "gitlab":
		return "gitlab.com"
	case "bitbucket":
		return "bitbucket.org"
	default:
		return "github.com"
	}
}
