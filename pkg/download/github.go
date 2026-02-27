package download

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// githubAPIBase is the GitHub API base URL
const githubAPIBase = "https://api.github.com"

// githubRawBase is the GitHub raw content base URL
const githubRawBase = "https://raw.githubusercontent.com"

// githubTreeEntry represents a single entry in a GitHub tree API response
type githubTreeEntry struct {
	Path string `json:"path"`
	Mode string `json:"mode"`
	Type string `json:"type"` // "blob" or "tree"
	SHA  string `json:"sha"`
	Size int    `json:"size"`
	URL  string `json:"url"`
}

// githubTreeResponse is the response from the GitHub tree API
type githubTreeResponse struct {
	SHA       string            `json:"sha"`
	URL       string            `json:"url"`
	Tree      []githubTreeEntry `json:"tree"`
	Truncated bool              `json:"truncated"`
}

// githubContentsEntry represents a file/dir in the GitHub contents API
type githubContentsEntry struct {
	Name        string `json:"name"`
	Path        string `json:"path"`
	Type        string `json:"type"` // "file" or "dir"
	DownloadURL string `json:"download_url"`
	HTMLURL     string `json:"html_url"`
	Size        int    `json:"size"`
}

// ParseGitHubURL parses a GitHub URL and returns owner, repo, branch, and path.
// Supports:
//   - https://github.com/{owner}/{repo}/blob/{branch}/{path}  (file)
//   - https://github.com/{owner}/{repo}/tree/{branch}/{path}  (directory)
//   - https://github.com/{owner}/{repo}                       (repo root)
func ParseGitHubURL(rawURL string) (owner, repo, branch, path string, isFile bool, err error) {
	// Strip trailing slash
	rawURL = strings.TrimRight(rawURL, "/")

	// Remove scheme and host
	u := rawURL
	u = strings.TrimPrefix(u, "https://github.com/")
	u = strings.TrimPrefix(u, "http://github.com/")

	parts := strings.SplitN(u, "/", -1)
	if len(parts) < 2 {
		err = fmt.Errorf("invalid GitHub URL: %s", rawURL)
		return
	}

	owner = parts[0]
	repo = parts[1]

	if len(parts) < 4 {
		// Just owner/repo — treat as repo root directory
		branch = ""
		path = ""
		isFile = false
		return
	}

	refType := parts[2] // "blob" or "tree"
	branch = parts[3]
	if len(parts) > 4 {
		path = strings.Join(parts[4:], "/")
	}

	isFile = refType == "blob"
	return
}

