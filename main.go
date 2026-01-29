package main

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type model struct {
	textInput textinput.Model
	tasks     []string
}

func initialModel() model {

ti := textinput.New()
ti.Placeholder = "Enter a task..."
ti.Focus()
ti.CharLimit = 156
ti.Width = 20

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
	}

	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

func (m model) View() string {
	s := fmt.Sprintf(
		"What do you need to do?\n\n%s\n\n",
		m.textInput.View(),
	)

	s += "Tasks:\n"
	for _, task := range m.tasks {
		s += fmt.Sprintf("- %s\n", task)
	}

	s += "\nPress Esc to quit.\n"
	return s
}

func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}