package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	bolt "go.etcd.io/bbolt"
)

type State int

const (
	StateBrowsing     State = iota
	StateAdding             // Typing in the main text box
	StateEditing            // Editing an existing task
	StateFocusMode          // "Winner" view
	StateHistory            // Viewing completed tasks
	StateConfirmClear       // Asking "are you sure?"
)

type model struct {
	db           *bolt.DB
	state        State
	textInput    textinput.Model
	confirmInput textinput.Model
	tasks        []Task // Pending tasks
	history      []Task // Completed tasks
	selectedTask *Task  // Pointer to the actual task in the slice/DB context
	editingTask  *Task  // Pointer to task being edited
	cursor       int
	err          error
	readdedIdx   int // Index of recently re-added history item (-1 if none)
	readdedTicks int // Ticks remaining to show flash
}
type tickMsg time.Time

func tick() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func initialModel(db *bolt.DB) model {
	ti := textinput.New()
	ti.Placeholder = "Enter a task..."
	ti.CharLimit = 300
	ti.Width = 100

	ci := textinput.New()
	ci.Placeholder = "Type 'done' to finish..."
	ci.CharLimit = 20
	ci.Width = 30

	allTasks, err := loadTasks(db)
	if err != nil {
		log.Fatal(err)
	}

	var pending, history []Task
	for _, t := range allTasks {
		if t.CompletedAt.IsZero() {
			pending = append(pending, t)
		} else {
			history = append(history, t)
		}
	}

	m := model{
		db:           db,
		state:        StateBrowsing,
		textInput:    ti,
		confirmInput: ci,
		tasks:        pending,
		history:      history,
		cursor:       0,
		readdedIdx:   -1,
	}

	// Restore active task if exists
	for i := range m.tasks {
		if !m.tasks[i].PickedAt.IsZero() {
			m.selectedTask = &m.tasks[i]
			m.cursor = i
			m.state = StateFocusMode
			m.confirmInput.Focus()
			m.textInput.Blur()
			break
		}
	}

	return m
}

func (m model) Init() tea.Cmd {
	return tea.Batch(textinput.Blink, tick())
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if model, cmd, handled := m.handleKey(msg); handled {
			return model, cmd
		}
	case tickMsg:
		// Clear re-added flash after countdown
		if m.readdedTicks > 0 {
			m.readdedTicks--
			if m.readdedTicks == 0 {
				m.readdedIdx = -1
			}
		}
		return m, tick()
	}

	var cmd tea.Cmd
	m.textInput, cmd = m.textInput.Update(msg)
	var cmd2 tea.Cmd
	m.confirmInput, cmd2 = m.confirmInput.Update(msg)

	return m, tea.Batch(cmd, cmd2)
}

