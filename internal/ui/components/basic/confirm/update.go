package confirm

import (
	"bgscan/internal/ui/shared/ui"

	tea "github.com/charmbracelet/bubbletea"
)

// Update handles incoming BubbleTea messages and updates the confirmation dialog state.
//
// Key bindings:
//
//	Left / l   → select "No"
//	Right / j  → select "Yes"
//	Enter      → confirm the current selection
//
// When Enter is pressed:
//   - The dialog always closes.
//   - If the current selection is "Yes", the configured confirmation
//     command is executed.
func (m *Model) Update(msg tea.Msg) (ui.Component, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.KeyMsg:
		switch {

		// Select "No"
		case msg.Type == tea.KeyLeft || msg.String() == "l":
			m.confirm = false

		// Select "Yes"
		case msg.Type == tea.KeyRight || msg.String() == "j":
			m.confirm = true

		// Confirm selection
		case msg.Type == tea.KeyEnter:
			cmds := []tea.Cmd{m.CloseCmd()}

			if m.confirm && m.confirmFunc != nil {
				cmds = append(cmds, m.confirmFunc())
			}

			return m, tea.Batch(cmds...)
		}
	}

	return m, nil
}
