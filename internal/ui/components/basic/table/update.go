package table

import (
	"bgscan/internal/ui/shared/ui"

	tea "github.com/charmbracelet/bubbletea"
)

func (m *Model) Update(msg tea.Msg) (ui.Component, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		cmd := m.Keys.Check(msg)
		if cmd != nil {
			return m, cmd
		}
		if msg.String() == "?" {
			m.FullHelp = !m.FullHelp
		}

	case tea.WindowSizeMsg:
		m.updateTableSize()
		return m, nil

	}

	var cmd tea.Cmd
	m.BubbleTable, cmd = m.BubbleTable.Update(msg)
	return m, cmd
}
