package download

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/dwirx/ghex/internal/platform"
	"github.com/dwirx/ghex/internal/ui"
)

// GitOptions configures git download behavior.
type GitOptions struct {
	Branch    string // Branch/tag/commit (empty = default branch)
	Output    string // Output filename for single file
	OutputDir string // Output directory
	Depth     int    // Max directory depth (0 = unlimited)
	Overwrite bool   // Overwrite existing files
	ShowInfo  bool   // Show file info before download
	Token     string // GitHub personal access token (falls back to GITHUB_TOKEN env var)
}

// ReleaseOptions configures release download behavior.
type ReleaseOptions struct {
	Version   string // Release version/tag (empty = latest)
	Asset     string // Asset name filter
	OutputDir string // Output directory
	ListOnly  bool   // Only list assets, don't download
	Token     string // GitHub personal access token
	Overwrite bool   // Overwrite existing files
}

// ParsedGitURL represents a parsed git URL.
type ParsedGitURL struct {
	Platform    string // github, gitlab, bitbucket
	Owner       string
	Repo        string
	Branch      string
	FilePath    string
	IsDirectory bool
}

// GitFile downloads a single file from a git repository.
func GitFile(url string, opts GitOptions) error {
	parsed, err := parseGitURL(url)
	if err != nil {
		return err
	}

	if opts.Branch != "" {
		parsed.Branch = opts.Branch
	}

	if parsed.IsDirectory {
		ui.ShowWarning("This appears to be a directory. Use GitDirectory instead.")
		return nil
	}

	// Resolve token: explicit option takes precedence over env var
	token := opts.Token
	if token == "" {
		token = os.Getenv("GITHUB_TOKEN")
	}

	rawURL := toRawURL(parsed)
	filename := opts.Output
	if filename == "" {
		filename = filepath.Base(parsed.FilePath)
	}

	ui.ShowSection("Downloading File")
	ui.ShowKeyValue("Repository", fmt.Sprintf("%s/%s", parsed.Owner, parsed.Repo))
	ui.ShowKeyValue("Branch", parsed.Branch)
	ui.ShowKeyValue("File", parsed.FilePath)
	fmt.Println()

	downloadOpts := Options{
		Output:          filename,
		OutputDir:       opts.OutputDir,
		Overwrite:       opts.Overwrite,
		ShowProgress:    true,
		ShowInfo:        opts.ShowInfo,
		FollowRedirects: true,
		Token:           token,
	}

	err = FromURL(rawURL, downloadOpts)
	if err != nil {
		// If main branch 404s and no explicit branch was set, try master
		var notFound *ErrNotFound
		if isErrNotFound(err, &notFound) && parsed.Branch == "main" && opts.Branch == "" {
			parsed.Branch = "master"
			rawURL = toRawURL(parsed)
			ui.ShowInfo("Branch 'main' not found, trying 'master'...")
			err = FromURL(rawURL, downloadOpts)
		}
	}
	return err
}

// isErrNotFound checks if err is an ErrNotFound and sets target if so.
func isErrNotFound(err error, target **ErrNotFound) bool {
	if nf, ok := err.(*ErrNotFound); ok {
		*target = nf
		return true
	}
	// Also check ErrHTTP with 404
	if he, ok := err.(*ErrHTTP); ok && he.StatusCode == 404 {
		return true
	}
	return false
}

