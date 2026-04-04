package main

import "github.com/charmbracelet/lipgloss"

type Styles struct {
	focusedToken lipgloss.Style
}

func defaultStyles() Styles {
	return Styles{
		focusedToken: lipgloss.NewStyle().Bold(true).Underline(true).Foreground(lipgloss.Color("10")),
	}
}
