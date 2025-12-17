package update

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

const (
	DefaultRepoOwner = "dwirx"
	DefaultRepoName  = "ghex"
	DefaultBinaryName = "ghex"
)

// Updater handles self-update operations
type Updater struct {
	CurrentVersion string
	RepoOwner      string
	RepoName       string
	BinaryName     string
	Client         *GitHubClient
	BinaryManager  *BinaryManager
}

// NewUpdater creates a new Updater instance
func NewUpdater(currentVersion string) (*Updater, error) {
	bm, err := NewBinaryManager()
	if err != nil {
		return nil, err
	}

	return &Updater{
		CurrentVersion: currentVersion,
		RepoOwner:      DefaultRepoOwner,
		RepoName:       DefaultRepoName,
		BinaryName:     DefaultBinaryName,
		Client:         NewGitHubClient(),
		BinaryManager:  bm,
	}, nil
}

// CheckForUpdate checks if a newer version is available
func (u *Updater) CheckForUpdate() (*ReleaseInfo, bool, error) {
	release, err := u.Client.GetLatestRelease(u.RepoOwner, u.RepoName)
	if err != nil {
		return nil, false, err
	}

	currentVer, err := ParseVersion(u.CurrentVersion)
	if err != nil {
		return release, false, fmt.Errorf("failed to parse current version: %w", err)
	}

	latestVer, err := ParseVersion(release.TagName)
	if err != nil {
		return release, false, fmt.Errorf("failed to parse latest version: %w", err)
	}

	hasUpdate := latestVer.IsNewerThan(currentVer)
	return release, hasUpdate, nil
}


// Update downloads and installs the latest version
func (u *Updater) Update(release *ReleaseInfo, progress ProgressCallback) error {
	// Check write permission
	if err := CheckWritePermission(u.BinaryManager.BinaryPath); err != nil {
		return err
	}

	// Select asset for current platform
	asset, err := SelectAsset(release)
	if err != nil {
		return err
	}

	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "ghex-update-*")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	// Download asset
	archivePath := filepath.Join(tmpDir, asset.Name)
	if err := u.Client.DownloadAsset(asset, archivePath, progress); err != nil {
		return err
	}

	// Download and verify checksum if available
	checksumContent, err := u.Client.DownloadChecksums(release)
	if err == nil && checksumContent != "" {
		entries, err := ParseChecksumFile(checksumContent)
		if err == nil {
			if expectedChecksum, found := FindChecksum(entries, asset.Name); found {
				if err := VerifyChecksum(archivePath, expectedChecksum); err != nil {
					return err
				}
			}
		}
	}

	// Extract binary from archive
	binaryPath, err := u.extractBinary(archivePath, tmpDir)
	if err != nil {
		return err
	}

	// Backup current binary
	if err := u.BinaryManager.Backup(); err != nil {
		return err
	}

	// Replace binary
	if err := u.BinaryManager.Replace(binaryPath); err != nil {
		// Try to restore from backup
		u.BinaryManager.Restore()
		return err
	}

	return nil
}

// extractBinary extracts the binary from the downloaded archive
func (u *Updater) extractBinary(archivePath, destDir string) (string, error) {
	if strings.HasSuffix(archivePath, ".zip") {
		return u.extractZip(archivePath, destDir)
	}
	return u.extractTarGz(archivePath, destDir)
}


// extractTarGz extracts a .tar.gz archive
func (u *Updater) extractTarGz(archivePath, destDir string) (string, error) {
	f, err := os.Open(archivePath)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrExtractFailed, err)
	}
	defer f.Close()

	gzr, err := gzip.NewReader(f)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrExtractFailed, err)
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)
	binaryName := u.BinaryName
	if runtime.GOOS == "windows" {
		binaryName += ".exe"
	}

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", fmt.Errorf("%w: %v", ErrExtractFailed, err)
		}

		// Look for the binary file
		baseName := filepath.Base(header.Name)
		if baseName == binaryName || baseName == u.BinaryName {
			destPath := filepath.Join(destDir, binaryName)
			outFile, err := os.Create(destPath)
			if err != nil {
				return "", fmt.Errorf("%w: %v", ErrExtractFailed, err)
			}

			if _, err := io.Copy(outFile, tr); err != nil {
				outFile.Close()
				return "", fmt.Errorf("%w: %v", ErrExtractFailed, err)
			}
			outFile.Close()

			return destPath, nil
		}
	}

	return "", fmt.Errorf("%w: binary not found in archive", ErrExtractFailed)
}

// extractZip extracts a .zip archive
func (u *Updater) extractZip(archivePath, destDir string) (string, error) {
	r, err := zip.OpenReader(archivePath)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrExtractFailed, err)
	}
	defer r.Close()

	binaryName := u.BinaryName
	if runtime.GOOS == "windows" {
		binaryName += ".exe"
	}

	for _, f := range r.File {
		baseName := filepath.Base(f.Name)
		if baseName == binaryName || baseName == u.BinaryName+".exe" {
			rc, err := f.Open()
			if err != nil {
				return "", fmt.Errorf("%w: %v", ErrExtractFailed, err)
			}

			destPath := filepath.Join(destDir, binaryName)
			outFile, err := os.Create(destPath)
			if err != nil {
				rc.Close()
				return "", fmt.Errorf("%w: %v", ErrExtractFailed, err)
			}

			if _, err := io.Copy(outFile, rc); err != nil {
				outFile.Close()
				rc.Close()
				return "", fmt.Errorf("%w: %v", ErrExtractFailed, err)
			}

			outFile.Close()
			rc.Close()
			return destPath, nil
		}
	}

	return "", fmt.Errorf("%w: binary not found in archive", ErrExtractFailed)
}


// Rollback restores the previous version from backup
func (u *Updater) Rollback() error {
	return u.BinaryManager.Restore()
}

// HasBackup checks if a backup exists for rollback
func (u *Updater) HasBackup() bool {
	return u.BinaryManager.HasBackup()
}

// GetChangelog fetches release notes between current version and latest
func (u *Updater) GetChangelog(fromVersion string) ([]ReleaseInfo, error) {
	releases, err := u.Client.GetReleases(u.RepoOwner, u.RepoName, 20)
	if err != nil {
		return nil, err
	}

	fromVer, err := ParseVersion(fromVersion)
	if err != nil {
		return nil, err
	}

	var changelog []ReleaseInfo
	for _, release := range releases {
		releaseVer, err := ParseVersion(release.TagName)
		if err != nil {
			continue
		}

		// Include releases newer than fromVersion
		if releaseVer.IsNewerThan(fromVer) {
			changelog = append(changelog, release)
		}
	}

	return changelog, nil
}

// FormatChangelog formats release notes for terminal display
func FormatChangelog(releases []ReleaseInfo) string {
	if len(releases) == 0 {
		return "No changes found."
	}

	var sb strings.Builder
	for _, release := range releases {
		sb.WriteString(fmt.Sprintf("\n## %s (%s)\n", release.Name, release.TagName))
		sb.WriteString(release.Body)
		sb.WriteString("\n")
	}

	return sb.String()
}
