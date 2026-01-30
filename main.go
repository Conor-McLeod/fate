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





type model struct {
	db           *bolt.DB
	textInput    textinput.Model
	confirmInput textinput.Model
	tasks        []Task // Pending tasks
	history      []Task // Completed tasks
	selectedTask *Task  // Pointer to the actual task in the slice/DB context
	editingTask  *Task  // Pointer to task being edited
	cursor       int
	showHistory  bool
	confirmClear bool
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
	ti.Focus()
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
		textInput:    ti,
		confirmInput: ci,
		tasks:        pending,
		history:      history,
		cursor:       0,
		showHistory:  false,
	}

	// Restore active task if exists
	for i := range m.tasks {
		if !m.tasks[i].PickedAt.IsZero() {
			m.selectedTask = &m.tasks[i]
			m.cursor = i
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
		return m, tick()
	}

	var cmd tea.Cmd
	m.textInput, cmd = m.textInput.Update(msg)
	var cmd2 tea.Cmd
	m.confirmInput, cmd2 = m.confirmInput.Update(msg)

	return m, tea.Batch(cmd, cmd2)
}

func (m model) handleKey(msg tea.KeyMsg) (model, tea.Cmd, bool) {
	isBlocked := m.selectedTask != nil

	switch msg.Type {
	case tea.KeyCtrlC:
		return m, tea.Quit, true
	case tea.KeyEsc:
		if m.confirmClear {
			m.confirmClear = false
			return m, nil, true
		}
		if m.textInput.Focused() || m.confirmInput.Focused() {
			m.textInput.Blur()
			m.confirmInput.Blur()
			return m, nil, true
		}
		return m, tea.Quit, true
	case tea.KeyTab:
		if isBlocked {
			if m.textInput.Focused() {
				m.textInput.Blur()
				m.confirmInput.Focus()
			} else {
				if m.confirmInput.Focused() {
					m.confirmInput.Blur()
					m.textInput.Focus()
				} else {
					m.confirmInput.Focus()
				}
			}
			return m, nil, true
		} else {
			if m.textInput.Focused() {
				m.textInput.Blur()
			} else {
				m.textInput.Focus()
			}
			return m, nil, true
		}
	case tea.KeyEnter:
		if m.textInput.Focused() {
			taskName := strings.TrimSpace(m.textInput.Value())
			if taskName == "" {
				return m, nil, true
			}

			if m.editingTask != nil {
				m.editingTask.Name = taskName
				_ = updateTask(m.db, *m.editingTask)

				// Update the task in the tasks slice
				for i, t := range m.tasks {
					if t.ID == m.editingTask.ID {
						m.tasks[i] = *m.editingTask
						break
					}
				}

				m.editingTask = nil
				m.textInput.SetValue("")
				m.textInput.Blur()
			} else {
				task, err := addTask(m.db, taskName)
				if err == nil {
					m.tasks = append(m.tasks, task)
					m.textInput.SetValue("")
					m.cursor = len(m.tasks) - 1
				}
			}
			return m, nil, true
		} else if m.confirmInput.Focused() {
			if m.confirmInput.Value() == "done" {
				m.completeSelectedTask()
				m.confirmInput.SetValue("")
				m.confirmInput.Blur()
				m.textInput.Focus()
			}
			return m, nil, true
		}
	case tea.KeyUp:
		if !m.textInput.Focused() && !m.confirmInput.Focused() {
			if m.cursor > 0 {
				m.cursor--
			}
			return m, nil, true
		}
	case tea.KeyDown:
		if !m.textInput.Focused() && !m.confirmInput.Focused() {
			maxLen := len(m.tasks)
			if m.showHistory {
				maxLen = len(m.history)
			}
			if m.cursor < maxLen-1 {
				m.cursor++
			}
			return m, nil, true
		}
	case tea.KeyBackspace, tea.KeyDelete:
		canDelete := !m.textInput.Focused() && !m.confirmInput.Focused() && !isBlocked
		hasItems := (m.showHistory && len(m.history) > 0) || (!m.showHistory && len(m.tasks) > 0)

		if canDelete && hasItems {
			m.deleteSelected()
			return m, nil, true
		}
	}

	// Only handle character commands if NO input is focused
	if !m.textInput.Focused() && !m.confirmInput.Focused() {
		if m.confirmClear {
			if msg.String() == "y" || msg.String() == "Y" {
				m.clearAll()
			}
			m.confirmClear = false
			return m, nil, true
		}

		switch msg.String() {
		case "k":
			if m.cursor > 0 {
				m.cursor--
			}
			return m, nil, true
		case "j":
			maxLen := len(m.tasks)
			if m.showHistory {
				maxLen = len(m.history)
			}
			if m.cursor < maxLen-1 {
				m.cursor++
			}
			return m, nil, true
		case "r":
			if !isBlocked && len(m.tasks) > 0 && !m.showHistory {
				randomIndex := rand.Intn(len(m.tasks))
				m.selectedTask = &m.tasks[randomIndex]
				m.selectedTask.PickedAt = time.Now()
				_ = updateTask(m.db, *m.selectedTask)
				m.cursor = randomIndex
				m.confirmInput.Focus()
				return m, nil, true
			}
		case "d":
			if !isBlocked && ((!m.showHistory && len(m.tasks) > 0) || (m.showHistory && len(m.history) > 0)) {
				m.deleteSelected()
				return m, nil, true
			}
		case "c":
			if !isBlocked && !m.showHistory {
				m.confirmClear = true
				return m, nil, true
			}
		case "h":
			m.showHistory = !m.showHistory
			m.cursor = 0
			return m, nil, true
		case "e":
			if !isBlocked && len(m.tasks) > 0 && !m.showHistory {
				m.editingTask = &m.tasks[m.cursor]
				m.textInput.SetValue(m.editingTask.Name)
				m.textInput.Focus()
				return m, nil, true
			}
		}
	}

	return m, nil, false
}

