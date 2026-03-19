package notice

import (
	"bgscan/internal/ui/shared/ui"

	tea "github.com/charmbracelet/bubbletea"
)

// Update handles incoming BubbleTea messages and updates the notice state.
//
// Behavior:
//   - Enter closes the notice dialog.
//   - All messages are forwarded to the internal viewport to support
//     scrolling for long messages.
func (m *Model) Update(msg tea.Msg) (ui.Component, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.KeyMsg:
		if msg.Type == tea.KeyEnter {
			return m, m.CloseCmd()
		}
	}

	// Delegate message handling to the viewport
	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)

	return m, cmd
}
