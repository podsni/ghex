// Package download provides utilities for downloading files from URLs and Git repositories.
package download

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Options configures a generic HTTP download
type Options struct {
	Output          string // Output filename (empty = auto-detect from URL)
	OutputDir       string // Output directory (empty = current directory)
	Overwrite       bool   // Overwrite existing files
	ShowProgress    bool   // Show download progress
	ShowInfo        bool   // Show file info before download
	FollowRedirects bool   // Follow HTTP redirects
}

// DefaultOptions returns sensible default download options
func DefaultOptions() Options {
	return Options{
		ShowProgress:    true,
		FollowRedirects: true,
	}
}

// GitOptions configures a Git repository download
type GitOptions struct {
	Branch    string // Branch/tag/commit (empty = default branch)
	Output    string // Output filename for single file
	OutputDir string // Output directory
	Depth     int    // Max directory depth (0 = unlimited)
	// PreserveStructure controls whether to keep the folder path from the repo
	// e.g. downloading skill/SKILL.md will create skill/SKILL.md locally
	PreserveStructure bool
}

// ReleaseOptions configures a GitHub release download
type ReleaseOptions struct {
	Version   string // Release version/tag (empty = latest)
	Asset     string // Asset name filter
	OutputDir string // Output directory
	ListOnly  bool   // Only list assets, don't download
}

// FromURL downloads a file from a generic HTTP/HTTPS URL
func FromURL(rawURL string, opts Options) error {
	client := &http.Client{
		Timeout: 5 * time.Minute,
	}
	if !opts.FollowRedirects {
		client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}
	}

	resp, err := client.Get(rawURL)
	if err != nil {
		return fmt.Errorf("failed to fetch URL: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	// Determine output filename
	outName := opts.Output
	if outName == "" {
		outName = filenameFromURL(rawURL)
	}
	if outName == "" {
		outName = "download"
	}

	// Determine output path
	outPath := outName
	if opts.OutputDir != "" {
		if err := os.MkdirAll(opts.OutputDir, 0755); err != nil {
			return fmt.Errorf("failed to create output directory: %w", err)
		}
		outPath = filepath.Join(opts.OutputDir, outName)
	}

	// Check overwrite
	if !opts.Overwrite {
		if _, err := os.Stat(outPath); err == nil {
			return fmt.Errorf("file already exists: %s (use --overwrite to replace)", outPath)
		}
	}

	if opts.ShowInfo {
		fmt.Printf("  URL:  %s\n", rawURL)
		fmt.Printf("  Size: %s\n", formatSize(resp.ContentLength))
		fmt.Printf("  Dest: %s\n", outPath)
	}

	// Write file
	f, err := os.Create(outPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer f.Close()

	if opts.ShowProgress {
		fmt.Printf("  Downloading → %s\n", outPath)
	}

	if _, err := io.Copy(f, resp.Body); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	if opts.ShowProgress {
		fmt.Printf("  ✓ Saved: %s\n", outPath)
	}

	return nil
}

// Multiple downloads multiple files from a list of URLs
func Multiple(urls []string, opts Options) error {
	var errs []string
	for i, u := range urls {
		fmt.Printf("[%d/%d] %s\n", i+1, len(urls), u)
		if err := FromURL(u, opts); err != nil {
			errs = append(errs, fmt.Sprintf("%s: %v", u, err))
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("some downloads failed:\n%s", strings.Join(errs, "\n"))
	}
	return nil
}

// filenameFromURL extracts the filename from a URL path
func filenameFromURL(rawURL string) string {
	// Remove query string and fragment
	u := rawURL
	if idx := strings.Index(u, "?"); idx != -1 {
		u = u[:idx]
	}
	if idx := strings.Index(u, "#"); idx != -1 {
		u = u[:idx]
	}
	parts := strings.Split(u, "/")
	for i := len(parts) - 1; i >= 0; i-- {
		if parts[i] != "" {
			return parts[i]
		}
	}
	return ""
}

// formatSize returns a human-readable file size
func formatSize(bytes int64) string {
	if bytes < 0 {
		return "unknown"
	}
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
