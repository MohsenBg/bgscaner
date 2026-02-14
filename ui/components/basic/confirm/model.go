package confirm

import (
	"bgscan/ui/shared/layout"
	"bgscan/ui/shared/overlay"

	tea "github.com/charmbracelet/bubbletea"
)

type Model struct {
	id            overlay.OverlayID
	confirm       bool
	Layout        *layout.Layout
	Message       string
	ConfirmAction tea.Cmd
}

func New(layout *layout.Layout, Message string, actionOnConfirm tea.Cmd, defaultYes bool) Model {
	return Model{
		Layout:        layout,
		Message:       Message,
		ConfirmAction: actionOnConfirm,
		confirm:       defaultYes,
	}

}

func (m Model) ID() overlay.OverlayID {
	return m.id
}

func (m Model) Init() tea.Cmd {
	return nil
}
