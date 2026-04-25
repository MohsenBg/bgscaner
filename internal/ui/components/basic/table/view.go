package table

import "github.com/charmbracelet/lipgloss"

func (m *Model) View() string {
	width := m.Layout.Body.Width
	tableView := tableViewStyle(width).Render(m.BubbleTable.View())
	return lipgloss.NewStyle().
		Width(width).
		Render(lipgloss.JoinVertical(
			lipgloss.Center,
			m.renderTitle(),
			tableView,
			m.renderHelpView(),
		))
}

func (m *Model) renderHelpView() string {
	width := m.Layout.Body.Width
	helpView := ""
	if m.FullHelp {
		helpView = helpViewStyle(width).
			Render(
				m.Help.FullHelpView(m.Keys.FullHelp()),
			)
	} else {
		helpView = helpViewStyle(width).
			Render(
				m.Help.ShortHelpView(m.Keys.ShortHelp()),
			)
	}
	return helpView
}

func (m *Model) renderTitle() string {
	width := m.Layout.Body.Width
	title := ""
	if m.Title != "" {
		title = titleStyles(width).Render(m.Title)
	}
	return title
}