// GitFile downloads a single file from a GitHub repository.
// The URL can be a GitHub blob URL: https://github.com/{owner}/{repo}/blob/{branch}/{path}
// The folder structure from the repo path is preserved by default.
func GitFile(rawURL string, opts GitOptions) error {
	owner, repo, branch, filePath, _, err := ParseGitHubURL(rawURL)
	if err != nil {
		return err
	}

	if filePath == "" {
		return fmt.Errorf("no file path found in URL: %s", rawURL)
	}

	// Determine branch
	if opts.Branch != "" {
		branch = opts.Branch
	}
	if branch == "" {
		branch = "main"
	}

	// Build raw content URL
	rawContentURL := fmt.Sprintf("%s/%s/%s/%s/%s", githubRawBase, owner, repo, branch, filePath)

	// Determine output filename
	outName := opts.Output
	if outName == "" {
		// By default, preserve the folder structure from the repo path
		// e.g. skill/SKILL.md -> skill/SKILL.md
		outName = filePath
	}

	// Determine output directory
	outDir := opts.OutputDir

	// Build final output path
	var outPath string
	if outDir != "" {
		outPath = filepath.Join(outDir, filepath.FromSlash(outName))
	} else {
		outPath = filepath.FromSlash(outName)
	}

	// Create parent directories
	parentDir := filepath.Dir(outPath)
	if parentDir != "." && parentDir != "" {
		if err := os.MkdirAll(parentDir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", parentDir, err)
		}
	}

	fmt.Printf("  Downloading: %s/%s/%s → %s\n", owner, repo, filePath, outPath)

	// Download the file
	client := &http.Client{Timeout: 5 * time.Minute}
	resp, err := client.Get(rawContentURL)
	if err != nil {
		return fmt.Errorf("failed to download %s: %w", filePath, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		// Try with "master" branch if "main" failed
		if branch == "main" {
			branch = "master"
			rawContentURL = fmt.Sprintf("%s/%s/%s/%s/%s", githubRawBase, owner, repo, branch, filePath)
			resp2, err2 := client.Get(rawContentURL)
			if err2 != nil {
				return fmt.Errorf("failed to download %s: %w", filePath, err2)
			}
			defer resp2.Body.Close()
			if resp2.StatusCode != http.StatusOK {
				return fmt.Errorf("HTTP %d for %s", resp2.StatusCode, filePath)
			}
			return writeFile(outPath, resp2.Body)
		}
		return fmt.Errorf("file not found: %s (branch: %s)", filePath, branch)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d for %s", resp.StatusCode, filePath)
	}

	if err := writeFile(outPath, resp.Body); err != nil {
		return err
	}

	fmt.Printf("  ✓ Saved: %s\n", outPath)
	return nil
}

// GitDirectory downloads all files in a directory from a GitHub repository.
// The URL can be a GitHub tree URL: https://github.com/{owner}/{repo}/tree/{branch}/{path}
func GitDirectory(rawURL string, opts GitOptions) error {
	owner, repo, branch, dirPath, _, err := ParseGitHubURL(rawURL)
	if err != nil {
		return err
	}

	// Determine branch
	if opts.Branch != "" {
		branch = opts.Branch
	}
	if branch == "" {
		branch = "main"
	}

	// Determine max depth
	maxDepth := opts.Depth
	if maxDepth <= 0 {
		maxDepth = 100 // effectively unlimited
	}

	fmt.Printf("  Fetching directory listing: %s/%s/%s\n", owner, repo, dirPath)

	// Use GitHub Contents API to list files recursively
	files, err := listGitHubDirectory(owner, repo, branch, dirPath, maxDepth)
	if err != nil {
		// Try with "master" branch
		if branch == "main" {
			branch = "master"
			files, err = listGitHubDirectory(owner, repo, branch, dirPath, maxDepth)
		}
		if err != nil {
			return fmt.Errorf("failed to list directory: %w", err)
		}
	}

	if len(files) == 0 {
		return fmt.Errorf("no files found in directory: %s", dirPath)
	}

	fmt.Printf("  Found %d file(s) to download\n", len(files))

	// Download each file
	outDir := opts.OutputDir
	errCount := 0
	for _, entry := range files {
		if entry.Type != "file" {
			continue
		}

		// Determine relative path (strip the dirPath prefix)
		relPath := entry.Path
		if dirPath != "" && strings.HasPrefix(relPath, dirPath+"/") {
			relPath = relPath[len(dirPath)+1:]
		} else if dirPath != "" && relPath == dirPath {
			relPath = filepath.Base(relPath)
		}

		// Build output path
		var outPath string
		if outDir != "" {
			outPath = filepath.Join(outDir, filepath.FromSlash(relPath))
		} else {
			outPath = filepath.FromSlash(relPath)
		}

		// Create parent directories
		parentDir := filepath.Dir(outPath)
		if parentDir != "." && parentDir != "" {
			if err := os.MkdirAll(parentDir, 0755); err != nil {
				fmt.Printf("  ✗ Failed to create dir %s: %v\n", parentDir, err)
				errCount++
				continue
			}
		}

		// Download file
		if entry.DownloadURL == "" {
			fmt.Printf("  ✗ No download URL for: %s\n", entry.Path)
			errCount++
			continue
		}

		fmt.Printf("  Downloading: %s → %s\n", entry.Path, outPath)

		client := &http.Client{Timeout: 5 * time.Minute}
		resp, err := client.Get(entry.DownloadURL)
		if err != nil {
			fmt.Printf("  ✗ Failed: %s: %v\n", entry.Path, err)
			errCount++
			continue
		}

		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			fmt.Printf("  ✗ HTTP %d: %s\n", resp.StatusCode, entry.Path)
			errCount++
			continue
		}

		if err := writeFile(outPath, resp.Body); err != nil {
			resp.Body.Close()
			fmt.Printf("  ✗ Failed to write %s: %v\n", outPath, err)
			errCount++
			continue
		}
		resp.Body.Close()

		fmt.Printf("  ✓ Saved: %s\n", outPath)
	}

	if errCount > 0 {
		return fmt.Errorf("%d file(s) failed to download", errCount)
	}

	return nil
}

// listGitHubDirectory recursively lists files in a GitHub directory using the Contents API
func listGitHubDirectory(owner, repo, branch, dirPath string, maxDepth int) ([]githubContentsEntry, error) {
	var results []githubContentsEntry
	err := listGitHubDirectoryRecursive(owner, repo, branch, dirPath, 0, maxDepth, &results)
	return results, err
}

