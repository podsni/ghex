package commands

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/dwirx/ghex/internal/ui"
	"github.com/spf13/cobra"
)

// Version is set during build
var Version = "1.0.0"

// NewRootCmd creates the root command
func NewRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "ghex",
		Short: "Beautiful GitHub Account Switcher & Universal Downloader",
		Long:  "GHEX - Interactive CLI tool for managing multiple GitHub accounts per repository with universal download capabilities",
		Run: func(cmd *cobra.Command, args []string) {
			runInteractive()
		},
	}

	// Add all subcommands
	rootCmd.AddCommand(NewVersionCmd())
	rootCmd.AddCommand(NewStatusCmd())
	rootCmd.AddCommand(NewListCmd())
	rootCmd.AddCommand(NewSwitchCmd())
	rootCmd.AddCommand(NewHealthCmd())
	rootCmd.AddCommand(NewLogCmd())
	rootCmd.AddCommand(NewAddCmd())
	rootCmd.AddCommand(NewRemoveCmd())
	rootCmd.AddCommand(NewEditCmd())

	// SSH commands
	rootCmd.AddCommand(NewSSHCmd())
	rootCmd.AddCommand(NewGlobalSSHCmd())
	rootCmd.AddCommand(NewTestCmd())

	// Download commands (dlx)
	rootCmd.AddCommand(NewDlxCmd())

	// Update command
	rootCmd.AddCommand(NewUpdateCmd())

	// Git shortcuts
	AddGitShortcuts(rootCmd)

	return rootCmd
}

// Execute runs the root command
func Execute() {
	// Handle Ctrl+C gracefully
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		fmt.Println()
		ui.ShowSuccess("Thank you for using GHEX! 👋")
		os.Exit(0)
	}()

	rootCmd := NewRootCmd()

	// Handle URL arguments for clone
	if len(os.Args) > 1 {
		arg := os.Args[1]
		if isGitURL(arg) {
			targetDir := ""
			if len(os.Args) > 2 {
				targetDir = os.Args[2]
			}
			runClone(arg, targetDir)
			return
		}
	}

	if err := rootCmd.Execute(); err != nil {
		ui.ShowError(err.Error())
		os.Exit(1)
	}
}
