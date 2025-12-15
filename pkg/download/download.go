package download

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/dwirx/ghex/internal/platform"
	"github.com/dwirx/ghex/internal/ui"
)

// Options configures download behavior
type Options struct {
	Output          string
	OutputDir       string
	Overwrite       bool
	ShowProgress    bool
	ShowInfo        bool
	FollowRedirects bool
	UserAgent       string
	Headers         map[string]string
	Timeout         time.Duration
}

// DefaultOptions returns default download options
func DefaultOptions() Options {
	return Options{
		FollowRedirects: true,
		UserAgent:       "ghe/1.0",
		Timeout:         30 * time.Second,
	}
}

// FromURL downloads a file from a URL
func FromURL(url string, opts Options) error {
	// Create HTTP client
	client := &http.Client{
		Timeout: opts.Timeout,
	}

	if !opts.FollowRedirects {
		client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}
	}

	// Create request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	if opts.UserAgent != "" {
		req.Header.Set("User-Agent", opts.UserAgent)
	}
	for k, v := range opts.Headers {
		req.Header.Set(k, v)
	}

	// Show info if requested
	if opts.ShowInfo {
		ui.ShowInfo(fmt.Sprintf("URL: %s", url))
	}

	// Execute request
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	// Determine output filename
	filename := opts.Output
	if filename == "" {
		filename = getFilenameFromURL(url)
		if filename == "" {
			filename = getFilenameFromHeader(resp.Header.Get("Content-Disposition"))
		}
		if filename == "" {
			filename = "download"
		}
	}

	// Determine output path
	outputPath := filename
	if opts.OutputDir != "" {
		if err := platform.EnsureDir(opts.OutputDir, 0755); err != nil {
			return fmt.Errorf("failed to create output directory: %w", err)
		}
		outputPath = filepath.Join(opts.OutputDir, filename)
	}

	// Check if file exists
	if !opts.Overwrite && platform.FileExists(outputPath) {
		return fmt.Errorf("file already exists: %s (use --overwrite to replace)", outputPath)
	}

	// Create output file
	out, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer out.Close()

	// Download with progress
	var reader io.Reader = resp.Body
	if opts.ShowProgress && resp.ContentLength > 0 {
		reader = &progressReader{
			reader: resp.Body,
			total:  resp.ContentLength,
		}
	}

	written, err := io.Copy(out, reader)
	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	if opts.ShowProgress {
		fmt.Println() // New line after progress
	}

	ui.ShowSuccess(fmt.Sprintf("Downloaded: %s (%d bytes)", outputPath, written))
	return nil
}

// getFilenameFromURL extracts filename from URL
func getFilenameFromURL(url string) string {
	parts := strings.Split(url, "/")
	if len(parts) > 0 {
		filename := parts[len(parts)-1]
		// Remove query string
		if idx := strings.Index(filename, "?"); idx != -1 {
			filename = filename[:idx]
		}
		return filename
	}
	return ""
}

// getFilenameFromHeader extracts filename from Content-Disposition header
func getFilenameFromHeader(header string) string {
	if header == "" {
		return ""
	}
	// Parse Content-Disposition: attachment; filename="example.pdf"
	parts := strings.Split(header, ";")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if strings.HasPrefix(part, "filename=") {
			filename := strings.TrimPrefix(part, "filename=")
			filename = strings.Trim(filename, `"'`)
			return filename
		}
	}
	return ""
}

// progressReader wraps a reader to show download progress
type progressReader struct {
	reader  io.Reader
	total   int64
	current int64
}

func (pr *progressReader) Read(p []byte) (int, error) {
	n, err := pr.reader.Read(p)
	pr.current += int64(n)

	// Print progress
	percent := float64(pr.current) / float64(pr.total) * 100
	fmt.Printf("\rDownloading: %.1f%% (%d/%d bytes)", percent, pr.current, pr.total)

	return n, err
}

// Multiple downloads multiple files from URLs
func Multiple(urls []string, opts Options) error {
	for _, url := range urls {
		if err := FromURL(url, opts); err != nil {
			ui.ShowError(fmt.Sprintf("Failed to download %s: %v", url, err))
		}
	}
	return nil
}
