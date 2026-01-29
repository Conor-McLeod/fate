package main

import (
	"encoding/binary"
	"encoding/json"
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

const (
	dbName     = "tasks.db"
	bucketName = "Tasks"
)

type Task struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	PickedAt    time.Time `json:"picked_at,omitempty"`
	CompletedAt time.Time `json:"completed_at,omitempty"`
}

func (t Task) Duration() time.Duration {
	if t.CompletedAt.IsZero() || t.PickedAt.IsZero() {
		return 0
	}
	return t.CompletedAt.Sub(t.PickedAt)
}

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
}

func initialModel(db *bolt.DB) model {
	ti := textinput.New()
	ti.Placeholder = "Enter a task..."
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = 30

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

	return model{
		db:           db,
		textInput:    ti,
		confirmInput: ci,
		tasks:        pending,
		history:      history,
		cursor:       0,
		showHistory:  false,
	}
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	isBlocked := m.selectedTask != nil

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			if m.editingTask != nil {
				// Cancel edit
				m.editingTask = nil
				m.textInput.SetValue("")
				m.textInput.Blur()
				return m, nil
			}
			return m, tea.Quit
		case tea.KeyTab:
			if m.textInput.Focused() {
				// If editing, Tab might cancel or just blur? Let's say blur cancels edit mode for now to be safe or just blurs.
				// If we blur while editing, we are kinda in limbo. 
				// Let's enforce that Tab cancels edit if active, or just behaves normally but keeps editing state?
				// Simplest: Tab blurs input. If editingTask is set, it stays set, so if they focus again they are still editing.
				m.textInput.Blur()
				if isBlocked {
					m.confirmInput.Focus()
				}
			} else if m.confirmInput.Focused() {
				m.confirmInput.Blur()
			} else {
				m.textInput.Focus()
			}
		case tea.KeyEnter:
			if m.textInput.Focused() {
				taskName := strings.TrimSpace(m.textInput.Value())
				if taskName == "" {
					return m, nil
				}

				if m.editingTask != nil {
					// Update existing
					m.editingTask.Name = taskName
					if err := updateTask(m.db, *m.editingTask); err != nil {
						// Handle error
					}
					// Update in slice (it's a pointer to the slice element, so actually it might already be updated in memory? 
					// NO, m.editingTask is a pointer to the slice element of the *previous* model state. 
					// But we are in the same update loop. 
					// Wait, Go slices... if we modified the struct via pointer, it modifies the underlying array if capacity holds.
					// But `m.tasks` is a value receiver in the function signature `(m model)`.
					// So `m.editingTask` points to the heap or the old array. 
					// We need to explicitly update the slice in the new model `m`.
					
					// Re-find and update to be safe and functional style-ish
					for i, t := range m.tasks {
						if t.ID == m.editingTask.ID {
							m.tasks[i] = *m.editingTask
							break
						}
					}
					
					m.editingTask = nil
					m.textInput.SetValue("")
					m.textInput.Blur() // Exit edit mode completely
				} else {
					// Create new
					task, err := addTask(m.db, taskName)
					if err != nil {
						// In a real app, handle error properly
					} else {
						m.tasks = append(m.tasks, task)
						m.textInput.SetValue("")
						// Move cursor to new item
						m.cursor = len(m.tasks) - 1
					}
				}
			} else if m.confirmInput.Focused() {
				if m.confirmInput.Value() == "done" {
					m.completeSelectedTask()
					m.confirmInput.SetValue("")
					m.confirmInput.Blur()
					m.textInput.Focus() // Return focus to main input
				}
			}
		case tea.KeyUp:
			if !m.showHistory {
				if m.cursor > 0 {
					m.cursor--
				}
			}
		case tea.KeyDown:
			if !m.showHistory {
				if m.cursor < len(m.tasks)-1 {
					m.cursor++
				}
			}
		case tea.KeyBackspace, tea.KeyDelete:
			if !m.textInput.Focused() && !m.confirmInput.Focused() && len(m.tasks) > 0 && !isBlocked && !m.showHistory {
				m.deleteSelected()
			}
		}

		// Only handle character commands if NO input is focused
		if !m.textInput.Focused() && !m.confirmInput.Focused() {
			switch msg.String() {
			case "r":
				if !isBlocked && len(m.tasks) > 0 && !m.showHistory {
					// Pick a random task
					randomIndex := rand.Intn(len(m.tasks))
					m.selectedTask = &m.tasks[randomIndex]
					
					// Set PickedAt
					m.selectedTask.PickedAt = time.Now()
					// Update DB
					_ = updateTask(m.db, *m.selectedTask)
					
					// Move cursor to winner for visibility
					m.cursor = randomIndex
				}
			case "d":
				if !isBlocked && len(m.tasks) > 0 && !m.showHistory {
					m.deleteSelected()
				}
			case "c":
				if !isBlocked && !m.showHistory {
					m.clearAll()
				}
			case "h":
				m.showHistory = !m.showHistory
			case "e":
				if !isBlocked && len(m.tasks) > 0 && !m.showHistory {
					// Start editing
					m.editingTask = &m.tasks[m.cursor]
					m.textInput.SetValue(m.editingTask.Name)
					m.textInput.Focus()
				}
			}
		}
	}

	m.textInput, cmd = m.textInput.Update(msg)
	var cmd2 tea.Cmd
	m.confirmInput, cmd2 = m.confirmInput.Update(msg)
	
	return m, tea.Batch(cmd, cmd2)
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
	if m.cursor < 0 || m.cursor >= len(m.tasks) {
		return
	}
	
	task := m.tasks[m.cursor]
	if err := deleteTask(m.db, task.ID); err != nil {
		return // Handle error
	}

	// Remove from slice
	m.tasks = append(m.tasks[:m.cursor], m.tasks[m.cursor+1:]...)

	// Adjust cursor
	if m.cursor >= len(m.tasks) && len(m.tasks) > 0 {
		m.cursor = len(m.tasks) - 1
	}
	
	// Reset selection if we deleted the winner (shouldn't happen if blocked, but safe to keep)
	if m.selectedTask != nil && m.selectedTask.ID == task.ID {
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
	s := titleStyle.Render("Random Task Picker") + "\n\n"
	
	if m.showHistory {
		s += listStyle.Render("Completed Tasks:") + "\n"
		if len(m.history) == 0 {
			s += dimStyle.Render("No history yet.") + "\n"
		}
		for _, task := range m.history {
			dur := task.Duration().Round(time.Minute)
			s += fmt.Sprintf("• %s %s\n", task.Name, dimStyle.Render(fmt.Sprintf("(%s)", dur)))
		}
		s += "\n" + dimStyle.Render("(h: back to tasks • Esc: quit)") + "\n"
		return s
	}

	if m.editingTask != nil {
		s += lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Render("EDITING MODE") + "\n"
	}
	s += m.textInput.View() + "\n"

	s += listStyle.Render("Tasks:") + "\n"
	for i, task := range m.tasks {
		cursor := " " // no cursor
		style := taskStyle
		if m.cursor == i {
			cursor = ">" // cursor!
			style = cursorStyle
		}
		
		// If editing this specific task, maybe mark it visually?
		if m.editingTask != nil && m.editingTask.ID == task.ID {
			style = style.Copy().Foreground(lipgloss.Color("205")).Bold(true)
		}
		
		s += fmt.Sprintf("%s %s\n", cursorStyle.Render(cursor), style.Render(task.Name))
	}

	if m.selectedTask != nil {
		s += "\n" + winnerStyle.Render(fmt.Sprintf("DO THIS: %s", m.selectedTask.Name)) + "\n"
		s += m.confirmInput.View() + "\n"
	}

	help := ""
	if m.textInput.Focused() {
		if m.editingTask != nil {
			help = "(Enter: save • Esc: cancel edit)"
		} else {
			help = "(Enter: add • Tab: commands • Esc: quit)"
		}
	} else if m.confirmInput.Focused() {
		help = "(Type 'done' & Enter: finish • Tab: commands • Esc: quit)"
	} else {
		if m.selectedTask != nil {
			// Blocked state
			help = fmt.Sprintf("(%s • %s • %s • %s • Tab: edit • Esc: quit)", 
				strikeStyle.Render("r: pick"), 
				strikeStyle.Render("d: delete"), 
				strikeStyle.Render("c: clear"),
				"h: history") // Allow history check while blocked? Sure.
		} else {
			help = "(r: pick • d: delete • c: clear • e: edit • h: history • Tab: input • Esc: quit)"
		}
	}
	s += "\n" + dimStyle.Render(help) + "\n"
	return s
}

