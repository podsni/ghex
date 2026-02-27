package commands

import (
	"fmt"
	"os"

	"github.com/dwirx/ghex/internal/config"
	"github.com/dwirx/ghex/internal/git"
	"github.com/dwirx/ghex/internal/platform"
	"github.com/dwirx/ghex/internal/ssh"
	"github.com/dwirx/ghex/internal/ui"
)

// PlatformInfo contains platform-specific information
type PlatformInfo struct {
	Host     string
	Name     string
	Icon     string
	Type     string
	KeysURL  string
	TokenURL string
}

// GetPlatformInfo returns platform information from account
func GetPlatformInfo(acc *config.Account) PlatformInfo {
	info := PlatformInfo{
		Host:     "github.com",
		Name:     "GitHub",
		Icon:     "🐙",
		Type:     "github",
		KeysURL:  "https://github.com/settings/keys",
		TokenURL: "https://github.com/settings/tokens",
	}

	if acc.Platform != nil {
		info.Type = acc.Platform.Type
		switch acc.Platform.Type {
		case "gitlab":
			info.Host = "gitlab.com"
			info.Name = "GitLab"
			info.Icon = "🦊"
			info.KeysURL = "https://gitlab.com/-/profile/keys"
			info.TokenURL = "https://gitlab.com/-/profile/personal_access_tokens"
		case "bitbucket":
			info.Host = "bitbucket.org"
			info.Name = "Bitbucket"
			info.Icon = "🪣"
			info.KeysURL = "https://bitbucket.org/account/settings/ssh-keys/"
			info.TokenURL = "https://bitbucket.org/account/settings/app-passwords/"
		case "gitea":
			info.Name = "Gitea"
			info.Icon = "🍵"
		case "codeberg":
			info.Host = "codeberg.org"
			info.Name = "Codeberg"
			info.Icon = "🏔️"
			info.KeysURL = "https://codeberg.org/user/settings/keys"
			info.TokenURL = "https://codeberg.org/user/settings/applications"
		}
		if acc.Platform.Domain != "" {
			info.Host = acc.Platform.Domain
		}
	}

	return info
}

// ExpandKeyPath expands ~ in key path to home directory
func ExpandKeyPath(keyPath string) string {
	return platform.ExpandPath(keyPath)
}

// TestAccountSSH tests SSH connection for an account and shows result
// Returns true if test passed
func TestAccountSSH(acc *config.Account, showDetails bool) bool {
	if acc.SSH == nil {
		ui.ShowWarning("Account has no SSH configuration")
		return false
	}

	platform := GetPlatformInfo(acc)
	keyPath := acc.SSH.KeyPath
	expandedPath := ExpandKeyPath(keyPath)

	// Check if key exists
	if _, err := os.Stat(expandedPath); os.IsNotExist(err) {
		ui.ShowError(fmt.Sprintf("SSH key not found: %s", keyPath))
		return false
	}

	// Fix permissions for ALL SSH keys
	fixedCount, _ := ssh.FixAllKeyPermissions()
	if fixedCount > 0 && showDetails {
		ui.ShowSuccess(fmt.Sprintf("✓ Fixed permissions for %d SSH key(s)", fixedCount))
	}

	if showDetails {
		fmt.Println()
		ui.ShowInfo(fmt.Sprintf("🔑 Using key: %s", keyPath))
		ui.ShowInfo(fmt.Sprintf("🌐 Host: %s %s (%s)", platform.Icon, platform.Name, platform.Host))
		fmt.Println()
	}

	spinner := ui.NewSpinner("Testing SSH connection...")
	spinner.Start()

	ok, msg, _ := ssh.TestConnectionWithKey(platform.Host, expandedPath)
	if ok {
		spinner.StopWithSuccess("✓ SSH connection test passed!")
		if showDetails {
			ui.ShowSuccess(fmt.Sprintf("Authenticated successfully to %s", platform.Host))
		}
		return true
	}

	spinner.StopWithError("✗ SSH connection test failed!")
	if showDetails {
		fmt.Println()
		ui.ShowWarning(fmt.Sprintf("Make sure your SSH key is added to %s:", platform.Name))
		ui.ShowInfo(fmt.Sprintf("1. Copy your public key: cat %s.pub", keyPath))
		ui.ShowInfo(fmt.Sprintf("2. Add it at: %s", platform.KeysURL))
		if msg != "" {
			fmt.Println()
			fmt.Println(ui.Muted(fmt.Sprintf("Details: %s", msg)))
		}
	}
	return false
}

