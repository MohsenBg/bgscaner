package notice

import (
	"bgscan/internal/ui/shared/layout"
	"bgscan/internal/ui/shared/ui"

	tea "github.com/charmbracelet/bubbletea"
)

// NewNoticeCmd creates a BubbleTea command that opens a Notice overlay.
//
// The notice is positioned using the overlay manager and centered
// horizontally while appearing near the top of the screen.
//
// Parameters:
//   - layout: application layout reference
//   - title: notice title
//   - message: notice body message
//   - level: severity level (INFO, ERROR, SUCCESS)
func NewNoticeCmd(
	layout *layout.Layout,
	title string,
	message string,
	level LEVEL,
) tea.Cmd {
	return func() tea.Msg {
		return ui.AddNewOverlay(
			New(layout, title, message, level),
			ui.Center, // horizontal alignment
			ui.Top,    // vertical alignment
			0,         // horizontal offset
			5,         // vertical offset
		)
	}
}

// NoticeUnderDevelopment returns a command that displays a standard
// "Under Development" informational notice.
//
// This helper is used for UI sections that are not yet implemented.
func NoticeUnderDevelopment(layout *layout.Layout) tea.Cmd {
	title := "Under Development"

	message := "This section is currently being built.\n" +
		"Thank you for your patience. Stay tuned for future updates."

	return NewNoticeCmd(layout, title, message, NOTICE_INFO)
}
