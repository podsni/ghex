package commands

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/dwirx/ghex/internal/ui"
	"github.com/dwirx/ghex/pkg/download"
	"github.com/spf13/cobra"
)

// NewDlxCmd creates the dlx (download) command group
func NewDlxCmd() *cobra.Command {
	dlxCmd := &cobra.Command{
		Use:   "dlx [url]",
		Short: "Universal file downloader",
		Long: `Download files from any URL (HTTP/HTTPS) or GitHub repositories.

GitHub URL formats supported:
  File:   https://github.com/{owner}/{repo}/blob/{branch}/{path}
  Folder: https://github.com/{owner}/{repo}/tree/{branch}/{path}

Examples:
  ghex dlx https://github.com/user/repo/blob/main/README.md
  ghex dlx https://github.com/user/repo/tree/main/src/
  ghex dlx https://example.com/file.tar.gz`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				output, _ := cmd.Flags().GetString("output")
				outputDir, _ := cmd.Flags().GetString("dir")
				overwrite, _ := cmd.Flags().GetBool("overwrite")
				showInfo, _ := cmd.Flags().GetBool("info")
				token, _ := cmd.Flags().GetString("token")
				if token == "" {
					token = os.Getenv("GITHUB_TOKEN")
				}

				rawURL := args[0]

				// Auto-detect GitHub URLs and route to the appropriate downloader
				if isGitHubURL(rawURL) {
					if err := runGitHubDownload(rawURL, output, outputDir, showInfo, overwrite, token); err != nil {
						ui.ShowError(err.Error())
						return err
					}
					return nil
				}

				// Generic HTTP/HTTPS download
				opts := download.Options{
					Output:          output,
					OutputDir:       outputDir,
					Overwrite:       overwrite,
					ShowProgress:    true,
					ShowInfo:        showInfo,
					FollowRedirects: true,
					Token:           token,
				}
				if err := download.FromURL(rawURL, opts); err != nil {
					ui.ShowError(err.Error())
					return err
				}
				return nil
			}
			runDlxMenu()
			return nil
		},
	}

	// Flags
	dlxCmd.Flags().StringP("output", "o", "", "Output filename")
	dlxCmd.Flags().StringP("dir", "d", "", "Output directory")
	dlxCmd.Flags().BoolP("overwrite", "w", false, "Overwrite existing files")
	dlxCmd.Flags().BoolP("info", "i", false, "Show file info before download")
	dlxCmd.Flags().StringP("token", "t", "", "GitHub personal access token (falls back to GITHUB_TOKEN env var)")

	// Subcommands
	dlxCmd.AddCommand(newDlxFileCmd())
	dlxCmd.AddCommand(newDlxDirCmd())
	dlxCmd.AddCommand(newDlxReleaseCmd())
	dlxCmd.AddCommand(newDlxListCmd())

	return dlxCmd
}

func newDlxFileCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "file [url]",
		Short: "Download a single file from Git repository",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			branch, _ := cmd.Flags().GetString("branch")
			output, _ := cmd.Flags().GetString("output")
			outputDir, _ := cmd.Flags().GetString("dir")
			overwrite, _ := cmd.Flags().GetBool("overwrite")
			showInfo, _ := cmd.Flags().GetBool("info")
			token, _ := cmd.Flags().GetString("token")
			if token == "" {
				token = os.Getenv("GITHUB_TOKEN")
			}

			opts := download.GitOptions{
				Branch:    branch,
				Output:    output,
				OutputDir: outputDir,
				Overwrite: overwrite,
				ShowInfo:  showInfo,
				Token:     token,
			}
			if err := download.GitFile(args[0], opts); err != nil {
				ui.ShowError(err.Error())
				return err
			}
			return nil
		},
	}

	cmd.Flags().StringP("branch", "b", "", "Branch/tag/commit")
	cmd.Flags().StringP("output", "o", "", "Output filename")
	cmd.Flags().StringP("dir", "d", "", "Output directory")
	cmd.Flags().BoolP("overwrite", "w", false, "Overwrite existing files")
	cmd.Flags().BoolP("info", "i", false, "Show file info before download")
	cmd.Flags().StringP("token", "t", "", "GitHub personal access token (falls back to GITHUB_TOKEN env var)")

	return cmd
}

func newDlxDirCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dir [url]",
		Short: "Download a directory from Git repository",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			branch, _ := cmd.Flags().GetString("branch")
			outputDir, _ := cmd.Flags().GetString("dir")
			depth, _ := cmd.Flags().GetInt("depth")
			overwrite, _ := cmd.Flags().GetBool("overwrite")
			showInfo, _ := cmd.Flags().GetBool("info")
			token, _ := cmd.Flags().GetString("token")
			if token == "" {
				token = os.Getenv("GITHUB_TOKEN")
			}

			opts := download.GitOptions{
				Branch:    branch,
				OutputDir: outputDir,
				Depth:     depth,
				Overwrite: overwrite,
				ShowInfo:  showInfo,
				Token:     token,
			}
			if err := download.GitDirectory(args[0], opts); err != nil {
				ui.ShowError(err.Error())
				return err
			}
			return nil
		},
	}

	cmd.Flags().StringP("branch", "b", "", "Branch/tag/commit")
	cmd.Flags().StringP("dir", "d", "", "Output directory")
	cmd.Flags().IntP("depth", "n", 100, "Max directory depth (0 = unlimited)")
	cmd.Flags().BoolP("overwrite", "w", false, "Overwrite existing files")
	cmd.Flags().BoolP("info", "i", false, "Show file info before download")
	cmd.Flags().StringP("token", "t", "", "GitHub personal access token (falls back to GITHUB_TOKEN env var)")

	return cmd
}

func newDlxReleaseCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "release [repo-url]",
		Short: "Download release assets from GitHub",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			version, _ := cmd.Flags().GetString("version")
			asset, _ := cmd.Flags().GetString("asset")
			outputDir, _ := cmd.Flags().GetString("dir")
			listOnly, _ := cmd.Flags().GetBool("list")
			overwrite, _ := cmd.Flags().GetBool("overwrite")
			token, _ := cmd.Flags().GetString("token")
			if token == "" {
				token = os.Getenv("GITHUB_TOKEN")
			}

			opts := download.ReleaseOptions{
				Version:   version,
				Asset:     asset,
				OutputDir: outputDir,
				ListOnly:  listOnly,
				Overwrite: overwrite,
				Token:     token,
			}
			if err := download.GitRelease(args[0], opts); err != nil {
				ui.ShowError(err.Error())
				return err
			}
			return nil
		},
	}

	cmd.Flags().StringP("version", "v", "", "Release version/tag (default: latest)")
	cmd.Flags().StringP("asset", "a", "", "Asset name filter")
	cmd.Flags().StringP("dir", "d", "", "Output directory")
	cmd.Flags().BoolP("list", "l", false, "List assets only")
	cmd.Flags().BoolP("overwrite", "w", false, "Overwrite existing files")
	cmd.Flags().StringP("token", "t", "", "GitHub personal access token (falls back to GITHUB_TOKEN env var)")

	return cmd
}

func newDlxListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list [file]",
		Short: "Download files from a URL list file",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			parallel, _ := cmd.Flags().GetInt("parallel")
			if err := downloadFromFileList(args[0], parallel); err != nil {
				ui.ShowError(err.Error())
				return err
			}
			return nil
		},
	}

	cmd.Flags().IntP("parallel", "p", 5, "Number of parallel downloads")

	return cmd
}

// isGitHubURL returns true if the URL is a GitHub repository URL.
func isGitHubURL(url string) bool {
	return strings.HasPrefix(url, "https://github.com/") ||
		strings.HasPrefix(url, "http://github.com/")
}

// runGitHubDownload auto-detects whether the GitHub URL points to a file (blob)
// or a directory (tree) and downloads accordingly.
// When downloading a file like https://github.com/owner/repo/blob/main/skill/SKILL.md
// the folder structure (skill/SKILL.md) is preserved in the output directory.
func runGitHubDownload(rawURL, output, outputDir string, showInfo, overwrite bool, token string) error {
	isTree := strings.Contains(rawURL, "/tree/")
	isBlob := strings.Contains(rawURL, "/blob/")

	if isBlob {
		// Single file download — preserve folder structure from repo path
		if showInfo {
			ui.ShowInfo(fmt.Sprintf("Downloading file from GitHub: %s", rawURL))
		}
		opts := download.GitOptions{
			Output:    output,    // empty = use repo path (preserves folder structure)
			OutputDir: outputDir, // base output directory
			Overwrite: overwrite,
			ShowInfo:  showInfo,
			Token:     token,
		}
		return download.GitFile(rawURL, opts)
	}

	if isTree {
		// Directory download
		if showInfo {
			ui.ShowInfo(fmt.Sprintf("Downloading directory from GitHub: %s", rawURL))
		}
		opts := download.GitOptions{
			OutputDir: outputDir,
			Depth:     100, // allow deep directories
			Overwrite: overwrite,
			ShowInfo:  showInfo,
			Token:     token,
		}
		return download.GitDirectory(rawURL, opts)
	}

	// Repo root or unknown GitHub URL — treat as directory download
	if showInfo {
		ui.ShowInfo(fmt.Sprintf("Downloading from GitHub: %s", rawURL))
	}
	opts := download.GitOptions{
		OutputDir: outputDir,
		Depth:     100,
		Overwrite: overwrite,
		ShowInfo:  showInfo,
		Token:     token,
	}
	return download.GitDirectory(rawURL, opts)
}

