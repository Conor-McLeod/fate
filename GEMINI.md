# Random Task Picker

## Project Overview

`random-task` is a command-line interface (CLI) tool designed to help users manage tasks and overcome decision fatigue by randomly selecting a task to focus on. Built with Go and the [Bubble Tea](https://github.com/charmbracelet/bubbletea) framework, it features an interactive Terminal User Interface (TUI).

Key features include:
*   **Task Management:** Add, edit, and delete tasks.
*   **Random Picker:** Randomly select a "winner" task to work on.
*   **Focus Mode:** When a task is picked, the UI focuses on it until completion.
*   **History:** Track completed tasks and their duration.
*   **Persistence:** Tasks are saved locally using [BoltDB](https://github.com/etcd-io/bbolt).

## Architecture & Technologies

The project is structured as a standard Go module:

*   **Language:** Go (1.25+)
*   **TUI Framework:** [Bubble Tea](https://github.com/charmbracelet/bubbletea) (Model-View-Update pattern)
*   **Styling:** [Lipgloss](https://github.com/charmbracelet/lipgloss)
*   **Database:** [BoltDB](https://github.com/etcd-io/bbolt) (Embedded key/value store)

### Key Files

*   `main.go`: Contains the main entry point and the Bubble Tea application logic (Model definition, Update loop, and View rendering).
*   `db.go`: Handles database setup, the `Task` struct definition, and all CRUD operations for BoltDB.
*   `styles.go`: Defines the visual styles using Lipgloss.
*   `tasks.db`: The local BoltDB database file (created automatically).

## Building and Running

### Prerequisites
*   Go 1.25 or higher

### Commands

**Run directly:**
```bash
go run .
```

**Build binary:**
```bash
go build -o random-task .
```

**Run binary:**
```bash
./random-task
```

## Usage Guide

The application is keyboard-driven:

*   **Add Task:** Type in the input box and press `Enter`.
*   **Navigation:** `j`/`k` or `Up`/`Down` arrows.
*   **Pick Task:** Press `r` to randomly select a task.
*   **Complete Task:** When a task is picked, type `done` and press `Enter` to finish it.
*   **Edit Task:** Press `e` on a selected task.
*   **Delete Task:** Press `d` or `Backspace`.
*   **History:** Press `h` to toggle completed tasks view.
*   **Clear All:** Press `c` to clear all tasks (requires confirmation).
*   **Quit:** `Esc` or `Ctrl+C`.

## Development Conventions

*   **Code Style:** Follows standard Go formatting (`gofmt`).
*   **Architecture:** The app follows the Elm Architecture (Model, View, Update) mandated by Bubble Tea.
    *   **Model:** Holds the state (database connection, inputs, lists of tasks).
    *   **Update:** Handles messages (key presses, commands) and updates the model.
    *   **View:** Renders the TUI as a string based on the current model.
*   **Database:** Operations are performed inside BoltDB transactions (`View` for read, `Update` for write).
