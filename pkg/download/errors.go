// Package download provides utilities for downloading files from URLs and Git repositories.
package download

import "fmt"

// ErrNotFound is returned when a resource is not found (HTTP 404).
type ErrNotFound struct {
	URL string
}

// Error implements the error interface.
func (e *ErrNotFound) Error() string {
	return fmt.Sprintf("not found: %s", e.URL)
}

// ErrRateLimit is returned when GitHub API rate limit is exceeded.
type ErrRateLimit struct {
	ResetAt string
}

// Error implements the error interface.
func (e *ErrRateLimit) Error() string {
	msg := "GitHub API rate limit exceeded"
	if e.ResetAt != "" {
		msg += fmt.Sprintf(" (resets at %s)", e.ResetAt)
	}
	return msg + ". Set GITHUB_TOKEN environment variable to increase limits."
}

// ErrHTTP is returned for unexpected HTTP status codes.
type ErrHTTP struct {
	StatusCode int
	Status     string
	URL        string
}

// Error implements the error interface.
func (e *ErrHTTP) Error() string {
	return fmt.Sprintf("HTTP %d %s: %s", e.StatusCode, e.Status, e.URL)
}

// ErrFileExists is returned when a file already exists and overwrite is not enabled.
type ErrFileExists struct {
	Path string
}

// Error implements the error interface.
func (e *ErrFileExists) Error() string {
	return fmt.Sprintf("file already exists: %s (use --overwrite to replace)", e.Path)
}
