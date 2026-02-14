package notice

import (
	"github.com/charmbracelet/lipgloss"
)

func (m Model) View() string {
	wrapped := lipgloss.NewStyle().
		Width(m.Width()).
		Render(m.Message)

	m.viewport.SetContent(wrapped)
	content := lipgloss.JoinVertical(lipgloss.Top, m.headerView(), m.viewport.View(), m.footerView())
	return containerStyle(m.Width()).Render(content)
}

func (m Model) headerView() string {
	p := levelPalette(m.NoticeType)
	return titleStyle(m.Width(), m.NoticeType).Render(p.Icon + m.Title)
}

func (m Model) footerView() string {
	p := levelPalette(m.NoticeType)
	return CenterStyle(m.Width()).Render(
		ButtonStyle().Render(p.FooterText),
	)
}
