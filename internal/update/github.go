package update

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

const (
	defaultGitHubAPI = "https://api.github.com"
	defaultTimeout   = 30 * time.Second
)

// ReleaseInfo contains information about a GitHub release
type ReleaseInfo struct {
	Version     string    `json:"-"`
	TagName     string    `json:"tag_name"`
	Name        string    `json:"name"`
	Body        string    `json:"body"`
	PublishedAt time.Time `json:"published_at"`
	HTMLURL     string    `json:"html_url"`
	Assets      []Asset   `json:"assets"`
}

// Asset represents a downloadable release asset
type Asset struct {
	Name        string `json:"name"`
	DownloadURL string `json:"browser_download_url"`
	Size        int64  `json:"size"`
	ContentType string `json:"content_type"`
}

// GitHubClient handles GitHub API interactions
type GitHubClient struct {
	HTTPClient *http.Client
	BaseURL    string
}

// NewGitHubClient creates a new GitHub client
func NewGitHubClient() *GitHubClient {
	return &GitHubClient{
		HTTPClient: &http.Client{
			Timeout: defaultTimeout,
		},
		BaseURL: defaultGitHubAPI,
	}
}


// GetLatestRelease fetches the latest release from GitHub
func (c *GitHubClient) GetLatestRelease(owner, repo string) (*ReleaseInfo, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/releases/latest", c.BaseURL, owner, repo)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrNetworkError, err)
	}

	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "ghex-updater")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrNetworkError, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("no releases found for %s/%s", owner, repo)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: HTTP %d", ErrNetworkError, resp.StatusCode)
	}

	var release ReleaseInfo
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, fmt.Errorf("failed to parse release info: %w", err)
	}

	// Parse version from tag
	version, err := ParseVersion(release.TagName)
	if err == nil {
		release.Version = version.String()
	} else {
		release.Version = release.TagName
	}

	return &release, nil
}

// GetReleases fetches all releases from GitHub
func (c *GitHubClient) GetReleases(owner, repo string, limit int) ([]ReleaseInfo, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/releases?per_page=%d", c.BaseURL, owner, repo, limit)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrNetworkError, err)
	}

	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "ghex-updater")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrNetworkError, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: HTTP %d", ErrNetworkError, resp.StatusCode)
	}

	var releases []ReleaseInfo
	if err := json.NewDecoder(resp.Body).Decode(&releases); err != nil {
		return nil, fmt.Errorf("failed to parse releases: %w", err)
	}

	// Parse versions
	for i := range releases {
		version, err := ParseVersion(releases[i].TagName)
		if err == nil {
			releases[i].Version = version.String()
		} else {
			releases[i].Version = releases[i].TagName
		}
	}

	return releases, nil
}


// ProgressCallback is called during download with current and total bytes
type ProgressCallback func(current, total int64)

// DownloadAsset downloads a release asset with progress
func (c *GitHubClient) DownloadAsset(asset *Asset, destPath string, progress ProgressCallback) error {
	req, err := http.NewRequest("GET", asset.DownloadURL, nil)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrDownloadFailed, err)
	}

	req.Header.Set("Accept", "application/octet-stream")
	req.Header.Set("User-Agent", "ghex-updater")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrDownloadFailed, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%w: HTTP %d", ErrDownloadFailed, resp.StatusCode)
	}

	// Create destination file
	out, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrDownloadFailed, err)
	}
	defer out.Close()

	// Download with progress
	var reader io.Reader = resp.Body
	if progress != nil && resp.ContentLength > 0 {
		reader = &progressReader{
			reader:   resp.Body,
			total:    resp.ContentLength,
			callback: progress,
		}
	}

	_, err = io.Copy(out, reader)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrDownloadFailed, err)
	}

	return nil
}

// progressReader wraps a reader to report download progress
type progressReader struct {
	reader   io.Reader
	total    int64
	current  int64
	callback ProgressCallback
}

func (pr *progressReader) Read(p []byte) (int, error) {
	n, err := pr.reader.Read(p)
	pr.current += int64(n)
	if pr.callback != nil {
		pr.callback(pr.current, pr.total)
	}
	return n, err
}

// DownloadChecksums downloads the checksums file for a release
func (c *GitHubClient) DownloadChecksums(release *ReleaseInfo) (string, error) {
	// Find checksums asset
	var checksumAsset *Asset
	for i := range release.Assets {
		name := release.Assets[i].Name
		if name == "checksums.txt" || name == "SHA256SUMS" || name == "sha256sums.txt" {
			checksumAsset = &release.Assets[i]
			break
		}
	}

	if checksumAsset == nil {
		return "", nil // No checksums file, not an error
	}

	req, err := http.NewRequest("GET", checksumAsset.DownloadURL, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("User-Agent", "ghex-updater")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to download checksums: HTTP %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}
