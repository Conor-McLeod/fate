# fate

**fate** is a tiny terminal-based task manager designed to eliminate decision fatigue. Write some todos, roll the dice, and let *fate* decide what you work on next.

Built with [Go](https://go.dev/) and [Bubble Tea](https://github.com/charmbracelet/bubbletea).

## Features

-   **🎯 Random Picker:** Let *fate* decide what you work on next.
-   **📝 Task Management:** Add, edit, and delete tasks quickly, without taking your hands off the keyboard.
-   **⚡️ Focus Mode:** Once a task is picked, the UI focuses purely on that task until it's done. Go and do that thing!
-   **📜 History:** Automatically tracks completed tasks and their duration.
-   **💾 Auto-Save:** Tasks are persisted locally using BoltDB.
-   **🎨 Beautiful TUI:** A clean, delightful interface styled with Lipgloss.

## Installation

### From Source

Ensure you have [Go 1.25+](https://go.dev/dl/) installed.

```bash
# Clone the repository
git clone https://github.com/conormcleod/fate.git
cd fate

# Install the binary
go install .
```

Ensure your `$(go env GOPATH)/bin` is in your system `$PATH`.

## Usage

Run the tool from your terminal:

```bash
fate
```

### Key Controls

| Key | Action |
| :--- | :--- |
| **Normal Mode** | |
| `Enter` | Add a new task (type in box first) |
| `r` | **Roll the die** (Pick a random task) |
| `j` / `k` | Navigate down / up |
| `d` | Delete selected task |
| `e` | Edit selected task |
| `h` | Toggle History view |
| `c` | Clear all pending tasks |
| `Esc` | Quit |
| **Focus Mode** (After picking a task) | |
| `Tab` | Focus the input box |
| `Type 'done'` | Mark the task as complete |
| `Esc` | Quit (Preserves state) |

## Data Storage

**fate** stores your tasks in a local BoltDB database located at:
`~/.local/share/fate/fate.db`

This allows you to access your centralized task list from any directory on your machine.

## Built With

-   [Bubble Tea](https://github.com/charmbracelet/bubbletea) - The TUI framework.
-   [Lipgloss](https://github.com/charmbracelet/lipgloss) - Style definitions.
-   [BoltDB](https://github.com/etcd-io/bbolt) - Embedded key/value database.

## License

[MIT](LICENSE)
