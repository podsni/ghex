package commands

import (
	"fmt"

	"github.com/dwirx/ghex/internal/account"
	"github.com/dwirx/ghex/internal/config"
	"github.com/dwirx/ghex/internal/git"
	"github.com/dwirx/ghex/internal/ssh"
	"github.com/dwirx/ghex/internal/ui"
	"github.com/spf13/cobra"
)

// NewHealthCmd creates the health command
func NewHealthCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "health",
		Short: "Check health of all accounts",
		Run: func(cmd *cobra.Command, args []string) {
			runHealthCheck()
		},
	}
}

// NewLogCmd creates the log command
func NewLogCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "log",
		Short: "Show activity log",
		Run: func(cmd *cobra.Command, args []string) {
			runActivityLog()
		},
	}
}

func runHealthCheck() {
	cfg, err := config.Load()
	if err != nil {
		ui.ShowError(fmt.Sprintf("Failed to load config: %v", err))
		return
	}

	if len(cfg.Accounts) == 0 {
		ui.ShowWarning("No accounts configured")
		return
	}

	ui.ShowSection("Health Check")

	// Fix permissions for ALL SSH keys first
	fixedCount, _ := ssh.FixAllKeyPermissions()
	if fixedCount > 0 {
		ui.ShowSuccess(fmt.Sprintf("✓ Fixed permissions for %d SSH key(s)", fixedCount))
	}

	// Track summary
	total := len(cfg.Accounts)
	healthy := 0
	warnings := 0
	errors := 0

	for _, acc := range cfg.Accounts {
		// Get platform info using helper
		platform := GetPlatformInfo(&acc)

		fmt.Printf("\n%s %s %s (%s)\n", ui.Primary("Checking:"), acc.Name, platform.Icon, platform.Name)

		accountHealthy := true

		if acc.SSH != nil {
			expandedPath := ExpandKeyPath(acc.SSH.KeyPath)

			spinner := ui.NewSpinner(fmt.Sprintf("  Testing SSH with %s...", acc.SSH.KeyPath))
			spinner.Start()

			ok, msg, _ := ssh.TestConnectionWithKey(platform.Host, expandedPath)
			if ok {
				spinner.StopWithSuccess(fmt.Sprintf("  SSH: %s", msg))
			} else {
				spinner.StopWithError(fmt.Sprintf("  SSH: %s", msg))
				accountHealthy = false
			}
		}

		if acc.Token != nil {
			spinner := ui.NewSpinner("  Testing Token...")
			spinner.Start()

			// Determine API host based on account platform
			apiHost := "github.com"
			if acc.Platform != nil && acc.Platform.Domain != "" {
				apiHost = acc.Platform.Domain
			}
			ok, msg, _ := git.TestTokenAuthForHost(acc.Token.Username, acc.Token.Token, apiHost)
			if ok {
				spinner.StopWithSuccess(fmt.Sprintf("  Token: %s", msg))
			} else {
				spinner.StopWithError(fmt.Sprintf("  Token: %s", msg))
				accountHealthy = false
			}
		}

		if accountHealthy {
			healthy++
		} else if acc.SSH != nil && acc.Token != nil {
			warnings++
		} else {
			errors++
		}
	}

	// Show summary
	fmt.Println()
	ui.ShowSeparator()
	fmt.Println()
	fmt.Printf("%s Total: %d | %s Healthy: %d | %s Warnings: %d | %s Errors: %d\n",
		ui.Primary("📊"),
		total,
		ui.Success("✓"),
		healthy,
		ui.Warning("⚠"),
		warnings,
		ui.Error("✗"),
		errors,
	)
}

func runActivityLog() {
	cfg, err := config.Load()
	if err != nil {
		ui.ShowError(fmt.Sprintf("Failed to load config: %v", err))
		return
	}

	if len(cfg.ActivityLog) == 0 {
		ui.ShowInfo("No activity logged yet")
		return
	}

	ui.ShowSection("Activity Log")

	manager := account.NewManager(cfg)
	entries := manager.GetRecentActivity(20)

	for _, entry := range entries {
		status := ui.Success("✓")
		if !entry.Success {
			status = ui.Error("✗")
		}

		fmt.Printf("%s %s %s %s",
			status,
			ui.Dim(entry.Timestamp[:19]),
			ui.Primary(entry.Action),
			entry.AccountName,
		)

		if entry.RepoPath != "" {
			fmt.Printf(" → %s", ui.Accent(entry.RepoPath))
		}
		if entry.Method != "" {
			fmt.Printf(" (%s)", entry.Method)
		}
		fmt.Println()
	}
}
