package main

import "github.com/charmbracelet/lipgloss"

var (
	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#7D56F4")).
			Padding(0, 1)

	winnerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#EE6FF8")).
			Bold(true).
			Padding(1, 2).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#EE6FF8"))

	listStyle = lipgloss.NewStyle().
			MarginTop(1)

	cursorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#EE6FF8"))

	taskStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FAFAFA"))

	dimStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#666666"))

	strikeStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#444444")).
			Strikethrough(true)
)
