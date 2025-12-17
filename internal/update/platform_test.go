package update

import (
	"testing"
)

func TestGetAssetName(t *testing.T) {
	tests := []struct {
		os       string
		arch     string
		expected string
	}{
		{"linux", "amd64", "ghex-linux-amd64.tar.gz"},
		{"linux", "arm64", "ghex-linux-arm64.tar.gz"},
		{"darwin", "amd64", "ghex-darwin-amd64.tar.gz"},
		{"darwin", "arm64", "ghex-darwin-arm64.tar.gz"},
		{"windows", "amd64", "ghex-windows-amd64.zip"},
		{"windows", "arm64", "ghex-windows-arm64.zip"},
	}

	for _, tt := range tests {
		t.Run(tt.os+"-"+tt.arch, func(t *testing.T) {
			got := GetAssetName(tt.os, tt.arch)
			if got != tt.expected {
				t.Errorf("GetAssetName(%q, %q) = %q, want %q", tt.os, tt.arch, got, tt.expected)
			}
		})
	}
}

func TestIsSupportedPlatform(t *testing.T) {
	tests := []struct {
		os       string
		arch     string
		expected bool
	}{
		{"linux", "amd64", true},
		{"linux", "arm64", true},
		{"darwin", "amd64", true},
		{"darwin", "arm64", true},
		{"windows", "amd64", true},
		{"windows", "arm64", true},
		{"freebsd", "amd64", false},
		{"linux", "386", false},
	}

	for _, tt := range tests {
		t.Run(tt.os+"-"+tt.arch, func(t *testing.T) {
			got := IsSupportedPlatform(tt.os, tt.arch)
			if got != tt.expected {
				t.Errorf("IsSupportedPlatform(%q, %q) = %v, want %v", tt.os, tt.arch, got, tt.expected)
			}
		})
	}
}

func TestSelectAssetForPlatform(t *testing.T) {
	release := &ReleaseInfo{
		Assets: []Asset{
			{Name: "ghex-linux-amd64.tar.gz", DownloadURL: "https://example.com/linux-amd64"},
			{Name: "ghex-darwin-arm64.tar.gz", DownloadURL: "https://example.com/darwin-arm64"},
			{Name: "ghex-windows-amd64.zip", DownloadURL: "https://example.com/windows-amd64"},
			{Name: "checksums.txt", DownloadURL: "https://example.com/checksums"},
		},
	}

	// Test found
	asset, err := SelectAssetForPlatform(release, "linux", "amd64")
	if err != nil {
		t.Errorf("SelectAssetForPlatform failed: %v", err)
	}
	if asset.Name != "ghex-linux-amd64.tar.gz" {
		t.Errorf("Expected ghex-linux-amd64.tar.gz, got %s", asset.Name)
	}

	// Test not found
	_, err = SelectAssetForPlatform(release, "freebsd", "amd64")
	if err == nil {
		t.Error("Expected error for unsupported platform")
	}
}

func TestGetPlatformDisplayName(t *testing.T) {
	tests := []struct {
		os       string
		arch     string
		expected string
	}{
		{"linux", "amd64", "Linux x64"},
		{"darwin", "arm64", "macOS ARM64"},
		{"windows", "amd64", "Windows x64"},
	}

	for _, tt := range tests {
		t.Run(tt.os+"-"+tt.arch, func(t *testing.T) {
			got := GetPlatformDisplayName(tt.os, tt.arch)
			if got != tt.expected {
				t.Errorf("GetPlatformDisplayName(%q, %q) = %q, want %q", tt.os, tt.arch, got, tt.expected)
			}
		})
	}
}
