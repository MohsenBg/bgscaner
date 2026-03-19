package menu

import (
	"bgscan/internal/ui/shared/ui"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Update handles incoming Bubble Tea messages and updates the menu state.
//
// It processes window resize events to recalculate the menu layout and
// keyboard input to trigger menu actions.
//
// Supported keys:
//
//   - "enter": Executes the currently selected menu item.
//   - "q": Exits the menu without executing an action.
//   - item shortcut: Selects and executes the corresponding menu item.
//
// Any messages that are not handled directly are forwarded to the underlying
// list component to preserve built‑in navigation behavior.
func (m *Model) Update(msg tea.Msg) (ui.Component, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
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
		case "q", tea.KeyEsc.String():
			return m, cmd
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

	// Update the underlying list component.
	m.List, cmd = m.List.Update(msg)
	return m, cmd
}

// updateMenuLayout recalculates and applies layout constraints for the menu.
//
// The menu width and height are clamped to a maximum size to prevent it from
// expanding excessively on large terminals while still adapting to smaller
// screens. The title bar is centered horizontally and vertically within the
// calculated width.
func (m *Model) updateMenuLayout() {
	width := min(m.Layout.BodyContentWidth(), 50)
	height := min(m.Layout.BodyContentHeight(), 20)

	m.List.Styles.TitleBar = m.List.Styles.TitleBar.
		Width(width).
		Align(lipgloss.Center, lipgloss.Center)

	m.List.SetWidth(width)
	m.List.SetHeight(height)
}
