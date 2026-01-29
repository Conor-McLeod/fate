package main

import (
	"fmt"
	"math/rand"
	"os"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

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
)

type model struct {
	textInput    textinput.Model
	tasks        []string
	selectedTask string
}

func initialModel() model {
	ti := textinput.New()
	ti.Placeholder = "Enter a task..."
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = 30

	return model{
		textInput: ti,
		tasks:     []string{},
	}
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		case tea.KeyEnter:
			if m.textInput.Value() != "" {
				m.tasks = append(m.tasks, m.textInput.Value())
				m.textInput.SetValue("")
			}
		}

		switch msg.String() {
		case "r":
			if len(m.tasks) > 0 {
				m.selectedTask = m.tasks[rand.Intn(len(m.tasks))]
			}
		}
	}

	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

func (m model) View() string {
	s := titleStyle.Render("Random Task Picker") + "\n\n"
	s += m.textInput.View() + "\n"

	s += listStyle.Render("Tasks:") + "\n"
	for _, task := range m.tasks {
		s += fmt.Sprintf("• %s\n", task)
	}

	if m.selectedTask != "" {
		s += "\n" + winnerStyle.Render(fmt.Sprintf("DO THIS: %s", m.selectedTask)) + "\n"
	}

	s += "\n(Enter: add • r: pick • Esc: quit)\n"
	return s
}

func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}