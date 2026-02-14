package scan

import (
	"bgscan/ui/components/basic/menu"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg.(type) {
	case tea.WindowSizeMsg:
		m = m.updateMenuScale()
	}

	var cmd tea.Cmd
	updated, cmd := m.menu.Update(msg)
	menuModel, ok := updated.(menu.Model)
	if !ok {
		return m, nil
	}
	m.menu = menuModel
	return m, cmd
}

func (m Model) updateMenuScale() Model {
	width := min(m.layout.BodyContentWidth(), 50)
	hight := min(m.layout.BodyContentHeight(), 20)
	m.menu.List.Styles.TitleBar = m.menu.List.Styles.TitleBar.Width(width).Align(lipgloss.Center, lipgloss.Center)
	m.menu.List.SetWidth(width)
	m.menu.List.SetHeight(hight)
	return m
}