// GitDirectory downloads a directory from a git repository.
func GitDirectory(url string, opts GitOptions) error {
	parsed, err := parseGitURL(url)
	if err != nil {
		return err
	}

	if opts.Branch != "" {
		parsed.Branch = opts.Branch
	}

	if parsed.Platform != "github" {
		return fmt.Errorf("directory download only supported for GitHub")
	}

	// Resolve token: explicit option takes precedence over env var
	token := opts.Token
	if token == "" {
		token = os.Getenv("GITHUB_TOKEN")
	}

	ui.ShowSection("Downloading Directory")
	ui.ShowKeyValue("Repository", fmt.Sprintf("%s/%s", parsed.Owner, parsed.Repo))
	ui.ShowKeyValue("Branch", parsed.Branch)
	if parsed.FilePath != "" {
		ui.ShowKeyValue("Path", parsed.FilePath)
	} else {
		ui.ShowKeyValue("Path", "(repository root)")
	}
	fmt.Println()

	// Fetch directory contents
	files, err := fetchDirectoryContents(parsed, opts.Depth, token)
	if err != nil {
		// If main branch fails and no explicit branch was set, try master
		if parsed.Branch == "main" && opts.Branch == "" {
			parsed.Branch = "master"
			ui.ShowInfo("Branch 'main' not found, trying 'master'...")
			files, err = fetchDirectoryContents(parsed, opts.Depth, token)
		}
		if err != nil {
			return err
		}
	}

	if len(files) == 0 {
		ui.ShowWarning("No files found in directory")
		return nil
	}

	ui.ShowInfo(fmt.Sprintf("Found %d files", len(files)))

	// Determine output directory
	outputDir := opts.OutputDir
	if outputDir == "" {
		outputDir = parsed.Repo
	}

	successful := 0
	for _, file := range files {
		relPath := file.Path
		if parsed.FilePath != "" {
			relPath = strings.TrimPrefix(file.Path, parsed.FilePath+"/")
		}

		outputPath := filepath.Join(outputDir, relPath)
		dir := filepath.Dir(outputPath)
		if err := platform.EnsureDir(dir, 0755); err != nil {
			ui.ShowError(fmt.Sprintf("Failed to create directory: %v", err))
			continue
		}

		downloadOpts := Options{
			Output:          filepath.Base(outputPath),
			OutputDir:       dir,
			Overwrite:       opts.Overwrite,
			ShowProgress:    false,
			FollowRedirects: true,
			Token:           token,
		}

		if err := FromURL(file.URL, downloadOpts); err != nil {
			ui.ShowError(fmt.Sprintf("Failed to download %s: %v", file.Path, err))
		} else {
			successful++
		}
	}

	ui.ShowSuccess(fmt.Sprintf("Downloaded %d/%d files to %s", successful, len(files), outputDir))
	return nil
}

