package picker

import "github.com/charmbracelet/lipgloss"

// View renders the file picker overlay

func (m Model) View() string {
	width := min(70, m.Layout.Content.Width)
	height := pickerHeight(m.Layout)

	title := ""
	if m.Title != "" {
		title = TitleStyle(width).Render(m.Title)
	}

	currentDir := currentDirStyle(width).Render(
		m.FilePicker.CurrentDirectory,
	)

	content := lipgloss.JoinVertical(
		lipgloss.Top,
		title,
		currentDir,
		m.FilePicker.View(),
		helpStyle(width).Render(helpView()),
	)

	return containerStyle(width, height).Render(content)
}

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
