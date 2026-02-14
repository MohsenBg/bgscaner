package confirm

import (
	"bgscan/ui/shared/layout"
	"bgscan/ui/shared/overlay"

	tea "github.com/charmbracelet/bubbletea"
)

func ExitConfirmCmd(layout *layout.Layout) tea.Cmd {
	return func() tea.Msg {
		return overlay.NewAddOverlay(
			New(layout, "Are you sure you want to exit?", tea.Quit, false),
			overlay.Center,
			overlay.Top,
			0,
			0,
		)
	}
}

func ConfirmCmd(
	layout *layout.Layout,
	message string,
	confirm tea.Cmd,
	defaultYes bool,
) tea.Cmd {
	return func() tea.Msg {
		return overlay.NewAddOverlay(
			New(layout, message, confirm, defaultYes),
			overlay.Center,
			overlay.Top,
			0,
			0,
		)
	}
}

func (m Model) CloseCmd() tea.Cmd {
	return func() tea.Msg {
		return overlay.CloseOverlayMsg{ID: m.ID()}
	}
}
