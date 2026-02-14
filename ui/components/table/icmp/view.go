package icmpconfig

import (
	"github.com/charmbracelet/lipgloss"
)

// ═══════════════════════════════════════════════════════════
// View
// ═══════════════════════════════════════════════════════════

func (m Model) View() string {
	contentWidth := m.layout.BodyContentWidth()

	title := lipgloss.NewStyle().
		Width(contentWidth).
		Align(lipgloss.Center).
		Foreground(lipgloss.Color("230")).
		Bold(true).
		Padding(1, 0).
		Render("ICMP Configuration")

	tableView := lipgloss.NewStyle().
		Width(contentWidth).
		Align(lipgloss.Center).
		Render(
			lipgloss.NewStyle().
				BorderStyle(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("240")).
				Padding(1, 2).
				Render(m.table.View()),
		)

	helpView := lipgloss.NewStyle().
		Width(contentWidth).
		Align(lipgloss.Center).
		Foreground(lipgloss.Color("241")).
		MarginTop(1).
		Padding(0, 2).
		Render(m.help.View(m.keys))

	return lipgloss.NewStyle().
		Width(contentWidth).
		Render(lipgloss.JoinVertical(
			lipgloss.Center,
			title,
			tableView,
			helpView,
		))
}
