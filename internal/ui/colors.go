package ui

import "github.com/charmbracelet/lipgloss"

// Color definitions inspired by Charm
var (
	// Primary colors
	PrimaryColor   = lipgloss.Color("#F25D94")
	SecondaryColor = lipgloss.Color("#FF9F40")
	AccentColor    = lipgloss.Color("#00D9FF")

	// Status colors
	SuccessColor = lipgloss.Color("#00FF88")
	WarningColor = lipgloss.Color("#FFCC00")
	ErrorColor   = lipgloss.Color("#FF6B6B")
	MutedColor   = lipgloss.Color("#666666")

	// Text colors
	TextColor = lipgloss.Color("#FFFFFF")
	DimColor  = lipgloss.Color("#888888")
)

// Styles
var (
	// Text styles
	PrimaryStyle   = lipgloss.NewStyle().Foreground(PrimaryColor)
	SecondaryStyle = lipgloss.NewStyle().Foreground(SecondaryColor)
	AccentStyle    = lipgloss.NewStyle().Foreground(AccentColor)
	SuccessStyle   = lipgloss.NewStyle().Foreground(SuccessColor)
	WarningStyle   = lipgloss.NewStyle().Foreground(WarningColor)
	ErrorStyle     = lipgloss.NewStyle().Foreground(ErrorColor)
	MutedStyle     = lipgloss.NewStyle().Foreground(MutedColor)
	TextStyle      = lipgloss.NewStyle().Foreground(TextColor)
	DimStyle       = lipgloss.NewStyle().Foreground(DimColor)

	// Bold styles
	BoldPrimaryStyle = lipgloss.NewStyle().Foreground(PrimaryColor).Bold(true)
	BoldTextStyle    = lipgloss.NewStyle().Foreground(TextColor).Bold(true)

	// Box styles
	BoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(PrimaryColor).
			Padding(1, 2)

	SuccessBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(SuccessColor).
			Padding(1, 2)

	WarningBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(WarningColor).
			Padding(1, 2)

	ErrorBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ErrorColor).
			Padding(1, 2)

	// Section header style
	SectionStyle = lipgloss.NewStyle().
			Foreground(AccentColor).
			Bold(true)
)

// Color applies a color to text
func Color(text string, color lipgloss.Color) string {
	return lipgloss.NewStyle().Foreground(color).Render(text)
}

// Primary applies primary color to text
func Primary(text string) string {
	return PrimaryStyle.Render(text)
}

// Secondary applies secondary color to text
func Secondary(text string) string {
	return SecondaryStyle.Render(text)
}

// Accent applies accent color to text
func Accent(text string) string {
	return AccentStyle.Render(text)
}

// Success applies success color to text
func Success(text string) string {
	return SuccessStyle.Render(text)
}

// Warning applies warning color to text
func Warning(text string) string {
	return WarningStyle.Render(text)
}

// Error applies error color to text
func Error(text string) string {
	return ErrorStyle.Render(text)
}

// Muted applies muted color to text
func Muted(text string) string {
	return MutedStyle.Render(text)
}

// Dim applies dim color to text
func Dim(text string) string {
	return DimStyle.Render(text)
}

// Bold applies bold style to text
func Bold(text string) string {
	return BoldTextStyle.Render(text)
}