// DB Helpers

func setupDB() (*bolt.DB, error) {
	db, err := bolt.Open(dbName, 0600, nil)
	if err != nil {
		return nil, err
	}
	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(bucketName))
		return err
	})
	return db, err
}

func addTask(db *bolt.DB, name string) (Task, error) {
	var task Task
	err := db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketName))
		id64, _ := b.NextSequence()
		id := int(id64)
		
		task = Task{
			ID:   id,
			Name: name,
		}
		
		buf, err := json.Marshal(task)
		if err != nil {
			return err
		}
		
		return b.Put(itob(id), buf)
	})
	return task, err
}

func updateTask(db *bolt.DB, task Task) error {
	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketName))
		
		buf, err := json.Marshal(task)
		if err != nil {
			return err
		}
		
		return b.Put(itob(task.ID), buf)
	})
}

func loadTasks(db *bolt.DB) ([]Task, error) {
	var tasks []Task
	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketName))
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			var t Task
			if err := json.Unmarshal(v, &t); err != nil {
				// Skip invalid entries or handle error
				continue 
			}
			tasks = append(tasks, t)
		}
		return nil
	})
	return tasks, err
}

func deleteTask(db *bolt.DB, id int) error {
	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketName))
		return b.Delete(itob(id))
	})
}

func clearTasks(db *bolt.DB) error {
	return db.Update(func(tx *bolt.Tx) error {
		if err := tx.DeleteBucket([]byte(bucketName)); err != nil {
			return err
		}
		_, err := tx.CreateBucket([]byte(bucketName))
		return err
	})
}

func itob(v int) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(v))
	return b
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