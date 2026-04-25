package tabs

import (
	"bgscan/internal/ui/shared/ui"

	tea "github.com/charmbracelet/bubbletea"
)

func (m *Model) Update(msg tea.Msg) (ui.Component, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == tea.KeyTab.String() {
			m.NextTab()
			cmd = m.selectTabCmd()
		}
		if msg.String() == tea.KeyShiftTab.String() {
			m.BackTab()
			cmd = m.selectTabCmd()
		}
	}
	return m, cmd
}
