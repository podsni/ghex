package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/dwirx/ghex/internal/config"
	"github.com/dwirx/ghex/internal/platform"
	"github.com/dwirx/ghex/internal/ssh"
	"github.com/dwirx/ghex/internal/ui"
	"github.com/spf13/cobra"
)

// NewGlobalSSHCmd creates the global SSH switch command
func NewGlobalSSHCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "global-ssh",
		Short: "Switch SSH globally",
		Long:  "Change global SSH configuration for github.com or other platforms",
		Run: func(cmd *cobra.Command, args []string) {
			cfg, _ := config.Load()
			runSwitchGlobalSSH(cfg)
		},
	}
}

// NewTestCmd creates the test connection command
func NewTestCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "test",
		Short: "Test SSH/Token connection",
		Long:  "Test SSH key or token authentication for an account",
		Run: func(cmd *cobra.Command, args []string) {
			cfg, _ := config.Load()
			runTestConnection(cfg)
		},
	}
}

// NewSSHCmd creates the SSH command group
func NewSSHCmd() *cobra.Command {
	sshCmd := &cobra.Command{
		Use:   "ssh",
		Short: "SSH key management",
		Run: func(cmd *cobra.Command, args []string) {
			cfg, _ := config.Load()
			runSSHMenu(cfg)
		},
	}

	sshCmd.AddCommand(&cobra.Command{
		Use:   "generate",
		Short: "Generate a new SSH key",
		Run: func(cmd *cobra.Command, args []string) {
			cfg, _ := config.Load()
			runGenerateSSHKey(cfg)
		},
	})

	sshCmd.AddCommand(&cobra.Command{
		Use:   "import",
		Short: "Import an existing SSH key",
		Run: func(cmd *cobra.Command, args []string) {
			cfg, _ := config.Load()
			runImportSSHKey(cfg)
		},
	})

	sshCmd.AddCommand(&cobra.Command{
		Use:   "test",
		Short: "Test SSH connection",
		Run: func(cmd *cobra.Command, args []string) {
			cfg, _ := config.Load()
			runTestConnection(cfg)
		},
	})

	sshCmd.AddCommand(&cobra.Command{
		Use:   "global",
		Short: "Switch SSH globally",
		Run: func(cmd *cobra.Command, args []string) {
			cfg, _ := config.Load()
			runSwitchGlobalSSH(cfg)
		},
	})

	sshCmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List SSH keys",
		Run: func(cmd *cobra.Command, args []string) {
			runListSSHKeys()
		},
	})

	return sshCmd
}

func runSSHMenu(cfg *config.AppConfig) {
	items := []ui.SelectorItem{
		{Title: "🔑 Generate SSH key", Description: "Create a new Ed25519 SSH key pair", Value: "generate"},
		{Title: "📥 Import SSH key", Description: "Import an existing private key", Value: "import"},
		{Title: "🌐 Switch SSH globally", Description: "Set default SSH key for github.com", Value: "global"},
		{Title: "🧪 Test connection", Description: "Test SSH authentication", Value: "test"},
		{Title: "📋 List SSH keys", Description: "Show all SSH keys in ~/.ssh", Value: "list"},
		{Title: "🔙 Back", Description: "Return to main menu", Value: "back"},
	}

	idx, err := ui.RunSelector("SSH Management", items)
	if err != nil || idx < 0 {
		return
	}

	switch items[idx].Value {
	case "generate":
		runGenerateSSHKey(cfg)
	case "import":
		runImportSSHKey(cfg)
	case "global":
		runSwitchGlobalSSH(cfg)
	case "test":
		runTestConnection(cfg)
	case "list":
		runListSSHKeys()
	case "back":
		return
	}
}

