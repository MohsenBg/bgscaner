package footer

import (
	"bgscan/ui/shared/nav"

	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tickMsg:
		return m, tickCmd()

	case UpdateAppVersion:
		m.appVersion = msg.AppVersion
		return m, cmd

	case UpdateStatus:
		m.status = msg.Status
		return m, cmd
	case nav.OpenViewMsg:
		m.status = nav.TitleFor(msg.View)
		return m, cmd
	}
	return m, cmd
}
