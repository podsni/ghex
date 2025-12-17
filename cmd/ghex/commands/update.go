package commands

import (
	"fmt"
	"strings"

	"github.com/dwirx/ghex/internal/ui"
	"github.com/dwirx/ghex/internal/update"
	"github.com/spf13/cobra"
)

var (
	updateCheck     bool
	updateChangelog bool
	updateRollback  bool
	updateForce     bool
	updateYes       bool
)

// NewUpdateCmd creates the update command
func NewUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update ghex to the latest version",
		Long:  "Check for updates, download and install the latest version of ghex",
		Run: func(cmd *cobra.Command, args []string) {
			runUpdate()
		},
	}

	cmd.Flags().BoolVarP(&updateCheck, "check", "c", false, "Check for updates without installing")
	cmd.Flags().BoolVar(&updateChangelog, "changelog", false, "Show changelog before updating")
	cmd.Flags().BoolVar(&updateRollback, "rollback", false, "Rollback to previous version")
	cmd.Flags().BoolVarP(&updateForce, "force", "f", false, "Force update without confirmation")
	cmd.Flags().BoolVarP(&updateYes, "yes", "y", false, "Auto-confirm prompts")

	return cmd
}

func runUpdate() {
	if updateRollback {
		runRollback()
		return
	}

	updater, err := update.NewUpdater(Version)
	if err != nil {
		ui.ShowError(fmt.Sprintf("Failed to initialize updater: %v", err))
		return
	}

	// Check for updates
	ui.ShowInfo("Checking for updates...")
	release, hasUpdate, err := updater.CheckForUpdate()
	if err != nil {
		ui.ShowError(fmt.Sprintf("Failed to check for updates: %v", err))
		return
	}

	if !hasUpdate {
		ui.ShowSuccess(fmt.Sprintf("You're already running the latest version (v%s)", Version))
		return
	}

	// Show update info
	fmt.Println()
	ui.ShowInfo(fmt.Sprintf("Current version: v%s", Version))
	ui.ShowSuccess(fmt.Sprintf("Latest version:  %s", release.TagName))
	fmt.Println()

	// Show changelog if requested
	if updateChangelog {
		showChangelog(updater)
	}

	// If only checking, stop here
	if updateCheck {
		ui.ShowInfo("Run 'ghex update' to install the latest version")
		return
	}

	// Confirm update
	if !updateForce && !updateYes {
		fmt.Print("Do you want to update? [y/N]: ")
		var response string
		fmt.Scanln(&response)
		response = strings.ToLower(strings.TrimSpace(response))
		if response != "y" && response != "yes" {
			ui.ShowInfo("Update cancelled")
			return
		}
	}

	// Perform update
	ui.ShowInfo("Downloading update...")
	err = updater.Update(release, func(current, total int64) {
		percent := float64(current) / float64(total) * 100
		fmt.Printf("\rDownloading: %.1f%% (%d/%d bytes)", percent, current, total)
	})
	fmt.Println() // New line after progress

	if err != nil {
		ui.ShowError(fmt.Sprintf("Update failed: %v", err))
		if updater.HasBackup() {
			ui.ShowInfo("You can rollback to the previous version with: ghex update --rollback")
		}
		return
	}

	ui.ShowSuccess(fmt.Sprintf("Successfully updated to %s!", release.TagName))
	ui.ShowInfo("Please restart ghex to use the new version")
}

func runRollback() {
	updater, err := update.NewUpdater(Version)
	if err != nil {
		ui.ShowError(fmt.Sprintf("Failed to initialize updater: %v", err))
		return
	}

	if !updater.HasBackup() {
		ui.ShowError("No backup available for rollback")
		return
	}

	// Confirm rollback
	if !updateForce && !updateYes {
		fmt.Print("Do you want to rollback to the previous version? [y/N]: ")
		var response string
		fmt.Scanln(&response)
		response = strings.ToLower(strings.TrimSpace(response))
		if response != "y" && response != "yes" {
			ui.ShowInfo("Rollback cancelled")
			return
		}
	}

	ui.ShowInfo("Rolling back to previous version...")
	if err := updater.Rollback(); err != nil {
		ui.ShowError(fmt.Sprintf("Rollback failed: %v", err))
		return
	}

	ui.ShowSuccess("Successfully rolled back to previous version!")
	ui.ShowInfo("Please restart ghex to use the restored version")
}

func showChangelog(updater *update.Updater) {
	releases, err := updater.GetChangelog(Version)
	if err != nil {
		ui.ShowError(fmt.Sprintf("Failed to fetch changelog: %v", err))
		return
	}

	if len(releases) == 0 {
		ui.ShowInfo("No changelog available")
		return
	}

	fmt.Println("\n📋 Changelog:")
	fmt.Println(strings.Repeat("-", 50))
	fmt.Println(update.FormatChangelog(releases))
	fmt.Println(strings.Repeat("-", 50))
}
