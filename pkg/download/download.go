// Package download provides utilities for downloading files from URLs and Git repositories.
// It supports single file downloads, GitHub file/folder downloads via the GitHub API,
// and release asset downloads.
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

// Options configures a generic URL download
type Options struct {
	// Output is the output filename (optional; auto-detected from URL if empty)
	Output string
	// OutputDir is the directory to save the file in (default: current directory)
	OutputDir string
	// Overwrite allows overwriting existing files
	Overwrite bool
	// ShowProgress enables progress output to stdout
	ShowProgress bool
	// ShowInfo prints file metadata before downloading
	ShowInfo bool
	// FollowRedirects enables following HTTP redirects
	FollowRedirects bool
}

// DefaultOptions returns Options with sensible defaults
func DefaultOptions() Options {
	return Options{
		ShowProgress:    true,
		FollowRedirects: true,
	}
}

// httpClient is a shared HTTP client with a reasonable timeout
var httpClient = &http.Client{
	Timeout: 60 * time.Second,
}

// FromURL downloads a file from any HTTP/HTTPS URL.
func FromURL(rawURL string, opts Options) error {
	// Auto-detect output filename from URL if not specified
	outputName := opts.Output
	if outputName == "" {
		parts := strings.Split(rawURL, "/")
		for i := len(parts) - 1; i >= 0; i-- {
			if parts[i] != "" {
				outputName = parts[i]
				// Strip query string if present
				if idx := strings.Index(outputName, "?"); idx != -1 {
					outputName = outputName[:idx]
				}
				break
			}
		}
	}
	if outputName == "" {
		outputName = "download"
	}

	destPath := filepath.Join(opts.OutputDir, outputName)

	if opts.ShowInfo {
		fmt.Printf("  URL:  %s\n", rawURL)
		fmt.Printf("  Dest: %s\n", destPath)
	}

	return downloadFile(rawURL, destPath, opts.Overwrite, opts.ShowProgress)
}

// Multiple downloads multiple files from a list of URLs.
func Multiple(urls []string, opts Options) error {
	var errs []string
	for i, u := range urls {
		if opts.ShowProgress {
			fmt.Printf("[%d/%d] %s\n", i+1, len(urls), u)
		}
		if err := FromURL(u, opts); err != nil {
			errs = append(errs, fmt.Sprintf("%s: %v", u, err))
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("some downloads failed:\n  %s", strings.Join(errs, "\n  "))
	}
	return nil
}

// downloadFile downloads a single file from a URL to a local path.
// If showProgress is true, a simple progress indicator is printed.
func downloadFile(url, destPath string, overwrite, showProgress bool) error {
	// Check if file already exists
	if !overwrite {
		if _, err := os.Stat(destPath); err == nil {
			if showProgress {
				fmt.Printf("  ✓ Already exists: %s\n", destPath)
			}
			return nil
		}
	}

	// Ensure parent directory exists
	if dir := filepath.Dir(destPath); dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("User-Agent", "ghex-downloader/1.0")

	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("download failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("file not found (404): %s", url)
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned status %d for %s", resp.StatusCode, url)
	}

	// Create destination file
	f, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", destPath, err)
	}
	defer f.Close()

	// Copy with optional progress
	if showProgress && resp.ContentLength > 0 {
		written, err := io.Copy(f, &progressReader{
			reader: resp.Body,
			total:  resp.ContentLength,
			dest:   destPath,
		})
		if err != nil {
			return fmt.Errorf("failed to write file: %w", err)
		}
		fmt.Printf("\r  ✓ %s (%s)\n", destPath, formatSize(written))
	} else {
		if _, err := io.Copy(f, resp.Body); err != nil {
			return fmt.Errorf("failed to write file: %w", err)
		}
		if showProgress {
			info, _ := f.Stat()
			size := int64(0)
			if info != nil {
				size = info.Size()
			}
			fmt.Printf("  ✓ %s (%s)\n", destPath, formatSize(size))
		}
	}

	return nil
}

// progressReader wraps an io.Reader and prints download progress.
type progressReader struct {
	reader  io.Reader
	total   int64
	written int64
	dest    string
}

func (p *progressReader) Read(buf []byte) (int, error) {
	n, err := p.reader.Read(buf)
	p.written += int64(n)
	if p.total > 0 {
		pct := float64(p.written) / float64(p.total) * 100
		fmt.Printf("\r  ↓ %s  %.0f%%  (%s / %s)   ",
			filepath.Base(p.dest),
			pct,
			formatSize(p.written),
			formatSize(p.total),
		)
	}
	return n, err
}