func (m model) handleKey(msg tea.KeyMsg) (model, tea.Cmd, bool) {
	// Global keys
	switch msg.Type {
	case tea.KeyCtrlC:
		return m, tea.Quit, true
	}

	switch m.state {
	case StateBrowsing:
		switch msg.String() {
		case "j", "down":
			if m.cursor < len(m.tasks)-1 {
				m.cursor++
			}
		case "k", "up":
			if m.cursor > 0 {
				m.cursor--
			}
		case "r":
			if len(m.tasks) > 0 {
				randomIndex := rand.Intn(len(m.tasks))
				m.selectedTask = &m.tasks[randomIndex]
				m.selectedTask.PickedAt = time.Now()
				if err := updateTask(m.db, *m.selectedTask); err != nil {
					m.err = err
					return m, nil, true
				}
				m.cursor = randomIndex
				m.state = StateFocusMode
				m.confirmInput.Focus()
			}
		case "d", "backspace", "delete":
			if len(m.tasks) > 0 {
				m.deleteSelected()
			}
		case "c":
			m.state = StateConfirmClear
		case "e":
			if len(m.tasks) > 0 {
				m.editingTask = &m.tasks[m.cursor]
				m.textInput.SetValue(m.editingTask.Name)
				m.textInput.Focus()
				m.state = StateEditing
			}
		case "h":
			m.state = StateHistory
			m.cursor = 0
		case "a":
			m.state = StateAdding
			m.textInput.Focus()
		case "tab":
			if m.selectedTask != nil {
				m.state = StateFocusMode
				m.confirmInput.Focus()
			}
		case "esc":
			return m, tea.Quit, true
		}

	case StateAdding:
		switch msg.Type {
		case tea.KeyTab:
			if m.selectedTask != nil {
				m.state = StateFocusMode
				m.confirmInput.Focus()
				m.textInput.Blur()
				return m, nil, true
			}
		case tea.KeyEsc:
			if m.selectedTask != nil {
				return m, tea.Quit, true
			}
			m.textInput.Blur()
			m.textInput.SetValue("")
			m.state = StateBrowsing
			return m, nil, true
		case tea.KeyEnter:
			taskName := strings.TrimSpace(m.textInput.Value())
			if taskName != "" {
				task, err := addTask(m.db, taskName)
				if err != nil {
					m.err = err
					return m, nil, true
				}
				m.tasks = append(m.tasks, task)
				m.cursor = len(m.tasks) - 1
			}
			m.textInput.SetValue("")
			// Stay in adding mode for rapid entry? Or go back?
			// Original behavior was stay focused.
			// Let's keep focus.
			return m, nil, true
		default:
			return m, nil, false
		}

	case StateEditing:
		switch msg.Type {
		case tea.KeyEsc:
			m.editingTask = nil
			m.textInput.Blur()
			m.textInput.SetValue("")
			m.state = StateBrowsing
			return m, nil, true
		case tea.KeyEnter:
			taskName := strings.TrimSpace(m.textInput.Value())
			if taskName != "" {
				m.editingTask.Name = taskName
				if err := updateTask(m.db, *m.editingTask); err != nil {
					m.err = err
					return m, nil, true
				}
				// Update slice
				for i, t := range m.tasks {
					if t.ID == m.editingTask.ID {
						m.tasks[i] = *m.editingTask
						break
					}
				}
			}
			m.editingTask = nil
			m.textInput.SetValue("")
			m.textInput.Blur()
			m.state = StateBrowsing
			return m, nil, true
		default:
			return m, nil, false
		}

	case StateFocusMode:
		// Logic: User must type "done" in the confirmInput
		switch msg.Type {
		case tea.KeyEsc:
			// Allow quitting app, but NOT leaving focus mode easily?
			return m, tea.Quit, true
		case tea.KeyTab:
			m.state = StateAdding
			m.textInput.Focus()
			m.confirmInput.Blur()
			return m, nil, true
		case tea.KeyEnter:
			if m.confirmInput.Value() == "done" {
				m.completeSelectedTask()
				m.confirmInput.SetValue("")
				m.confirmInput.Blur()
				m.state = StateBrowsing
			}
			return m, nil, true
		default:
			return m, nil, false
		}

	case StateHistory:
		switch msg.String() {
		case "j", "down":
			if m.cursor < len(m.history)-1 {
				m.cursor++
			}
		case "k", "up":
			if m.cursor > 0 {
				m.cursor--
			}
		case "a":
			if len(m.history) > 0 && m.cursor < len(m.history) {
				task, err := addTask(m.db, m.history[m.cursor].Name)
				if err != nil {
					m.err = err
					return m, nil, true
				}
				m.tasks = append(m.tasks, task)
				m.readdedIdx = m.cursor
				m.readdedTicks = 2
			}
		case "d", "backspace", "delete":
			if len(m.history) > 0 {
				m.deleteSelected()
			}
		case "h":
			m.state = StateBrowsing
			m.cursor = 0
		case "esc":
			return m, tea.Quit, true
		}

	case StateConfirmClear:
		switch msg.String() {
		case "y", "Y":
			m.clearAll()
			m.state = StateBrowsing
		case "n", "N", "esc":
			m.state = StateBrowsing
		}
	}

	return m, nil, true
}

func (m *model) completeSelectedTask() {
	if m.selectedTask == nil {
		return
	}

	// Update timestamp
	m.selectedTask.CompletedAt = time.Now()
	if err := updateTask(m.db, *m.selectedTask); err != nil {
		m.err = err
		return
	}

	// Move to history
	m.history = append(m.history, *m.selectedTask)

	// Remove from pending tasks
	// Need to find index because selectedTask is a pointer
	for i, t := range m.tasks {
		if t.ID == m.selectedTask.ID {
			m.tasks = append(m.tasks[:i], m.tasks[i+1:]...)
			break
		}
	}

	m.selectedTask = nil

	// Fix cursor
	if m.cursor >= len(m.tasks) && len(m.tasks) > 0 {
		m.cursor = len(m.tasks) - 1
	}
}

func (m *model) deleteSelected() {
	var targetList *[]Task
	if m.state == StateHistory {
		targetList = &m.history
	} else {
		targetList = &m.tasks
	}

	if m.cursor < 0 || m.cursor >= len(*targetList) {
		return
	}

	task := (*targetList)[m.cursor]
	if err := deleteTask(m.db, task.ID); err != nil {
		m.err = err
		return
	}

	// Remove from slice
	*targetList = append((*targetList)[:m.cursor], (*targetList)[m.cursor+1:]...)

	// Adjust cursor
	if m.cursor >= len(*targetList) && len(*targetList) > 0 {
		m.cursor = len(*targetList) - 1
	}
}

func (m *model) clearAll() {
	// Clear pending tasks only

	for _, t := range m.tasks {
		if err := deleteTask(m.db, t.ID); err != nil {
			m.err = err
			return
		}
	}

	m.tasks = []Task{}
	m.cursor = 0
	m.selectedTask = nil
}