// GitRelease downloads release assets from GitHub.
func GitRelease(url string, opts ReleaseOptions) error {
	parsed, err := parseGitURL(url)
	if err != nil {
		return err
	}

	if parsed.Platform != "github" {
		return fmt.Errorf("release download only supported for GitHub")
	}

	// Resolve token: explicit option takes precedence over env var
	token := opts.Token
	if token == "" {
		token = os.Getenv("GITHUB_TOKEN")
	}

	ui.ShowSection("GitHub Release")
	ui.ShowKeyValue("Repository", fmt.Sprintf("%s/%s", parsed.Owner, parsed.Repo))

	// Fetch release info
	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", parsed.Owner, parsed.Repo)
	if opts.Version != "" {
		apiURL = fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/tags/%s", parsed.Owner, parsed.Repo, opts.Version)
	}

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "ghex-downloader/1.0")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to fetch release: %w", err)
	}
	defer resp.Body.Close()

	// Check for rate limiting
	if resp.StatusCode == http.StatusForbidden {
		if resp.Header.Get("X-RateLimit-Remaining") == "0" {
			resetAt := resp.Header.Get("X-RateLimit-Reset")
			return &ErrRateLimit{ResetAt: resetAt}
		}
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("release not found: %s", resp.Status)
	}

	var release struct {
		TagName     string `json:"tag_name"`
		Name        string `json:"name"`
		PublishedAt string `json:"published_at"`
		Assets      []struct {
			Name               string `json:"name"`
			Size               int64  `json:"size"`
			BrowserDownloadURL string `json:"browser_download_url"`
		} `json:"assets"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return fmt.Errorf("failed to parse release: %w", err)
	}

	ui.ShowKeyValue("Version", release.TagName)
	if len(release.PublishedAt) >= 10 {
		ui.ShowKeyValue("Published", release.PublishedAt[:10])
	}
	fmt.Println()

	if len(release.Assets) == 0 {
		ui.ShowWarning("No assets found in this release")
		return nil
	}

	// Filter assets
	assets := release.Assets
	if opts.Asset != "" {
		var filtered []struct {
			Name               string `json:"name"`
			Size               int64  `json:"size"`
			BrowserDownloadURL string `json:"browser_download_url"`
		}
		for _, a := range assets {
			if strings.Contains(strings.ToLower(a.Name), strings.ToLower(opts.Asset)) {
				filtered = append(filtered, a)
			}
		}
		assets = filtered
	}

	if len(assets) == 0 {
		ui.ShowWarning(fmt.Sprintf("No assets found matching: %s", opts.Asset))
		return nil
	}

	// List assets
	fmt.Println(ui.Primary("Available assets:"))
	for i, asset := range assets {
		size := formatSize(asset.Size)
		fmt.Printf("  %s %s (%s)\n", ui.Dim(fmt.Sprintf("[%d]", i+1)), asset.Name, size)
	}
	fmt.Println()

	if opts.ListOnly {
		return nil
	}

	// Select asset
	choice := ui.Prompt("Select asset to download (number or 'all')")
	if choice == "" {
		return nil
	}

	var toDownload []struct {
		Name               string `json:"name"`
		Size               int64  `json:"size"`
		BrowserDownloadURL string `json:"browser_download_url"`
	}

	if choice == "all" {
		toDownload = assets
	} else {
		var idx int
		_, _ = fmt.Sscanf(choice, "%d", &idx)
		if idx < 1 || idx > len(assets) {
			return fmt.Errorf("invalid selection")
		}
		toDownload = append(toDownload, assets[idx-1])
	}

	// Download selected assets
	for _, asset := range toDownload {
		downloadOpts := Options{
			Output:          asset.Name,
			OutputDir:       opts.OutputDir,
			Overwrite:       opts.Overwrite,
			ShowProgress:    true,
			FollowRedirects: true,
			Token:           token,
		}

		if err := FromURL(asset.BrowserDownloadURL, downloadOpts); err != nil {
			ui.ShowError(fmt.Sprintf("Failed to download %s: %v", asset.Name, err))
		}
	}

	return nil
}

// parseGitURL parses a git repository URL.
func parseGitURL(url string) (*ParsedGitURL, error) {
	parsed := &ParsedGitURL{
		Branch: "main",
	}

	// GitHub patterns
	githubBlobPattern := regexp.MustCompile(`github\.com/([^/]+)/([^/]+)/blob/([^/]+)/(.+)`)
	githubTreePattern := regexp.MustCompile(`github\.com/([^/]+)/([^/]+)/tree/([^/]+)/(.+)`)
	githubRepoPattern := regexp.MustCompile(`github\.com/([^/]+)/([^/]+)`)

	if matches := githubBlobPattern.FindStringSubmatch(url); matches != nil {
		parsed.Platform = "github"
		parsed.Owner = matches[1]
		parsed.Repo = matches[2]
		parsed.Branch = matches[3]
		parsed.FilePath = matches[4]
		parsed.IsDirectory = false
		return parsed, nil
	}

	if matches := githubTreePattern.FindStringSubmatch(url); matches != nil {
		parsed.Platform = "github"
		parsed.Owner = matches[1]
		parsed.Repo = matches[2]
		parsed.Branch = matches[3]
		parsed.FilePath = matches[4]
		parsed.IsDirectory = true
		return parsed, nil
	}

	if matches := githubRepoPattern.FindStringSubmatch(url); matches != nil {
		parsed.Platform = "github"
		parsed.Owner = matches[1]
		parsed.Repo = strings.TrimSuffix(matches[2], ".git")
		parsed.FilePath = "" // repo root
		parsed.IsDirectory = true
		return parsed, nil
	}

	// GitLab patterns
	gitlabBlobPattern := regexp.MustCompile(`gitlab\.com/([^/]+)/([^/]+)/-/blob/([^/]+)/(.+)`)
	gitlabTreePattern := regexp.MustCompile(`gitlab\.com/([^/]+)/([^/]+)/-/tree/([^/]+)/(.+)`)

	if matches := gitlabBlobPattern.FindStringSubmatch(url); matches != nil {
		parsed.Platform = "gitlab"
		parsed.Owner = matches[1]
		parsed.Repo = matches[2]
		parsed.Branch = matches[3]
		parsed.FilePath = matches[4]
		parsed.IsDirectory = false
		return parsed, nil
	}

	if matches := gitlabTreePattern.FindStringSubmatch(url); matches != nil {
		parsed.Platform = "gitlab"
		parsed.Owner = matches[1]
		parsed.Repo = matches[2]
		parsed.Branch = matches[3]
		parsed.FilePath = matches[4]
		parsed.IsDirectory = true
		return parsed, nil
	}

	return nil, fmt.Errorf("unsupported URL format: %s", url)
}

// toRawURL converts a parsed URL to raw download URL.
func toRawURL(parsed *ParsedGitURL) string {
	switch parsed.Platform {
	case "github":
		return fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/%s/%s",
			parsed.Owner, parsed.Repo, parsed.Branch, parsed.FilePath)
	case "gitlab":
		return fmt.Sprintf("https://gitlab.com/%s/%s/-/raw/%s/%s",
			parsed.Owner, parsed.Repo, parsed.Branch, parsed.FilePath)
	default:
		return ""
	}
}

type fileInfo struct {
	Path string
	URL  string
}

// fetchDirectoryContents fetches all files in a directory using the GitHub Contents API.
// token is optional; if provided it is sent as Authorization: Bearer <token>.
func fetchDirectoryContents(parsed *ParsedGitURL, maxDepth int, token string) ([]fileInfo, error) {
	var files []fileInfo

	var fetchRecursive func(path string, depth int) error
	fetchRecursive = func(path string, depth int) error {
		if maxDepth > 0 && depth > maxDepth {
			return nil
		}

		apiURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/contents/%s?ref=%s",
			parsed.Owner, parsed.Repo, path, parsed.Branch)

		req, err := http.NewRequest("GET", apiURL, nil)
		if err != nil {
			return err
		}
		req.Header.Set("Accept", "application/vnd.github.v3+json")
		req.Header.Set("User-Agent", "ghex-downloader/1.0")
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		// Check for rate limiting
		if resp.StatusCode == http.StatusForbidden {
			if resp.Header.Get("X-RateLimit-Remaining") == "0" {
				resetAt := resp.Header.Get("X-RateLimit-Reset")
				return &ErrRateLimit{ResetAt: resetAt}
			}
		}

		if resp.StatusCode == http.StatusNotFound {
			return &ErrNotFound{URL: apiURL}
		}

		if resp.StatusCode != http.StatusOK {
			return &ErrHTTP{StatusCode: resp.StatusCode, Status: resp.Status, URL: apiURL}
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		var contents []struct {
			Name        string `json:"name"`
			Path        string `json:"path"`
			Type        string `json:"type"`
			DownloadURL string `json:"download_url"`
		}

		if err := json.Unmarshal(body, &contents); err != nil {
			return err
		}

		for _, item := range contents {
			if item.Type == "file" {
				files = append(files, fileInfo{
					Path: item.Path,
					URL:  item.DownloadURL,
				})
			} else if item.Type == "dir" {
				if err := fetchRecursive(item.Path, depth+1); err != nil {
					// Continue on error but log it
					ui.ShowWarning(fmt.Sprintf("Failed to list %s: %v", item.Path, err))
				}
			}
		}

		return nil
	}

	// Handle repo root (empty FilePath) — pass empty string to API
	startPath := parsed.FilePath
	if err := fetchRecursive(startPath, 0); err != nil {
		return nil, err
	}

	return files, nil
}