func runGenerateSSHKey(cfg *config.AppConfig) {
	if len(cfg.Accounts) == 0 {
		ui.ShowWarning("No accounts configured. Add an account first.")
		return
	}

	// Build items for selector
	items := make([]ui.SelectorItem, len(cfg.Accounts))
	for i, acc := range cfg.Accounts {
		desc := ""
		if acc.SSH != nil {
			desc = acc.SSH.KeyPath
		} else {
			desc = "No SSH configured"
		}
		items[i] = ui.SelectorItem{
			Title:       acc.Name,
			Description: desc,
			Value:       acc.Name,
		}
	}

	idx, err := ui.RunSelector("Select Account for SSH Key Generation", items)
	if err != nil {
		ui.ShowError(fmt.Sprintf("Selection error: %v", err))
		return
	}
	if idx < 0 {
		ui.ShowInfo("Cancelled")
		return
	}

	acc := &cfg.Accounts[idx]
	if acc.SSH == nil {
		ui.ShowWarning("Account has no SSH configuration")
		return
	}

	comment := acc.GitEmail
	if comment == "" {
		comment = acc.GitUserName
	}
	if comment == "" {
		comment = fmt.Sprintf("%s@github", acc.Name)
	}

	fmt.Println()
	spinner := ui.NewSpinner("Generating SSH key...")
	spinner.Start()

	if err := ssh.GenerateKey(acc.SSH.KeyPath, comment); err != nil {
		spinner.StopWithError(fmt.Sprintf("Failed to generate key: %v", err))
		return
	}

	spinner.StopWithSuccess(fmt.Sprintf("Generated SSH key: %s", acc.SSH.KeyPath))
	ui.ShowInfo(fmt.Sprintf("Public key: %s.pub", acc.SSH.KeyPath))
}

func runImportSSHKey(cfg *config.AppConfig) {
	if len(cfg.Accounts) == 0 {
		ui.ShowWarning("No accounts configured. Add an account first.")
		return
	}

	// Build items for selector
	items := make([]ui.SelectorItem, len(cfg.Accounts))
	for i, acc := range cfg.Accounts {
		desc := ""
		if acc.SSH != nil {
			desc = acc.SSH.KeyPath
		} else {
			desc = "No SSH configured"
		}
		items[i] = ui.SelectorItem{
			Title:       acc.Name,
			Description: desc,
			Value:       acc.Name,
		}
	}

	idx, err := ui.RunSelector("Select Account for SSH Key Import", items)
	if err != nil {
		ui.ShowError(fmt.Sprintf("Selection error: %v", err))
		return
	}
	if idx < 0 {
		ui.ShowInfo("Cancelled")
		return
	}

	acc := &cfg.Accounts[idx]

	// Show existing SSH keys for selection
	existingKeys, _ := ssh.ListPrivateKeys()
	var srcPath string

	if len(existingKeys) > 0 {
		keyItems := make([]ui.SelectorItem, len(existingKeys)+1)
		for i, key := range existingKeys {
			keyItems[i] = ui.SelectorItem{
				Title: key,
				Value: key,
			}
		}
		keyItems[len(existingKeys)] = ui.SelectorItem{
			Title:       "📝 Enter custom path",
			Description: "Type a new SSH key path",
			Value:       "__custom__",
		}

		keyIdx, err := ui.RunSelector("Select Source SSH Key", keyItems)
		if err != nil || keyIdx < 0 {
			ui.ShowInfo("Cancelled")
			return
		}

		if keyItems[keyIdx].Value == "__custom__" {
			srcPath = ui.Prompt("Source private key path")
		} else {
			srcPath = keyItems[keyIdx].Value
		}
	} else {
		srcPath = ui.Prompt("Source private key path")
	}

	if srcPath == "" {
		ui.ShowError("Source path is required")
		return
	}

	destName := ui.PromptWithDefault("Destination filename", fmt.Sprintf("id_ed25519_%s", acc.Name))
	sshDir := platform.GetSSHDir()
	destPath := filepath.Join(sshDir, destName)

	if err := ssh.ImportKey(srcPath, destPath); err != nil {
		ui.ShowError(fmt.Sprintf("Failed to import key: %v", err))
		return
	}

	if acc.SSH == nil {
		acc.SSH = &config.SshConfig{}
	}
	acc.SSH.KeyPath = destPath

	// Ask if user wants to set as default
	if ui.Confirm("Set as default SSH key for github.com?") {
		host := "github.com"
		if acc.Platform != nil && acc.Platform.Domain != "" {
			host = acc.Platform.Domain
		}
		if err := ssh.EnsureConfigBlock(host, destPath, host); err != nil {
			ui.ShowWarning(fmt.Sprintf("Failed to configure SSH: %v", err))
		} else {
			ui.ShowSuccess(fmt.Sprintf("Set as default Host %s", host))
		}
	}

	if err := config.Save(cfg); err != nil {
		ui.ShowWarning(fmt.Sprintf("Failed to save config: %v", err))
	}

	ui.ShowSuccess(fmt.Sprintf("Imported SSH key: %s", destPath))

	pubPath, err := ssh.EnsurePublicKey(destPath)
	if err == nil {
		ui.ShowInfo(fmt.Sprintf("Public key: %s", pubPath))
	}

	// Ask if user wants to test connection
	if ui.Confirm("Test SSH connection now?") {
		host := "github.com"
		if acc.Platform != nil && acc.Platform.Domain != "" {
			host = acc.Platform.Domain
		}

		// Expand destPath for testing
		expandedDest := platform.ExpandPath(destPath)

		// Auto-fix permissions for ALL keys
		fixedCount, _ := ssh.FixAllKeyPermissions()
		if fixedCount > 0 {
			ui.ShowInfo(fmt.Sprintf("Fixed permissions for %d SSH key(s)", fixedCount))
		}

		ui.ShowInfo(fmt.Sprintf("Testing with key: %s", destPath))
		spinner := ui.NewSpinner(fmt.Sprintf("Testing SSH connection to %s...", host))
		spinner.Start()

		ok, msg, _ := ssh.TestConnectionWithKey(host, expandedDest)
		if ok {
			spinner.StopWithSuccess(fmt.Sprintf("SSH: %s", msg))
		} else {
			spinner.StopWithError(fmt.Sprintf("SSH: %s", msg))
			ui.ShowWarning("Make sure your SSH key is added to your Git service:")
			ui.ShowInfo(fmt.Sprintf("1. Copy your public key: cat %s.pub", destPath))
			ui.ShowInfo("2. Add it to your Git service settings")
		}
	}
}

