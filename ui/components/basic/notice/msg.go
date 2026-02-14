package notice

import (
	"bgscan/ui/shared/layout"
	"bgscan/ui/shared/overlay"

	tea "github.com/charmbracelet/bubbletea"
)

func NewNoticeCmd(
	layout *layout.Layout,
	title,
	message string,
	level LEVEL,
) tea.Cmd {
	return func() tea.Msg {
		return overlay.NewAddOverlay(
			New(layout, title, message, level),
			overlay.Center,
			overlay.Center,
			0,
			0,
		)
	}
}