func (m *model) completeSelectedTask() {
	if m.selectedTask == nil {
		return
	}
	
	// Update timestamp
	m.selectedTask.CompletedAt = time.Now()
	_ = updateTask(m.db, *m.selectedTask)
	
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
	if m.showHistory {
		targetList = &m.history
	} else {
		targetList = &m.tasks
	}

	if m.cursor < 0 || m.cursor >= len(*targetList) {
		return
	}
	
	task := (*targetList)[m.cursor]
	if err := deleteTask(m.db, task.ID); err != nil {
		return // Handle error
	}

	// Remove from slice
	*targetList = append((*targetList)[:m.cursor], (*targetList)[m.cursor+1:]...)

	// Adjust cursor
	if m.cursor >= len(*targetList) && len(*targetList) > 0 {
		m.cursor = len(*targetList) - 1
	}
	
	// Reset selection if we deleted the winner (only relevant for pending tasks)
	if !m.showHistory && m.selectedTask != nil && m.selectedTask.ID == task.ID {
		m.selectedTask = nil
	}
}

func (m *model) clearAll() {
	// Only clear pending tasks? Or all? User likely expects current list cleared.
	// Since we are in pending view (h blocked), let's clear pending.
	
	// NOTE: Implementation of clearTasks wipes the bucket, which kills history too.
	// We should probably iterate and delete only pending.
	// For now, let's keep simple clearTasks but re-save history? 
	// Or better, iterate pending tasks and delete them one by one.
	
	for _, t := range m.tasks {
		_ = deleteTask(m.db, t.ID)
	}
	
	m.tasks = []Task{}
	m.cursor = 0
	m.selectedTask = nil
}

func (m model) View() string {
	s := titleStyle.Render("fate: a random task picker") + "\n\n"
	
	if m.showHistory {
		s += listStyle.Render("Completed Tasks:") + "\n"
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
				cStr = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF")).Render(cursor)
			}
			
			dur := task.Duration().Round(time.Minute).String()
			dur = strings.TrimSuffix(dur, "0s")
			if dur == "" {
				dur = "0m"
			}
			
			tStr := style.UnsetWidth().Render(task.Name) + " " + cursorStyle.Render(fmt.Sprintf("(%s)", dur))
			s += lipgloss.JoinHorizontal(lipgloss.Top, cStr+" ", tStr) + "\n"
		}
		s += "\n" + dimStyle.Render("(h: back • j/k: nav • d: delete • Esc: quit)") + "\n"
		return s
	}

	if m.editingTask != nil {
		s += lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Render("EDITING MODE") + "\n"
	}
	if m.confirmClear {
		s += lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0000")).Bold(true).Render("Are you sure you want to clear? y/N") + "\n"
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
			cStr = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF")).Render(cursor)
		}
		
		// If editing this specific task, maybe mark it visually?
		if m.editingTask != nil && m.editingTask.ID == task.ID {
			style = style.Copy().Foreground(lipgloss.Color("205")).Bold(true)
		}
		
		tStr := style.Render(task.Name)
		s += lipgloss.JoinHorizontal(lipgloss.Top, cStr+" ", tStr) + "\n"
	}

	if m.selectedTask != nil {
		elapsed := time.Since(m.selectedTask.PickedAt).Round(time.Second)
		
		accentColor := lipgloss.Color("#EE6FF8")
		whiteColor := lipgloss.Color("#FFFFFF")
		
		label := lipgloss.NewStyle().Foreground(accentColor).Bold(true).Render("DO THIS:")
		
		// Create the content for the box
		boxWidth := 60
		taskName := lipgloss.NewStyle().Foreground(whiteColor).Render(m.selectedTask.Name)
		separator := lipgloss.NewStyle().Foreground(accentColor).Render(strings.Repeat("─", boxWidth-2))
		
		timerLabel := lipgloss.NewStyle().Foreground(accentColor).Render("Elapsed: ")
		timerValue := lipgloss.NewStyle().Foreground(accentColor).Render(elapsed.String())
		
		boxContent := fmt.Sprintf("%s\n%s\n%s%s", taskName, separator, timerLabel, timerValue)
		
		s += "\n" + label + "\n"
		s += winnerStyle.Render(boxContent) + "\n"
		s += m.confirmInput.View() + "\n"
	}

	help := ""
	if m.confirmClear {
		help = "(y: yes, clear all • n: cancel)"
	} else if m.textInput.Focused() {
		if m.editingTask != nil {
			help = "(Enter: save • Esc: cancel edit)"
		} else {
			if m.selectedTask != nil {
				help = "(Enter: add • Tab: confirm • Esc: blur)"
			} else {
				help = "(Enter: add • Tab: commands • Esc: blur)"
			}
		}
	} else if m.confirmInput.Focused() {
		help = "(Type 'done' & Enter: finish • Tab: input • Esc: blur)"
	} else {
		if m.selectedTask != nil {
			// Blocked state
			help = fmt.Sprintf("(%s • %s • %s • %s • Tab: done • Esc: quit)", 
				strikeStyle.Render("r: pick"), 
				strikeStyle.Render("d: delete"), 
				strikeStyle.Render("c: clear"),
				"h: history") // Allow history check while blocked? Sure.
		} else {
			help = "(j/k: nav • r: pick • d: delete • c: clear • e: edit • h: history • Tab: input • Esc: quit)"
		}
	}
	s += "\n" + dimStyle.Render(help) + "\n"
	return s
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