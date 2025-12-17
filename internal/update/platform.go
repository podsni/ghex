package update

import (
	"fmt"
	"runtime"
	"strings"
)

// SupportedPlatforms lists all supported OS/Arch combinations
var SupportedPlatforms = []struct {
	OS   string
	Arch string
}{
	{"linux", "amd64"},
	{"linux", "arm64"},
	{"darwin", "amd64"},
	{"darwin", "arm64"},
	{"windows", "amd64"},
	{"windows", "arm64"},
}

// GetAssetName constructs the asset filename for a given platform
func GetAssetName(os, arch string) string {
	ext := ".tar.gz"
	if os == "windows" {
		ext = ".zip"
	}
	return fmt.Sprintf("ghex-%s-%s%s", os, arch, ext)
}

// GetCurrentAssetName returns the asset name for the current platform
func GetCurrentAssetName() string {
	return GetAssetName(runtime.GOOS, runtime.GOARCH)
}

// SelectAsset finds the matching asset for the current platform from a release
func SelectAsset(release *ReleaseInfo) (*Asset, error) {
	return SelectAssetForPlatform(release, runtime.GOOS, runtime.GOARCH)
}

// SelectAssetForPlatform finds the matching asset for a specific platform
func SelectAssetForPlatform(release *ReleaseInfo, os, arch string) (*Asset, error) {
	expectedName := GetAssetName(os, arch)

	for i := range release.Assets {
		if release.Assets[i].Name == expectedName {
			return &release.Assets[i], nil
		}
	}

	// Try alternative naming patterns
	alternatives := []string{
		fmt.Sprintf("ghex_%s_%s", os, arch),
		fmt.Sprintf("ghex-%s-%s", os, arch),
	}

	for i := range release.Assets {
		for _, alt := range alternatives {
			if strings.HasPrefix(release.Assets[i].Name, alt) {
				return &release.Assets[i], nil
			}
		}
	}

	return nil, fmt.Errorf("%w: %s/%s", ErrAssetNotFound, os, arch)
}


// IsSupportedPlatform checks if the given OS/Arch combination is supported
func IsSupportedPlatform(os, arch string) bool {
	for _, p := range SupportedPlatforms {
		if p.OS == os && p.Arch == arch {
			return true
		}
	}
	return false
}

// GetCurrentPlatform returns the current OS and architecture
func GetCurrentPlatform() (os, arch string) {
	return runtime.GOOS, runtime.GOARCH
}

// GetPlatformDisplayName returns a human-readable platform name
func GetPlatformDisplayName(os, arch string) string {
	osName := os
	switch os {
	case "darwin":
		osName = "macOS"
	case "linux":
		osName = "Linux"
	case "windows":
		osName = "Windows"
	}

	archName := arch
	switch arch {
	case "amd64":
		archName = "x64"
	case "arm64":
		archName = "ARM64"
	}

	return fmt.Sprintf("%s %s", osName, archName)
}

// GetArchiveExtension returns the archive extension for a given OS
func GetArchiveExtension(os string) string {
	if os == "windows" {
		return ".zip"
	}
	return ".tar.gz"
}
