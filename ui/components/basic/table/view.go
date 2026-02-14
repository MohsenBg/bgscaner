package table

import "github.com/charmbracelet/lipgloss"

func (m Model) View() string {
	contentWidth := m.Layout.Body.Width

	title := titleStyles(contentWidth).Render(m.Title)
	tableView := tableViewStyle(contentWidth).Render(m.BubbleTable.View())

	helpView := ""
	if m.FullHelp {
		helpView = helpViewStyle(contentWidth).
			Render(
				m.Help.FullHelpView(m.Keys.FullHelp()),
			)
	} else {
		helpView = helpViewStyle(contentWidth).
			Render(
				m.Help.ShortHelpView(m.Keys.ShortHelp()),
			)
	}

	return lipgloss.NewStyle().
		Width(contentWidth).
		Render(lipgloss.JoinVertical(
			lipgloss.Center,
			title,
			tableView,
			helpView,
		))
}
