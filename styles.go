package main

import "github.com/charmbracelet/lipgloss"

type Styles struct {
	normalToken, focusedToken, tokenField, modal lipgloss.Style
}

func defaultStyles() Styles {
	mainColor := lipgloss.Color("104")
	textColor := lipgloss.Color("7")
	focusedTokenColor := lipgloss.Color("10")

	return Styles{
		normalToken: lipgloss.NewStyle().
			Foreground(textColor),
		focusedToken: lipgloss.NewStyle().
			Bold(true).
			Underline(true).
			//Italic(true).
			Foreground(focusedTokenColor),
		tokenField: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(mainColor),
		modal: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(mainColor).
			Padding(1, 2).
			Align(lipgloss.Center, lipgloss.Top),
	}
}
