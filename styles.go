package main

import "github.com/charmbracelet/lipgloss"

// Palette defines the application's color scheme.
var (
	ColorTextMain = lipgloss.Color("#FAFAFA")
	ColorBrand    = lipgloss.Color("#7D56F4")
	ColorAccent   = lipgloss.Color("#EE6FF8")
	ColorDim      = lipgloss.Color("#666666")
	ColorStrike   = lipgloss.Color("#444444")
	ColorError    = lipgloss.Color("#FF0000")
	ColorSpecial  = lipgloss.Color("205") // Hot Pink
)

var (
	titleStyle = lipgloss.NewStyle().
			Foreground(ColorTextMain).
			Background(ColorBrand).
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