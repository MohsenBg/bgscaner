package table

import (
	"bgscan/ui/shared/nav"

	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		cmd := m.Keys.Check(msg)
		if cmd != nil {
			return m, cmd
		}
		if msg.String() == "?" {
			m.FullHelp = !m.FullHelp
		}

	case tea.WindowSizeMsg, nav.OpenViewMsg:
		m = m.updateTableSize()
		return m, nil

	}

	var cmd tea.Cmd
	m.BubbleTable, cmd = m.BubbleTable.Update(msg)
	return m, cmd
}