func (m model) View() string {
	s := titleStyle.Render("fate: a random task picker") + "\n\n"

	switch m.state {
	case StateHistory:
		s += m.viewHistory()
	case StateFocusMode:
		s += m.viewPending()
		s += m.viewWinner()
	default:
		// Browsing, Adding, Editing, ConfirmClear all share the main list view
		s += m.viewPending()
		// If we have a winner but we are temporarily doing something else (like adding), show it!
		if m.selectedTask != nil {
			s += m.viewWinner()
		}
	}

	s += "\n" + dimStyle.Render(m.viewHelp()) + "\n"

	if m.err != nil {
		s += lipgloss.NewStyle().Foreground(ColorError).Render(fmt.Sprintf("\nError: %v", m.err))
	}

	return s
}

func (m model) viewHistory() string {
	s := listStyle.Render("Completed Tasks:") + "\n"
	if len(m.history) == 0 {
		s += dimStyle.Render("No history yet.") + "\n"
	}
	for i, task := range m.history {
		cursor := "•"
		style := taskStyle
		var cStr string
		if m.cursor == i {
			cursor = ">"
			style = taskStyle.Copy().Foreground(cursorStyle.GetForeground())
			cStr = cursorStyle.Render(cursor)
		} else {
			cStr = lipgloss.NewStyle().Foreground(ColorTextMain).Render(cursor)
		}

		dur := task.Duration().Round(time.Minute).String()
		dur = strings.TrimSuffix(dur, "0s")
		if dur == "" {
			dur = "0m"
		}

		tStr := style.UnsetWidth().Render(task.Name) + " " + cursorStyle.Render(fmt.Sprintf("(%s)", dur))
		if m.readdedIdx == i {
			tStr += " " + lipgloss.NewStyle().Foreground(ColorAccent).Render("re-added!")
		}
		s += lipgloss.JoinHorizontal(lipgloss.Top, cStr+" ", tStr) + "\n"
	}
	return s
}

func (m model) viewPending() string {
	s := ""
	if m.state == StateEditing {
		s += lipgloss.NewStyle().Foreground(ColorSpecial).Render("EDITING MODE") + "\n"
	}
	if m.state == StateConfirmClear {
		s += lipgloss.NewStyle().Foreground(ColorError).Bold(true).Render("Are you sure you want to clear? y/N") + "\n"
	}
	s += m.textInput.View() + "\n"

	s += listStyle.Render("Tasks:") + "\n"
	for i, task := range m.tasks {
		cursor := "•"
		style := taskStyle
		var cStr string
		if m.cursor == i {
			cursor = ">" // cursor!
			style = taskStyle.Copy().Foreground(cursorStyle.GetForeground())
			cStr = cursorStyle.Render(cursor)
		} else {
			cStr = lipgloss.NewStyle().Foreground(ColorTextMain).Render(cursor)
		}

		// If editing this specific task, maybe mark it visually?
		if m.state == StateEditing && m.editingTask != nil && m.editingTask.ID == task.ID {
			style = style.Copy().Foreground(ColorSpecial).Bold(true)
		}

		tStr := style.Render(task.Name)
		s += lipgloss.JoinHorizontal(lipgloss.Top, cStr+" ", tStr) + "\n"
	}
	return s
}

func (m model) viewWinner() string {
	elapsed := time.Since(m.selectedTask.PickedAt).Round(time.Second)

	label := lipgloss.NewStyle().Foreground(ColorAccent).Bold(true).Render("DO THIS:")

	// Create the content for the box
	boxWidth := 60
	taskName := lipgloss.NewStyle().Foreground(ColorTextMain).Render(m.selectedTask.Name)
	separator := lipgloss.NewStyle().Foreground(ColorAccent).Render(strings.Repeat("─", boxWidth-2))

	timerLabel := lipgloss.NewStyle().Foreground(ColorAccent).Render("Elapsed: ")
	timerValue := lipgloss.NewStyle().Foreground(ColorTextMain).Render(elapsed.String())

	boxContent := fmt.Sprintf("%s\n%s\n%s%s", taskName, separator, timerLabel, timerValue)

	s := "\n" + label + "\n"
	s += winnerStyle.Render(boxContent) + "\n"
	s += m.confirmInput.View() + "\n"
	return s
}

func (m model) viewHelp() string {
	switch m.state {
	case StateConfirmClear:
		return "(y: yes, clear all • n: cancel)"
	case StateEditing:
		return "(Enter: save • Esc: cancel edit)"
	case StateAdding:
		if m.selectedTask != nil {
			return "(Enter: save • Tab: toggle • Esc: quit)"
		}
		return "(Enter: save • Esc: cancel)"
	case StateFocusMode:
		return "(Enter: submit • Tab: toggle • Esc: quit)"
	case StateHistory:
		return "(h: back • j/k: nav • a: re-add • d: delete • Esc: quit)"
	default:
		return "(j/k: nav • r: pick • d: delete • c: clear • e: edit • h: history • a: add • Esc: quit)"
	}
}

func main() {
	db, err := setupDB()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	p := tea.NewProgram(initialModel(db))
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
