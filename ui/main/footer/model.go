package footer

import (
	"bgscan/ui/shared/layout"
	"bgscan/ui/shared/nav"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type Model struct {
	layout     *layout.Layout
	appVersion string
	status     string
}

type tickMsg time.Time

func (m Model) Init() tea.Cmd {
	return tickCmd()
}

func tickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func New(l *layout.Layout) Model {
	return Model{
		layout:     l,
		appVersion: "1.1.0",
		status:     nav.TitleFor(nav.ViewMainMenu),
	}
}
