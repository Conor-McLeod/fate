package main

import "github.com/charmbracelet/lipgloss"

var (
	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#7D56F4")).
			Padding(0, 1)

	winnerStyle = lipgloss.NewStyle().
			Bold(true).
			Padding(0, 1).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#EE6FF8")).
			Width(60)

	listStyle = lipgloss.NewStyle().
			Bold(true).
			Italic(true).
			MarginTop(1)

	cursorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#EE6FF8"))

	taskStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FAFAFA")).
			Width(60)

	dimStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#666666")).
			Width(60)

	strikeStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#444444")).
			Strikethrough(true)
)
