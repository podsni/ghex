package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// ASCII art title for GHEX
const asciiTitle = `
  ██████╗ ██╗  ██╗███████╗██╗  ██╗
 ██╔════╝ ██║  ██║██╔════╝╚██╗██╔╝
 ██║  ███╗███████║█████╗   ╚███╔╝ 
 ██║   ██║██╔══██║██╔══╝   ██╔██╗ 
 ╚██████╔╝██║  ██║███████╗██╔╝ ██╗
  ╚═════╝ ╚═╝  ╚═╝╚══════╝╚═╝  ╚═╝`

// ShowTitle displays the application title
func ShowTitle() {
	// Apply gradient-like effect using primary and secondary colors
	lines := strings.Split(asciiTitle, "\n")
	for i, line := range lines {
		if i < len(lines)/2 {
			fmt.Println(PrimaryStyle.Render(line))
		} else {
			fmt.Println(SecondaryStyle.Render(line))
		}
	}
	fmt.Println(MutedStyle.Render("✨ GitHub Account Switcher & Universal Downloader ✨"))
	fmt.Println()
}

// BoxOptions configures box display
type BoxOptions struct {
	Title   string
	Type    string // "info", "success", "warning", "error"
	Padding int
}

// ShowBox displays content in a styled box
func ShowBox(content string, opts BoxOptions) {
	var style lipgloss.Style

	switch opts.Type {
	case "success":
		style = SuccessBoxStyle
	case "warning":
		style = WarningBoxStyle
	case "error":
		style = ErrorBoxStyle
	default:
		style = BoxStyle
	}

	if opts.Title != "" {
		titleStyle := lipgloss.NewStyle().Bold(true)
		switch opts.Type {
		case "success":
			titleStyle = titleStyle.Foreground(SuccessColor)
		case "warning":
			titleStyle = titleStyle.Foreground(WarningColor)
		case "error":
			titleStyle = titleStyle.Foreground(ErrorColor)
		default:
			titleStyle = titleStyle.Foreground(PrimaryColor)
		}
		fmt.Println(titleStyle.Render("┌─ " + opts.Title + " ─┐"))
	}

	fmt.Println(style.Render(content))
	fmt.Println()
}

// ShowSuccess displays a success message
func ShowSuccess(message string) {
	fmt.Println(SuccessStyle.Render("✓ ") + TextStyle.Render(message))
}

// ShowError displays an error message
func ShowError(message string) {
	fmt.Println(ErrorStyle.Render("✗ ") + TextStyle.Render(message))
}

// ShowWarning displays a warning message
func ShowWarning(message string) {
	fmt.Println(WarningStyle.Render("⚠ ") + TextStyle.Render(message))
}

// ShowInfo displays an info message
func ShowInfo(message string) {
	fmt.Println(AccentStyle.Render("ℹ ") + TextStyle.Render(message))
}

// ShowSection displays a section header
func ShowSection(title string) {
	fmt.Println()
	fmt.Println(SectionStyle.Render("▶ " + title))
	fmt.Println(MutedStyle.Render(strings.Repeat("─", 50)))
	fmt.Println()
}

// ShowSeparator displays a separator line
func ShowSeparator() {
	fmt.Println(MutedStyle.Render(strings.Repeat("─", 60)))
}

// ShowList displays a list of items
func ShowList(items []string) {
	for i, item := range items {
		bullet := AccentStyle.Render("●")
		index := DimStyle.Render(fmt.Sprintf("[%d]", i+1))
		fmt.Printf("  %s %s %s\n", bullet, index, TextStyle.Render(item))
	}
	fmt.Println()
}

// ShowKeyValue displays a key-value pair
func ShowKeyValue(key, value string) {
	fmt.Printf("%s: %s\n", AccentStyle.Render(key), TextStyle.Render(value))
}

// ShowIndentedKeyValue displays an indented key-value pair
func ShowIndentedKeyValue(key, value string, indent int) {
	prefix := strings.Repeat("  ", indent)
	fmt.Printf("%s%s: %s\n", prefix, MutedStyle.Render(key), TextStyle.Render(value))
}

// Confirm prompts for yes/no confirmation
func Confirm(message string) bool {
	fmt.Printf("%s %s [y/N]: ", PrimaryStyle.Render("◉"), TextStyle.Render(message))
	var response string
	fmt.Scanln(&response)
	response = strings.ToLower(strings.TrimSpace(response))
	return response == "y" || response == "yes"
}

// Prompt prompts for text input
func Prompt(message string) string {
	fmt.Printf("%s %s: ", PrimaryStyle.Render("◇"), TextStyle.Render(message))
	var response string
	fmt.Scanln(&response)
	return strings.TrimSpace(response)
}

// PromptWithDefault prompts for text input with a default value
func PromptWithDefault(message, defaultValue string) string {
	if defaultValue != "" {
		fmt.Printf("%s %s [%s]: ", PrimaryStyle.Render("◇"), TextStyle.Render(message), DimStyle.Render(defaultValue))
	} else {
		fmt.Printf("%s %s: ", PrimaryStyle.Render("◇"), TextStyle.Render(message))
	}
	var response string
	fmt.Scanln(&response)
	response = strings.TrimSpace(response)
	if response == "" {
		return defaultValue
	}
	return response
}

// PromptPassword prompts for password input (note: this doesn't hide input in basic implementation)
func PromptPassword(message string) string {
	fmt.Printf("%s %s: ", PrimaryStyle.Render("◇"), TextStyle.Render(message))
	var response string
	fmt.Scanln(&response)
	return strings.TrimSpace(response)
}
