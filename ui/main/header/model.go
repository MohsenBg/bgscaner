package header

import (
	"bgscan/ui/shared/layout"
	tea "github.com/charmbracelet/bubbletea"
)

type Model struct {
	layout *layout.Layout
}

func New(l *layout.Layout) Model {
	return Model{
		layout: l,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}