// TestAccountToken tests token authentication for an account and shows result
// Returns true if test passed
func TestAccountToken(acc *config.Account, showDetails bool) bool {
	if acc.Token == nil {
		ui.ShowWarning("Account has no token configuration")
		return false
	}

	platformInfo := GetPlatformInfo(acc)

	spinner := ui.NewSpinner("Testing token authentication...")
	spinner.Start()

	ok, msg, _ := git.TestTokenAuthForHost(acc.Token.Username, acc.Token.Token, platformInfo.Host)
	if ok {
		spinner.StopWithSuccess("✓ Token authentication test passed!")
		if showDetails {
			ui.ShowInfo(fmt.Sprintf("Successfully authenticated as %s", acc.Token.Username))
		}
		return true
	}

	spinner.StopWithError("✗ Token authentication failed!")
	if showDetails {
		ui.ShowWarning("Please check:")
		ui.ShowInfo("• Token has not expired")
		ui.ShowInfo("• Token has correct permissions (repo access)")
		ui.ShowInfo("• Username is correct")
		ui.ShowInfo(fmt.Sprintf("\nCreate a new token at: %s", platformInfo.TokenURL))
		if msg != "" {
			fmt.Println(ui.Muted(fmt.Sprintf("\nDetails: %s", msg)))
		}
	}
	return false
}

// TestSSHKeyDirect tests an SSH key directly against a host
// Returns true if test passed
func TestSSHKeyDirect(keyPath, host string, showDetails bool) bool {
	expandedPath := ExpandKeyPath(keyPath)

	// Check if key exists
	if _, err := os.Stat(expandedPath); os.IsNotExist(err) {
		ui.ShowError(fmt.Sprintf("SSH key not found: %s", keyPath))
		return false
	}

	// Fix permissions for ALL SSH keys
	fixedCount, _ := ssh.FixAllKeyPermissions()
	if fixedCount > 0 && showDetails {
		ui.ShowSuccess(fmt.Sprintf("✓ Fixed permissions for %d SSH key(s)", fixedCount))
	}

	if showDetails {
		fmt.Println()
		ui.ShowInfo(fmt.Sprintf("🔑 Using key: %s", keyPath))
		ui.ShowInfo(fmt.Sprintf("🌐 Host: %s", host))
		fmt.Println()
	}

	spinner := ui.NewSpinner("Testing SSH connection...")
	spinner.Start()

	ok, msg, _ := ssh.TestConnectionWithKey(host, expandedPath)
	if ok {
		spinner.StopWithSuccess("✓ SSH connection test passed!")
		if showDetails {
			ui.ShowSuccess(fmt.Sprintf("Authenticated successfully to %s", host))
		}
		return true
	}

	spinner.StopWithError("✗ SSH connection test failed!")
	if showDetails {
		fmt.Println()
		ui.ShowWarning(fmt.Sprintf("Make sure your SSH key is added to %s:", host))
		ui.ShowInfo(fmt.Sprintf("1. Copy your public key: cat %s.pub", keyPath))
		ui.ShowInfo("2. Add it to your Git service settings")
		if msg != "" {
			fmt.Println()
			fmt.Println(ui.Muted(fmt.Sprintf("Details: %s", msg)))
		}
	}
	return false
}