func runSwitchGlobalSSH(cfg *config.AppConfig) {
	if len(cfg.Accounts) == 0 {
		ui.ShowWarning("No accounts configured")
		return
	}

	var sshAccounts []config.Account
	for _, acc := range cfg.Accounts {
		if acc.SSH != nil {
			sshAccounts = append(sshAccounts, acc)
		}
	}

	if len(sshAccounts) == 0 {
		// Show available SSH keys instead
		keys, _ := ssh.ListPrivateKeys()
		if len(keys) == 0 {
			ui.ShowWarning("No SSH keys found")
			return
		}

		items := make([]ui.SelectorItem, len(keys))
		for i, key := range keys {
			items[i] = ui.SelectorItem{
				Title: key,
				Value: key,
			}
		}

		idx, err := ui.RunSelector("Select SSH Key for Global Use", items)
		if err != nil || idx < 0 {
			return
		}

		if err := ssh.EnsureConfigBlock("github.com", keys[idx], "github.com"); err != nil {
			ui.ShowError(fmt.Sprintf("Failed to configure SSH: %v", err))
			return
		}

		ui.ShowSuccess(fmt.Sprintf("Set global SSH to: %s", keys[idx]))

		// Ask to test connection
		if ui.Confirm("Test SSH connection now?") {
			// Auto-fix permissions for ALL keys
			fixedCount, _ := ssh.FixAllKeyPermissions()
			if fixedCount > 0 {
				ui.ShowInfo(fmt.Sprintf("Fixed permissions for %d SSH key(s)", fixedCount))
			}

			ui.ShowInfo(fmt.Sprintf("Testing with key: %s", keys[idx]))
			spinner := ui.NewSpinner("Testing SSH connection to github.com...")
			spinner.Start()

			ok, msg, _ := ssh.TestConnectionWithKey("github.com", keys[idx])
			if ok {
				spinner.StopWithSuccess(fmt.Sprintf("SSH: %s", msg))
			} else {
				spinner.StopWithError(fmt.Sprintf("SSH: %s", msg))
				ui.ShowWarning("Make sure your SSH key is added to GitHub:")
				ui.ShowInfo(fmt.Sprintf("1. Copy your public key: cat %s.pub", keys[idx]))
				ui.ShowInfo("2. Add it at: https://github.com/settings/keys")
			}
		}
		return
	}

	// Build items for selector
	items := make([]ui.SelectorItem, len(sshAccounts))
	for i, acc := range sshAccounts {
		platformName := "GitHub"
		if acc.Platform != nil {
			switch acc.Platform.Type {
			case "gitlab":
				platformName = "GitLab"
			case "bitbucket":
				platformName = "Bitbucket"
			case "gitea":
				platformName = "Gitea"
			case "codeberg":
				platformName = "Codeberg"
			}
		}
		items[i] = ui.SelectorItem{
			Title:       acc.Name,
			Description: fmt.Sprintf("%s • %s", platformName, acc.SSH.KeyPath),
			Value:       acc.Name,
		}
	}

	idx, err := ui.RunSelector("Select Account for Global SSH", items)
	if err != nil {
		ui.ShowError(fmt.Sprintf("Selection error: %v", err))
		return
	}
	if idx < 0 {
		ui.ShowInfo("Cancelled")
		return
	}

	acc := sshAccounts[idx]

	// Get platform-specific host
	host := "github.com"
	platformName := "GitHub"
	platformIcon := "🐙"
	if acc.Platform != nil {
		switch acc.Platform.Type {
		case "gitlab":
			host = "gitlab.com"
			platformName = "GitLab"
			platformIcon = "🦊"
		case "bitbucket":
			host = "bitbucket.org"
			platformName = "Bitbucket"
			platformIcon = "🪣"
		case "gitea":
			platformName = "Gitea"
			platformIcon = "🍵"
		case "codeberg":
			host = "codeberg.org"
			platformName = "Codeberg"
			platformIcon = "🏔️"
		}
		if acc.Platform.Domain != "" {
			host = acc.Platform.Domain
		}
	}

	keyPath := acc.SSH.KeyPath

	// Check if key exists
	expandedPath := platform.ExpandPath(keyPath)

	if _, err := os.Stat(expandedPath); os.IsNotExist(err) {
		if ui.Confirm(fmt.Sprintf("SSH key not found at %s. Generate now?", keyPath)) {
			comment := acc.GitEmail
			if comment == "" {
				comment = acc.GitUserName
			}
			if comment == "" {
				platformType := "github"
				if acc.Platform != nil {
					platformType = acc.Platform.Type
				}
				comment = fmt.Sprintf("%s@%s", acc.Name, platformType)
			}

			spinner := ui.NewSpinner("Generating SSH key...")
			spinner.Start()

			if err := ssh.GenerateKey(keyPath, comment); err != nil {
				spinner.StopWithError(fmt.Sprintf("Failed to generate key: %v", err))
				return
			}
			spinner.StopWithSuccess(fmt.Sprintf("Generated SSH key: %s", keyPath))
		} else {
			ui.ShowInfo("Aborted")
			return
		}
	}

	fmt.Println()
	if err := ssh.EnsureConfigBlock(host, keyPath, host); err != nil {
		ui.ShowError(fmt.Sprintf("Failed to configure SSH: %v", err))
		return
	}

	ui.ShowSuccess(fmt.Sprintf("Updated ~/.ssh/config → Host %s %s (%s) using: %s", platformIcon, platformName, host, keyPath))

	// Ask to test connection
	if ui.Confirm("Test SSH connection now?") {
		// Auto-fix permissions for ALL keys
		fixedCount, _ := ssh.FixAllKeyPermissions()
		if fixedCount > 0 {
			ui.ShowInfo(fmt.Sprintf("Fixed permissions for %d SSH key(s)", fixedCount))
		}

		ui.ShowInfo(fmt.Sprintf("Testing with key: %s", keyPath))
		spinner := ui.NewSpinner(fmt.Sprintf("Testing SSH connection to %s (%s)...", platformName, host))
		spinner.Start()

		ok, msg, _ := ssh.TestConnectionWithKey(host, expandedPath)
		if ok {
			spinner.StopWithSuccess(fmt.Sprintf("SSH: %s", msg))
		} else {
			spinner.StopWithError(fmt.Sprintf("SSH: %s", msg))
			ui.ShowWarning(fmt.Sprintf("Make sure your SSH key is added to %s:", platformName))
			ui.ShowInfo(fmt.Sprintf("1. Copy your public key: cat %s.pub", keyPath))
			platformType := "github"
			if acc.Platform != nil {
				platformType = acc.Platform.Type
			}
			switch platformType {
			case "gitlab":
				ui.ShowInfo("2. Add it at: https://gitlab.com/-/profile/keys")
			case "bitbucket":
				ui.ShowInfo("2. Add it at: https://bitbucket.org/account/settings/ssh-keys/")
			case "codeberg":
				ui.ShowInfo("2. Add it at: https://codeberg.org/user/settings/keys")
			case "gitea":
				ui.ShowInfo("2. Add it at your Gitea instance: /user/settings/keys")
			default:
				ui.ShowInfo("2. Add it at: https://github.com/settings/keys")
			}
		}
	}
}

