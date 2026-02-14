package menu

import (
	"bgscan/ui/shared/nav"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Update handles incoming messages and updates the menu state.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg, nav.OpenViewMsg:
		m.updateMenuLayout()

	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			if item, ok := m.GetSelected(); ok {
				if item.action != nil {
					return m, item.action()
				}
				if m.onSelect != nil {
					return m, m.onSelect(item)
				}
			}
		}

		for i, l := range m.items {
			if l.shortcut == msg.String() {
				m.List.Select(i)
				if item, ok := m.GetSelected(); ok {
					if item.action != nil {
						return m, item.action()
					}
					if m.onSelect != nil {
						return m, m.onSelect(item)
					}
				}
			}
		}
	}
	// Update list
	m.List, cmd = m.List.Update(msg)
	return m, cmd
}

func (m *Model) updateMenuLayout() {
	width := min(m.Layout.BodyContentWidth(), 50)
	height := min(m.Layout.BodyContentHeight(), 20)
	m.List.Styles.TitleBar = m.List.Styles.TitleBar.
		Width(width).
		Align(lipgloss.Center, lipgloss.Center)
	m.List.SetWidth(width)
	m.List.SetHeight(height)
}
