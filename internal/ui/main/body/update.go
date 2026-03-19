package body

import (
	"bgscan/internal/ui/components/basic/confirm"
	"bgscan/internal/ui/main/footer"
	"bgscan/internal/ui/shared/env"
	"bgscan/internal/ui/shared/ui"

	tea "github.com/charmbracelet/bubbletea"
)

func (m *Model) Update(msg tea.Msg) (ui.Component, tea.Cmd) {
	// Guard: Ensure we have at least one component to avoid panic
	if len(m.components) == 0 {
		return m, nil
	}

	lastIdx := len(m.components) - 1

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if env.IsBackKey(msg, m.components[lastIdx].Mode()) && len(m.components) > 1 {
			return m.popComponent()
		}

		if env.IsQuitKey(msg, m.components[lastIdx].Mode()) {
			return m, confirm.ExitConfirmCmd(m.layout)
		}

	case ui.OpenComponentMsg:
		if msg.Component != nil {
			return m.pushComponent(msg.Component)
		}

	case ui.CloseComponentMsg:
		for i, c := range m.components {
			if c.ID() == msg.ID {
				cmd := c.OnClose()
				m.components[i] = nil // GC: Avoid memory leak
				m.components = append(m.components[:i], m.components[i+1:]...)
				return m, cmd
			}
		}

	case ui.ResetComponentStacksMsg:
		var cmds []tea.Cmd
		for i := 1; i < len(m.components); i++ {
			cmd := m.components[i].OnClose()
			m.components[i] = nil
			cmds = append(cmds, cmd)
		}
		m.components = m.components[:1]
		return m, tea.Batch(cmds...)
	}

	// Only update the active (top) component
	activeComp, cmd := m.components[lastIdx].Update(msg)
	m.components[lastIdx] = activeComp

	return m, cmd
}

// --- Helper methods to keep Update clean ---

func (m *Model) pushComponent(c ui.Component) (ui.Component, tea.Cmd) {
	m.components = append(m.components, c)

	return m, tea.Batch(
		c.Init(),
		m.updateStatusCmd(c.Name()),
	)
}

func (m *Model) popComponent() (ui.Component, tea.Cmd) {
	lastIdx := len(m.components) - 1
	c := m.components[lastIdx]

	closeCmd := c.OnClose()
	m.components[lastIdx] = nil // GC: Clear reference
	m.components = m.components[:lastIdx]

	// After popping, update status to the new top component
	newTop := m.components[len(m.components)-1]

	return m, tea.Batch(
		closeCmd,
		m.updateStatusCmd(newTop.Name()),
	)
}

func (m *Model) updateStatusCmd(name string) tea.Cmd {
	return func() tea.Msg {
		return footer.UpdateStatus{Status: name}
	}
}
