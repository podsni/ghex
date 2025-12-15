package ui

import (
	"fmt"
	"strings"
)

// MenuItem represents a menu item
type MenuItem struct {
	Title       string
	Description string
	Value       string
}

// SelectMenu displays a simple menu and returns the selected index
func SelectMenu(title string, items []MenuItem) (int, error) {
	fmt.Println()
	fmt.Println(BoldPrimaryStyle.Render(title))
	fmt.Println(MutedStyle.Render(strings.Repeat("─", 50)))
	fmt.Println()

	for i, item := range items {
		prefix := MutedStyle.Render(fmt.Sprintf("[%d]", i+1))
		fmt.Printf("  %s %s\n", prefix, TextStyle.Render(item.Title))
		if item.Description != "" {
			fmt.Printf("      %s\n", MutedStyle.Render(item.Description))
		}
	}

	fmt.Println()
	fmt.Print(AccentStyle.Render("Select option: "))

	var choice int
	_, err := fmt.Scanf("%d", &choice)
	if err != nil {
		return -1, err
	}

	if choice < 1 || choice > len(items) {
		return -1, fmt.Errorf("invalid selection")
	}

	return choice - 1, nil
}

// SelectAccount displays account selection menu
func SelectAccount(accounts []string, activeAccount string) (int, error) {
	items := make([]MenuItem, len(accounts))
	for i, acc := range accounts {
		title := acc
		if acc == activeAccount {
			title = SuccessStyle.Render("● ") + acc + SuccessStyle.Render(" (ACTIVE)")
		} else {
			title = MutedStyle.Render("○ ") + acc
		}
		items[i] = MenuItem{Title: title, Value: acc}
	}

	return SelectMenu("Choose account", items)
}

// SelectMethod displays method selection menu (SSH/Token)
func SelectMethod(hasSSH, hasToken bool) (string, error) {
	var items []MenuItem

	if hasSSH {
		items = append(items, MenuItem{Title: "SSH", Value: "ssh"})
	}
	if hasToken {
		items = append(items, MenuItem{Title: "Token (HTTPS)", Value: "token"})
	}

	if len(items) == 0 {
		return "", fmt.Errorf("no authentication methods available")
	}

	if len(items) == 1 {
		return items[0].Value, nil
	}

	idx, err := SelectMenu("Choose authentication method", items)
	if err != nil {
		return "", err
	}

	return items[idx].Value, nil
}

// SelectPlatform displays platform selection menu
func SelectPlatform() (string, error) {
	items := []MenuItem{
		{Title: "🐙 GitHub", Value: "github"},
		{Title: "🦊 GitLab", Value: "gitlab"},
		{Title: "🪣 Bitbucket", Value: "bitbucket"},
		{Title: "🍵 Gitea", Value: "gitea"},
		{Title: "🌐 Other", Value: "other"},
	}

	idx, err := SelectMenu("Choose platform", items)
	if err != nil {
		return "", err
	}

	return items[idx].Value, nil
}