func runTestConnection(cfg *config.AppConfig) {
	ui.ShowSection("Test Connection")

	// Fix permissions for ALL SSH keys first
	fixedCount, _ := ssh.FixAllKeyPermissions()
	if fixedCount > 0 {
		ui.ShowInfo(fmt.Sprintf("Fixed permissions for %d SSH key(s)", fixedCount))
	}

	// If no accounts, offer to test SSH keys directly
	if len(cfg.Accounts) == 0 {
		keys, _ := ssh.ListPrivateKeys()
		if len(keys) == 0 {
			ui.ShowWarning("No accounts or SSH keys found")
			return
		}

		ui.ShowInfo("No accounts configured. Testing SSH keys directly...")
		testSSHKeyDirectly(keys)
		return
	}

	// Build items for selector - add option to test SSH key directly
	items := make([]ui.SelectorItem, len(cfg.Accounts)+1)

	// Add "Test SSH key directly" option first
	items[0] = ui.SelectorItem{
		Title:       "🔑 Test SSH key directly",
		Description: "Select any SSH key from ~/.ssh to test",
		Value:       "__direct__",
	}

	for i, acc := range cfg.Accounts {
		methods := []string{}
		if acc.SSH != nil {
			methods = append(methods, "🔑 SSH")
		}
		if acc.Token != nil {
			methods = append(methods, "🔐 Token")
		}
		platformName := "GitHub"
		platformIcon := "🐙"
		if acc.Platform != nil {
			switch acc.Platform.Type {
			case "gitlab":
				platformName = "GitLab"
				platformIcon = "🦊"
			case "bitbucket":
				platformName = "Bitbucket"
				platformIcon = "🪣"
			case "gitea":
				platformName = "Gitea"
				platformIcon = "🍵"
			case "codeberg":
				platformName = "Codeberg"
				platformIcon = "🏔️"
			}
		}
		items[i+1] = ui.SelectorItem{
			Title:       acc.Name,
			Description: fmt.Sprintf("%s %s • %s", platformIcon, platformName, strings.Join(methods, ", ")),
			Value:       acc.Name,
		}
	}

	idx, err := ui.RunSelector("Select Account to Test", items)
	if err != nil {
		ui.ShowError(fmt.Sprintf("Selection error: %v", err))
		return
	}
	if idx < 0 {
		ui.ShowInfo("Cancelled")
		return
	}

	// Check if user selected "Test SSH key directly"
	if items[idx].Value == "__direct__" {
		keys, _ := ssh.ListPrivateKeys()
		if len(keys) == 0 {
			ui.ShowWarning("No SSH keys found in ~/.ssh")
			return
		}
		testSSHKeyDirectly(keys)
		return
	}

	// Get the account (index is offset by 1 because of the direct test option)
	acc := cfg.Accounts[idx-1]

	// Get platform info
	host := "github.com"
	platformName := "GitHub"
	platformIcon := "🐙"
	if acc.Platform != nil {
		switch acc.Platform.Type {
		case "gitlab":
			host = "gitlab.com"
			platformName = "GitLab"
			platformIcon = "🦊"
		case "bitbucket":
			host = "bitbucket.org"
			platformName = "Bitbucket"
			platformIcon = "🪣"
		case "gitea":
			platformName = "Gitea"
			platformIcon = "🍵"
		case "codeberg":
			host = "codeberg.org"
			platformName = "Codeberg"
			platformIcon = "🏔️"
		}
		if acc.Platform.Domain != "" {
			host = acc.Platform.Domain
		}
	}

	// If both methods available, ask which to test
	if acc.SSH != nil && acc.Token != nil {
		methodItems := []ui.SelectorItem{
			{Title: "🔑 SSH", Description: "Test SSH key authentication", Value: "ssh"},
			{Title: "🔐 Token", Description: "Test Personal Access Token", Value: "token"},
			{Title: "🔄 Both", Description: "Test both methods", Value: "both"},
		}

		methodIdx, err := ui.RunSelector("Test which authentication method?", methodItems)
		if err != nil || methodIdx < 0 {
			ui.ShowInfo("Cancelled")
			return
		}

		fmt.Println()

		switch methodItems[methodIdx].Value {
		case "ssh":
			testSSHConnection(acc, host, platformName, platformIcon)
		case "token":
			testTokenConnection(acc, platformName)
		case "both":
			testSSHConnection(acc, host, platformName, platformIcon)
			fmt.Println()
			testTokenConnection(acc, platformName)
		}
		return
	}

	fmt.Println()

	if acc.SSH != nil {
		testSSHConnection(acc, host, platformName, platformIcon)
	}

	if acc.Token != nil {
		testTokenConnection(acc, platformName)
	}
}

