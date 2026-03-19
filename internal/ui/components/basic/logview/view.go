package logview

import (
	"github.com/charmbracelet/lipgloss"
)

// View renders the log viewer component.
//
// Layout:
//
//	title
//	log viewport
//	help bar
//
// When no log messages are available, a loading placeholder is displayed.
func (m *Model) View() string {
	container := ContainerStyle(m.containerWidth)

	// Title
	title := container.Render(
		TitleStyle(m.viewport.Width).Render(m.title),
	)

	// Content
	content := m.renderContentView()
	content = container.Render(content)

	// Help bar
	help := container.Render(
		helpStyle(m.viewport.Width).Render(helpView()),
	)

	return lipgloss.JoinVertical(
		lipgloss.Top,
		title,
		content,
		help,
	)
}

// renderContentView renders the viewport or loading state.
func (m *Model) renderContentView() string {
	if len(m.messages) == 0 {
		return "Loading Content...!"
	}

	content := m.viewport.View()

	if m.showBorder {
		content = BorderStyle(m.viewport.Width).Render(content)
	}

	return content
}

// helpView renders the keyboard help bar.
func helpView() string {
	return lipgloss.JoinHorizontal(
		lipgloss.Top,

		helpKeyStyle().Render("↑ ↓"),
		" move  ",

		helpKeyStyle().Render("b/esc"),
		" close",
	)
}
