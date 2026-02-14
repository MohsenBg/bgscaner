package confirm

import (
	"bgscan/ui/shared/overlay"

	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) Update(msg tea.Msg) (overlay.Overlay, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:

		// Left NO
		if msg.Type == tea.KeyLeft || msg.String() == "l" {
			m.confirm = false
		}

		// Right YES
		if msg.Type == tea.KeyRight || msg.String() == "j" {
			m.confirm = true
		}

		if msg.Type == tea.KeyEnter {
			if m.confirm {
				return m, m.ConfirmAction
			}
			return m, m.CloseCmd()
		}
	}

	return m, nil
}
