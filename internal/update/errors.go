// Package update provides self-update functionality for ghex
package update

import "errors"

// Error types for update operations
var (
	ErrNoUpdateAvailable = errors.New("already running the latest version")
	ErrDownloadFailed    = errors.New("failed to download update")
	ErrChecksumMismatch  = errors.New("checksum verification failed - possible security issue")
	ErrPermissionDenied  = errors.New("insufficient permissions to update")
	ErrBackupFailed      = errors.New("failed to create backup")
	ErrRestoreFailed     = errors.New("failed to restore from backup")
	ErrNoBackupAvailable = errors.New("no backup available for rollback")
	ErrInvalidVersion    = errors.New("invalid version format")
	ErrAssetNotFound     = errors.New("no compatible asset found for this platform")
	ErrNetworkError      = errors.New("network error while contacting GitHub")
	ErrExtractFailed     = errors.New("failed to extract downloaded archive")
)
