Output fate_demo_2.gif

Set TypingSpeed 125ms
Set FontSize 22
Set Width 1300
Set Height 650
Set Shell zsh

Sleep 1.5s
Type "fate"
Enter
Sleep 1.5s

# Add some tasks
Type "a"
Sleep 500ms

Type "Tidy desk"
Enter
Sleep 1s

Type "Walk the dog"
Enter
Sleep 1s

Type "Book dentist appointment"
Enter
Escape
Sleep 1.5s

# Go back and edit a task
Type "k"
Sleep 0.5
Type "k"
Sleep 0.5
Type "e"
Sleep 1s
Type " and drawers"
Sleep 1.5s
Enter
Sleep 1.5s

# Pick a task to do at random
Type "r"
Sleep 2s

# Add another task while the current task is running
Tab
Type "A task you forgot to add"
Enter
Tab

# Mark the current task as completed
Sleep 2s
Type "done"
Sleep 1s
Enter
Sleep 1s

# View the history
Type "h"
Sleep 3s

# Re-add task
Type "a"
Sleep 2s

# Go back to main view again
Type "h"
Sleep 2s

# Clear tasks
Type "c"
Sleep 2s
Type "y"

# End, pause before loop
Sleep 5s
