package main

import "github.com/charmbracelet/lipgloss"

// Palette defines the application's color scheme.
var (
	ColorTextMain    = lipgloss.Color("#FAFAFA")
	ColorTextInverse = lipgloss.Color("#1A1B26") // Dark text for light backgrounds
	ColorAccent      = lipgloss.Color("#7AA2F7") // Periwinkle Blue
	ColorDim         = lipgloss.Color("#6B7280") // Cool Slate
	ColorStrike      = lipgloss.Color("#374151") // Darker Slate
	ColorError       = lipgloss.Color("#EF4444") // Red
	ColorSpecial     = lipgloss.Color("#4361EE") // Vibrant Blue
)

var (
	titleStyle = lipgloss.NewStyle().
			Foreground(ColorTextInverse).
			Background(ColorAccent).
			Padding(0, 1)

	winnerStyle = lipgloss.NewStyle().
			Bold(true).
			Padding(0, 1).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorAccent).
			Width(60)

	listStyle = lipgloss.NewStyle().
			Bold(true).
			Italic(true).
			MarginTop(1)

	cursorStyle = lipgloss.NewStyle().
			Foreground(ColorAccent)

	taskStyle = lipgloss.NewStyle().
			Foreground(ColorTextMain).
			Width(60)

	dimStyle = lipgloss.NewStyle().
			Foreground(ColorDim).
			Width(60)

	strikeStyle = lipgloss.NewStyle().
			Foreground(ColorStrike).
			Strikethrough(true)
)