// testSSHConnection uses helper function to test SSH connection
func testSSHConnection(acc config.Account, host, platformName, platformIcon string) {
	TestAccountSSH(&acc, true)
}

// testTokenConnection uses helper function to test token connection
func testTokenConnection(acc config.Account, platformName string) {
	TestAccountToken(&acc, true)
}

func runListSSHKeys() {
	keys, err := ssh.ListPrivateKeys()
	if err != nil {
		ui.ShowError(fmt.Sprintf("Failed to list SSH keys: %v", err))
		return
	}

	if len(keys) == 0 {
		ui.ShowWarning("No SSH keys found in ~/.ssh")
		return
	}

	ui.ShowSection("SSH Keys")
	for _, key := range keys {
		fmt.Printf("  • %s\n", ui.Accent(key))
	}
	fmt.Println()
	ui.ShowInfo(fmt.Sprintf("Total: %d keys", len(keys)))
}

// testSSHKeyDirectly allows testing any SSH key directly without an account
func testSSHKeyDirectly(keys []string) {
	// Build items for selector
	items := make([]ui.SelectorItem, len(keys))
	for i, key := range keys {
		items[i] = ui.SelectorItem{
			Title: key,
			Value: key,
		}
	}

	idx, err := ui.RunSelector("Select SSH Key to Test", items)
	if err != nil || idx < 0 {
		ui.ShowInfo("Cancelled")
		return
	}

	selectedKey := keys[idx]

	// Select platform/host to test
	hostItems := []ui.SelectorItem{
		{Title: "🐙 GitHub", Description: "github.com", Value: "github.com"},
		{Title: "🦊 GitLab", Description: "gitlab.com", Value: "gitlab.com"},
		{Title: "🪣 Bitbucket", Description: "bitbucket.org", Value: "bitbucket.org"},
		{Title: "🏔️ Codeberg", Description: "codeberg.org", Value: "codeberg.org"},
		{Title: "🌐 Custom", Description: "Enter custom host", Value: "__custom__"},
	}

	hostIdx, err := ui.RunSelector("Select Host to Test", hostItems)
	if err != nil || hostIdx < 0 {
		ui.ShowInfo("Cancelled")
		return
	}

	host := hostItems[hostIdx].Value
	if host == "__custom__" {
		host = ui.PromptWithDefault("Enter host", "github.com")
		if host == "" {
			host = "github.com"
		}
	}

	// Use helper function
	TestSSHKeyDirect(selectedKey, host, true)
}
