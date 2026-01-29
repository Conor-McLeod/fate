# Development Plan: Random Task Picker TUI

- [x] **Unit 1: Project Initialization**
    - Goal: Initialize the Go environment and dependencies.
    - Requirements:
        - Initialize a new Go module (`random-task`).
        - Install `github.com/charmbracelet/bubbletea`.
        - Install `github.com/charmbracelet/bubbles`.
        - Install `github.com/charmbracelet/lipgloss`.

- [x] **Unit 2: Application Skeleton**
    - Goal: Create the basic structure of a Bubble Tea program.
    - Requirements:
        - Create `main.go`.
        - Define the main `model` struct.
        - Implement the mandatory interface methods: `Init`, `Update`, `View`.
        - Ensure the app runs and can be exited with `ctrl+c` or `q`.

- [x] **Unit 3: Task Input & List Display**
    - Goal: Enable users to add tasks to a list.
    - Requirements:
        - Integrate `textinput.Model` into the main model.
        - Handle `Enter` key: Add the current input value to a `[]string` of tasks and clear the input.
        - Render the list of added tasks below the input field in the `View`.

- [x] **Unit 4: Random Selection Logic**
    - Goal: Implement the core "randomizer" feature.
    - Requirements:
        - Define a key binding (e.g., `r` or a specific button) to trigger selection.
        - Implement logic to randomly select an item from the task list.
        - Update the `View` to prominently display the selected task (the "winner").

- [x] **Unit 5: Styling & Polish**
    - Goal: Improve the visual experience.
    - Requirements:
        - Use `lipgloss` to style the "winner" text (e.g., bold, color, border).
        - Style the task list (e.g., bullet points).
        - Add a help/instruction line (e.g., "Type a task & press Enter. Press 'r' to pick. 'q' to quit.").

- [x] **Unit 6: Persistence & Management**
    - Goal: Save tasks to disk and allow management.
    - Requirements:
        - Add `go.etcd.io/bbolt` dependency.
        - Initialize a BoltDB database.
        - Persist tasks: Save on add, delete on remove. Load on startup.
        - Add "Delete Task" functionality (requires cursor selection).
        - Add "Clear List" functionality.

- [x] **Unit 7: Enhanced Data Model & Storage**
    - Goal: Support rich task data (timestamps, completion status) in persistence layer.
    - Requirements:
        - Update `Task` struct: Add `PickedAt`, `CompletedAt` (time.Time).
        - Change BoltDB storage value from `string` (name) to JSON (serialized struct).
        - Update `addTask` to store JSON.
        - Update `loadTasks` to unmarshal JSON.
        - *Note: This will likely require clearing the existing DB or handling migration.*

- [x] **Unit 8: Time Tracking & History View**
    - Goal: Record durations and display completed tasks.
    - Requirements:
        - Logic: Set `PickedAt` when `r` is pressed.
        - Logic: Set `CompletedAt` when 'done' is entered. Calculate duration.
        - Move completed tasks to a separate list/bucket in memory (or filter by status).
        - Add `h` key to toggle "History View".
        - Render completed tasks with their duration (e.g., "Task A (5m)").

- [x] **Unit 9: Task Editing**
    - Goal: Allow modifying pending tasks.
    - Requirements:
        - Add `e` key to Edit.
        - Logic: Populate input with selected task text. Track `editingID`.
        - Logic: On Enter, update existing record instead of creating new.
        - Constraint: Block editing if a task is currently Picked/Active.

- [x] **Unit 10: Input Validation**
    - Goal: Ensure data quality.
    - Requirements:
        - Use `strings.TrimSpace` on input value.
        - Prevent Add/Edit if result is empty string.