// downloadFromFileList reads URLs from a file and downloads them with bounded concurrency.
func downloadFromFileList(filePath string, parallel int) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file list: %w", err)
	}

	// Normalize line endings (handle Windows \r\n)
	content := strings.ReplaceAll(string(data), "\r\n", "\n")
	content = strings.ReplaceAll(content, "\r", "\n")
	lines := strings.Split(content, "\n")
	var urls []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "#") {
			urls = append(urls, line)
		}
	}

	if len(urls) == 0 {
		return fmt.Errorf("no URLs found in file list")
	}

	opts := download.DefaultOptions()
	opts.ShowProgress = true
	if parallel > 0 {
		// parallel hint is passed via the bounded concurrency in Multiple()
		// Multiple() uses its own internal cap of 5; this is informational
		_ = parallel
	}
	return download.Multiple(urls, opts)
}

// promptLine reads a full line from stdin, supporting spaces in input.
func promptLine(message string) string {
	fmt.Printf("%s %s: ", ui.Primary("◇"), ui.Muted(message))
	reader := bufio.NewReader(os.Stdin)
	line, _ := reader.ReadString('\n')
	return strings.TrimSpace(line)
}

func runDlxMenu() {
	ui.ShowSection("Download (dlx)")

	options := []string{
		"📥 Download from URL",
		"📄 Download file from Git repo",
		"📁 Download directory from Git repo",
		"🏷️  Download release assets",
		"📋 Download from URL list",
		"🔙 Back to main menu",
	}

	fmt.Println(ui.Primary("Choose an action:"))
	for i, opt := range options {
		fmt.Printf("  %s %s\n", ui.Dim(fmt.Sprintf("[%d]", i+1)), opt)
	}
	fmt.Println()

	choice := promptLine("Enter choice (1-6)")

	switch choice {
	case "1":
		runDownloadURL()
	case "2":
		runDownloadGitFile()
	case "3":
		runDownloadGitDir()
	case "4":
		runDownloadRelease()
	case "5":
		runDownloadFromList()
	case "6":
		return
	default:
		ui.ShowWarning("Invalid choice")
	}
}

func runDownloadURL() {
	url := promptLine("Enter URL to download")
	if url == "" {
		ui.ShowError("URL is required")
		return
	}

	output := promptLine("Output filename (optional, press Enter for auto)")
	outputDir := promptLine("Output directory (optional, press Enter for current)")

	opts := download.Options{
		Output:          output,
		OutputDir:       outputDir,
		ShowProgress:    true,
		FollowRedirects: true,
	}

	if err := download.FromURL(url, opts); err != nil {
		ui.ShowError(err.Error())
	}
}

func runDownloadGitFile() {
	url := promptLine("Enter Git file URL (e.g., https://github.com/user/repo/blob/main/file.txt)")
	if url == "" {
		ui.ShowError("URL is required")
		return
	}

	branch := promptLine("Branch/tag/commit (optional, press Enter for default)")
	output := promptLine("Output filename (optional)")

	opts := download.GitOptions{
		Branch: branch,
		Output: output,
	}

	if err := download.GitFile(url, opts); err != nil {
		ui.ShowError(err.Error())
	}
}

func runDownloadGitDir() {
	url := promptLine("Enter Git directory URL (e.g., https://github.com/user/repo/tree/main/src)")
	if url == "" {
		ui.ShowError("URL is required")
		return
	}

	branch := promptLine("Branch/tag/commit (optional)")
	outputDir := promptLine("Output directory (optional)")

	opts := download.GitOptions{
		Branch:    branch,
		OutputDir: outputDir,
		Depth:     100,
	}

	if err := download.GitDirectory(url, opts); err != nil {
		ui.ShowError(err.Error())
	}
}

func runDownloadRelease() {
	url := promptLine("Enter GitHub repo URL (e.g., https://github.com/user/repo)")
	if url == "" {
		ui.ShowError("URL is required")
		return
	}

	version := promptLine("Version/tag (optional, press Enter for latest)")
	asset := promptLine("Asset name filter (optional)")
	outputDir := promptLine("Output directory (optional)")

	opts := download.ReleaseOptions{
		Version:   version,
		Asset:     asset,
		OutputDir: outputDir,
	}

	if err := download.GitRelease(url, opts); err != nil {
		ui.ShowError(err.Error())
	}
}

func runDownloadFromList() {
	filePath := promptLine("Enter path to URL list file")
	if filePath == "" {
		ui.ShowError("File path is required")
		return
	}

	if err := downloadFromFileList(filePath, 5); err != nil {
		ui.ShowError(err.Error())
	}
}
