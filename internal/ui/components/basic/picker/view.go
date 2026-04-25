package picker

import (
	"github.com/charmbracelet/lipgloss"
)

// View renders the file picker overlay.
//
// Layout structure:
//
//	Title
//	Current Directory
//	File List
//	Help Bar
//
// The entire block is wrapped in a styled container sized
// according to the current layout constraints.
func (m *Model) View() string {

	// Limit picker width so it doesn't become too wide
	width := min(70, m.Layout.Content.Width-10)
	height := pickerHeight(m.Layout)

	// Optional title
	title := ""
	if m.Title != "" {
		title = TitleStyle(width).Render(m.Title)
	}

	// Current directory indicator
	currentDir := currentDirStyle(width).Render(
		m.FilePicker.CurrentDirectory,
	)

	// Compose overlay content
	content := lipgloss.JoinVertical(
		lipgloss.Top,
		title,
		currentDir,
		m.FilePicker.View(),
		helpStyle(width).Render(helpView()),
	)

	// Wrap everything in the picker container
	return containerStyle(width, height).Render(content)
}

// helpView renders the picker keyboard help bar.
func helpView() string {
	return lipgloss.JoinHorizontal(
		lipgloss.Top,

		helpKeyStyle().Render("← →"),
		" dir  ",

		helpKeyStyle().Render("↑ ↓"),
		" move  ",

		helpKeyStyle().Render("enter"),
		" select  ",

		helpKeyStyle().Render("b/esc"),
		" close",
	)
}
