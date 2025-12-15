package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// SelectorItem represents an item in the selector
type SelectorItem struct {
	Title       string
	Description string
	Value       string
}

// SelectorModel is the bubbletea model for interactive selection
type SelectorModel struct {
	items    []SelectorItem
	cursor   int
	selected int
	title    string
	done     bool
	canceled bool
}

// NewSelector creates a new selector model
func NewSelector(title string, items []SelectorItem) SelectorModel {
	return SelectorModel{
		items:    items,
		cursor:   0,
		selected: -1,
		title:    title,
	}
}

func (m SelectorModel) Init() tea.Cmd {
	return nil
}

func (m SelectorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			m.canceled = true
			m.done = true
			return m, tea.Quit

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			} else {
				m.cursor = len(m.items) - 1 // Wrap to bottom
			}

		case "down", "j":
			if m.cursor < len(m.items)-1 {
				m.cursor++
			} else {
				m.cursor = 0 // Wrap to top
			}

		case "home", "g":
			m.cursor = 0

		case "end", "G":
			m.cursor = len(m.items) - 1

		case "enter", " ", "l":
			m.selected = m.cursor
			m.done = true
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m SelectorModel) View() string {
	if m.done {
		return ""
	}

	var b strings.Builder

	// Title
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(PrimaryColor).
		MarginBottom(1)

	b.WriteString(titleStyle.Render(m.title))
	b.WriteString("\n\n")

	// Items
	for i, item := range m.items {
		cursor := "  "
		style := lipgloss.NewStyle().Foreground(TextColor)

		if i == m.cursor {
			cursor = "▸ "
			style = lipgloss.NewStyle().
				Foreground(AccentColor).
				Bold(true)
		}

		line := fmt.Sprintf("%s%s", cursor, item.Title)
		b.WriteString(style.Render(line))
		b.WriteString("\n")

		if item.Description != "" {
			descStyle := lipgloss.NewStyle().
				Foreground(MutedColor).
				MarginLeft(4)
			b.WriteString(descStyle.Render(item.Description))
			b.WriteString("\n")
		}
	}

	// Help
	helpStyle := lipgloss.NewStyle().
		Foreground(MutedColor).
		MarginTop(1)

	b.WriteString("\n")
	b.WriteString(helpStyle.Render("↑/k up • ↓/j down • enter/l select • q/esc cancel"))

	return b.String()
}

// Selected returns the selected index (-1 if canceled)
func (m SelectorModel) Selected() int {
	if m.canceled {
		return -1
	}
	return m.selected
}

// RunSelector runs the interactive selector and returns the selected index
func RunSelector(title string, items []SelectorItem) (int, error) {
	model := NewSelector(title, items)
	p := tea.NewProgram(model)

	finalModel, err := p.Run()
	if err != nil {
		return -1, err
	}

	return finalModel.(SelectorModel).Selected(), nil
}

// SelectFromStrings is a convenience function to select from string slice
func SelectFromStrings(title string, options []string) (int, string, error) {
	items := make([]SelectorItem, len(options))
	for i, opt := range options {
		items[i] = SelectorItem{Title: opt, Value: opt}
	}

	idx, err := RunSelector(title, items)
	if err != nil {
		return -1, "", err
	}

	if idx < 0 || idx >= len(options) {
		return -1, "", nil
	}

	return idx, options[idx], nil
}

// SelectSSHKey shows SSH key selector with auto-suggestions
func SelectSSHKey(keys []string, currentKey string) (int, string, error) {
	items := make([]SelectorItem, len(keys))
	for i, key := range keys {
		desc := ""
		if key == currentKey {
			desc = "Currently selected"
		}
		items[i] = SelectorItem{
			Title:       key,
			Description: desc,
			Value:       key,
		}
	}

	idx, err := RunSelector("Select SSH Key", items)
	if err != nil {
		return -1, "", err
	}

	if idx < 0 || idx >= len(keys) {
		return -1, "", nil
	}

	return idx, keys[idx], nil
}

// SelectAccount shows account selector
func SelectAccountInteractive(accounts []string, activeAccount string) (int, string, error) {
	items := make([]SelectorItem, len(accounts))
	for i, acc := range accounts {
		desc := ""
		if acc == activeAccount {
			desc = "● Active"
		}
		items[i] = SelectorItem{
			Title:       acc,
			Description: desc,
			Value:       acc,
		}
	}

	idx, err := RunSelector("Select Account", items)
	if err != nil {
		return -1, "", err
	}

	if idx < 0 || idx >= len(accounts) {
		return -1, "", nil
	}

	return idx, accounts[idx], nil
}

// SelectMethodInteractive shows method selector (SSH/Token)
func SelectMethodInteractive(hasSSH, hasToken bool) (string, error) {
	var items []SelectorItem

	if hasSSH {
		items = append(items, SelectorItem{
			Title:       "🔑 SSH",
			Description: "Use SSH key authentication",
			Value:       "ssh",
		})
	}
	if hasToken {
		items = append(items, SelectorItem{
			Title:       "🔐 Token (HTTPS)",
			Description: "Use Personal Access Token",
			Value:       "token",
		})
	}

	if len(items) == 0 {
		return "", fmt.Errorf("no authentication methods available")
	}

	if len(items) == 1 {
		return items[0].Value, nil
	}

	idx, err := RunSelector("Select Authentication Method", items)
	if err != nil {
		return "", err
	}

	if idx < 0 || idx >= len(items) {
		return "", nil
	}

	return items[idx].Value, nil
}

// SelectPlatformInteractive shows platform selector
func SelectPlatformInteractive() (string, error) {
	items := []SelectorItem{
		{Title: "🐙 GitHub", Description: "github.com", Value: "github"},
		{Title: "🦊 GitLab", Description: "gitlab.com", Value: "gitlab"},
		{Title: "🪣 Bitbucket", Description: "bitbucket.org", Value: "bitbucket"},
		{Title: "🍵 Gitea", Description: "Self-hosted Gitea", Value: "gitea"},
		{Title: "🌐 Other", Description: "Custom Git server", Value: "other"},
	}

	idx, err := RunSelector("Select Platform", items)
	if err != nil {
		return "", err
	}

	if idx < 0 || idx >= len(items) {
		return "", nil
	}

	return items[idx].Value, nil
}
