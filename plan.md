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

- [ ] **Unit 4: Random Selection Logic**
    - Goal: Implement the core "randomizer" feature.
    - Requirements:
        - Define a key binding (e.g., `r` or a specific button) to trigger selection.
        - Implement logic to randomly select an item from the task list.
        - Update the `View` to prominently display the selected task (the "winner").

- [ ] **Unit 5: Styling & Polish**
    - Goal: Improve the visual experience.
    - Requirements:
        - Use `lipgloss` to style the "winner" text (e.g., bold, color, border).
        - Style the task list (e.g., bullet points).
        - Add a help/instruction line (e.g., "Type a task & press Enter. Press 'r' to pick. 'q' to quit.").