func listGitHubDirectoryRecursive(owner, repo, branch, path string, depth, maxDepth int, results *[]githubContentsEntry) error {
	if depth > maxDepth {
		return nil
	}

	apiURL := fmt.Sprintf("%s/repos/%s/%s/contents/%s?ref=%s", githubAPIBase, owner, repo, path, branch)

	client := &http.Client{Timeout: 30 * time.Second}
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "ghex-downloader/1.0")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to fetch directory listing: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("path not found: %s (branch: %s)", path, branch)
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("GitHub API returned HTTP %d for path: %s", resp.StatusCode, path)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	// The response can be either a single file object or an array
	// Try array first
	var entries []githubContentsEntry
	if err := json.Unmarshal(body, &entries); err != nil {
		// Try single file
		var single githubContentsEntry
		if err2 := json.Unmarshal(body, &single); err2 != nil {
			return fmt.Errorf("failed to parse GitHub API response: %w", err)
		}
		*results = append(*results, single)
		return nil
	}

	for _, entry := range entries {
		if entry.Type == "file" {
			*results = append(*results, entry)
		} else if entry.Type == "dir" {
			// Recurse into subdirectory
			if err := listGitHubDirectoryRecursive(owner, repo, branch, entry.Path, depth+1, maxDepth, results); err != nil {
				// Log but continue
				fmt.Printf("  ⚠ Warning: failed to list %s: %v\n", entry.Path, err)
			}
		}
	}

	return nil
}

// GitRelease downloads release assets from a GitHub repository
func GitRelease(rawURL string, opts ReleaseOptions) error {
	// Parse owner/repo from URL
	u := strings.TrimRight(rawURL, "/")
	u = strings.TrimPrefix(u, "https://github.com/")
	u = strings.TrimPrefix(u, "http://github.com/")
	parts := strings.SplitN(u, "/", 3)
	if len(parts) < 2 {
		return fmt.Errorf("invalid GitHub repository URL: %s", rawURL)
	}
	owner := parts[0]
	repo := parts[1]

	// Determine release API URL
	var apiURL string
	if opts.Version == "" || opts.Version == "latest" {
		apiURL = fmt.Sprintf("%s/repos/%s/%s/releases/latest", githubAPIBase, owner, repo)
	} else {
		apiURL = fmt.Sprintf("%s/repos/%s/%s/releases/tags/%s", githubAPIBase, owner, repo, opts.Version)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "ghex-downloader/1.0")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to fetch release info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("release not found for %s/%s (version: %s)", owner, repo, opts.Version)
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("GitHub API returned HTTP %d", resp.StatusCode)
	}

	var release struct {
		TagName string `json:"tag_name"`
		Name    string `json:"name"`
		Assets  []struct {
			Name               string `json:"name"`
			BrowserDownloadURL string `json:"browser_download_url"`
			Size               int64  `json:"size"`
		} `json:"assets"`
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}
	if err := json.Unmarshal(body, &release); err != nil {
		return fmt.Errorf("failed to parse release info: %w", err)
	}

	fmt.Printf("  Release: %s (%s)\n", release.Name, release.TagName)
	fmt.Printf("  Assets: %d\n", len(release.Assets))

	if opts.ListOnly {
		for _, asset := range release.Assets {
			fmt.Printf("    • %s (%s)\n", asset.Name, formatSize(asset.Size))
		}
		return nil
	}

	// Filter assets
	var toDownload []struct {
		Name string
		URL  string
		Size int64
	}
	for _, asset := range release.Assets {
		if opts.Asset == "" || strings.Contains(strings.ToLower(asset.Name), strings.ToLower(opts.Asset)) {
			toDownload = append(toDownload, struct {
				Name string
				URL  string
				Size int64
			}{asset.Name, asset.BrowserDownloadURL, asset.Size})
		}
	}

	if len(toDownload) == 0 {
		return fmt.Errorf("no assets matched filter: %q", opts.Asset)
	}

	outDir := opts.OutputDir
	if outDir != "" {
		if err := os.MkdirAll(outDir, 0755); err != nil {
			return fmt.Errorf("failed to create output directory: %w", err)
		}
	}

	for _, asset := range toDownload {
		outPath := asset.Name
		if outDir != "" {
			outPath = filepath.Join(outDir, asset.Name)
		}

		fmt.Printf("  Downloading: %s (%s) → %s\n", asset.Name, formatSize(asset.Size), outPath)

		dlResp, err := client.Get(asset.URL)
		if err != nil {
			fmt.Printf("  ✗ Failed: %v\n", err)
			continue
		}

		if dlResp.StatusCode != http.StatusOK {
			dlResp.Body.Close()
			fmt.Printf("  ✗ HTTP %d\n", dlResp.StatusCode)
			continue
		}

		if err := writeFile(outPath, dlResp.Body); err != nil {
			dlResp.Body.Close()
			fmt.Printf("  ✗ Failed to write: %v\n", err)
			continue
		}
		dlResp.Body.Close()

		fmt.Printf("  ✓ Saved: %s\n", outPath)
	}

	return nil
}

// writeFile writes data from a reader to a file path, creating parent dirs as needed
func writeFile(path string, r io.Reader) error {
	dir := filepath.Dir(path)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}
	}

	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer f.Close()

	if _, err := io.Copy(f, r); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}
