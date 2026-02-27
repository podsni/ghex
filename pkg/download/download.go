// Package download provides utilities for downloading files from URLs and Git repositories.
package download

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// Options configures a generic HTTP download.
type Options struct {
	Output          string            // Output filename (empty = auto-detect from URL)
	OutputDir       string            // Output directory (empty = current directory)
	Overwrite       bool              // Overwrite existing files
	ShowProgress    bool              // Show download progress
	ShowInfo        bool              // Show file info before download
	FollowRedirects bool              // Follow HTTP redirects
	Token           string            // Bearer token for authentication
	Retries         int               // Max retry attempts (0 = use default 3)
	Timeout         time.Duration     // HTTP timeout (0 = use default 5 minutes)
	Headers         map[string]string // Additional HTTP headers
}

// DefaultOptions returns sensible default download options.
func DefaultOptions() Options {
	return Options{
		ShowProgress:    true,
		FollowRedirects: true,
		Timeout:         5 * time.Minute,
		Retries:         3,
	}
}

// effectiveTimeout returns the timeout to use, applying the default if not set.
func (o Options) effectiveTimeout() time.Duration {
	if o.Timeout <= 0 {
		return 5 * time.Minute
	}
	return o.Timeout
}

// effectiveRetries returns the retry count to use, applying the default if not set.
func (o Options) effectiveRetries() int {
	if o.Retries <= 0 {
		return 3
	}
	return o.Retries
}

// FromURL downloads a file from a generic HTTP/HTTPS URL.
func FromURL(rawURL string, opts Options) error {
	if !strings.HasPrefix(rawURL, "http://") && !strings.HasPrefix(rawURL, "https://") {
		return fmt.Errorf("invalid URL (must start with http:// or https://): %s", rawURL)
	}

	client := &http.Client{
		Timeout: opts.effectiveTimeout(),
	}
	if !opts.FollowRedirects {
		client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}
	}

	// Build request with auth headers
	req, err := http.NewRequest("GET", rawURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	if opts.Token != "" {
		req.Header.Set("Authorization", "Bearer "+opts.Token)
	}
	for k, v := range opts.Headers {
		req.Header.Set(k, v)
	}

	// Retry loop with exponential backoff
	maxRetries := opts.effectiveRetries()
	var resp *http.Response
	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			backoff := time.Duration(1<<uint(attempt-1)) * time.Second
			time.Sleep(backoff)
			// Re-create request for retry (body already consumed)
			req, err = http.NewRequest("GET", rawURL, nil)
			if err != nil {
				return fmt.Errorf("failed to create request: %w", err)
			}
			if opts.Token != "" {
				req.Header.Set("Authorization", "Bearer "+opts.Token)
			}
			for k, v := range opts.Headers {
				req.Header.Set(k, v)
			}
		}

		resp, err = client.Do(req)
		if err != nil {
			if attempt < maxRetries {
				continue
			}
			return fmt.Errorf("failed to fetch URL: %w", err)
		}

		// Don't retry on client errors (4xx) except 429
		if resp.StatusCode == http.StatusTooManyRequests ||
			resp.StatusCode >= 500 {
			resp.Body.Close()
			if attempt < maxRetries {
				continue
			}
			return &ErrHTTP{StatusCode: resp.StatusCode, Status: resp.Status, URL: rawURL}
		}

		break
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return &ErrNotFound{URL: rawURL}
	}
	if resp.StatusCode != http.StatusOK {
		return &ErrHTTP{StatusCode: resp.StatusCode, Status: resp.Status, URL: rawURL}
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
			return &ErrFileExists{Path: outPath}
		}
	}

	if opts.ShowInfo {
		fmt.Printf("  URL:  %s\n", rawURL)
		fmt.Printf("  Size: %s\n", formatSize(resp.ContentLength))
		fmt.Printf("  Dest: %s\n", outPath)
	}

	if opts.ShowProgress {
		fmt.Printf("  Downloading → %s\n", outPath)
	}

	// Write atomically: write to temp file then rename
	if err := WriteAtomic(outPath, resp.Body); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	if opts.ShowProgress {
		fmt.Printf("  ✓ Saved: %s\n", outPath)
	}

	return nil
}

// WriteAtomic writes data from r to path atomically by writing to a temp file
// first and then renaming it to the final path. This prevents partial writes.
func WriteAtomic(path string, r io.Reader) error {
	// Ensure parent directory exists
	dir := filepath.Dir(path)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	// Create temp file in same directory as target (for atomic rename)
	tmpFile, err := os.CreateTemp(dir, ".download-*.tmp")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpPath := tmpFile.Name()

	// Clean up temp file on failure
	success := false
	defer func() {
		if !success {
			tmpFile.Close()
			os.Remove(tmpPath)
		}
	}()

	if _, err := io.Copy(tmpFile, r); err != nil {
		return fmt.Errorf("failed to write data: %w", err)
	}

	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("failed to close temp file: %w", err)
	}

	if err := os.Rename(tmpPath, path); err != nil {
		return fmt.Errorf("failed to rename temp file to %s: %w", path, err)
	}

	success = true
	return nil
}

// Multiple downloads multiple files from a list of URLs with bounded concurrency.
// At most 5 downloads run in parallel.
func Multiple(urls []string, opts Options) error {
	const maxParallel = 5

	type result struct {
		url string
		err error
	}

	results := make([]result, len(urls))
	sem := make(chan struct{}, maxParallel)
	var wg sync.WaitGroup

	for i, u := range urls {
		wg.Add(1)
		go func(idx int, url string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			fmt.Printf("[%d/%d] %s\n", idx+1, len(urls), url)
			err := FromURL(url, opts)
			results[idx] = result{url: url, err: err}
		}(i, u)
	}

	wg.Wait()

	var errs []string
	succeeded := 0
	for _, r := range results {
		if r.err != nil {
			errs = append(errs, fmt.Sprintf("%s: %v", r.url, r.err))
		} else {
			succeeded++
		}
	}

	fmt.Printf("\nSummary: %d succeeded, %d failed\n", succeeded, len(errs))

	if len(errs) > 0 {
		return fmt.Errorf("some downloads failed:\n%s", strings.Join(errs, "\n"))
	}
	return nil
}

// filenameFromURL extracts the filename from a URL path.
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

// formatSize returns a human-readable file size.
func formatSize(bytes int64) string {
	if bytes < 0 {
		return "unknown"
	}
	if bytes == 0 {
		return "0 B"
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
