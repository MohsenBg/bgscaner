package body

import (
	"bgscan/ui/shared/nav"
	"bgscan/ui/shared/overlay"

	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {

	// ----------------------------------------
	// Add new overlay
	// ----------------------------------------
	case overlay.AddOverlayMsg:
		m.overlays = append(m.overlays, msg.Overlay)
		m.SetOverlayPlacement(msg.Overlay.ID(), OverlayPlacement{
			msg.XPos, msg.YPos, msg.XOffset, msg.YOffset,
		})
		return m, msg.Overlay.Init()

	// ----------------------------------------
	// Close overlay by ID
	// ----------------------------------------
	case overlay.CloseOverlayMsg:
		for i, ov := range m.overlays {
			if ov.ID() == msg.ID {
				m.overlays = append(m.overlays[:i], m.overlays[i+1:]...)
				break
			}
		}
	}

	// ----------------------------------------
	// Update overlays
	// ----------------------------------------

	key, isKey := msg.(tea.KeyMsg)
	// ----------------------------------------
	// Global back handling (esc / b)
	// ----------------------------------------
	if isKey && isBackKey(key) {
		if len(m.overlays) > 0 {
			m.overlays = m.overlays[:len(m.overlays)-1]
			return m, nil
		}
		return m.handleBack()
	}

	for i := range m.overlays {
		ov := m.overlays[i]

		// Only top overlay gets keyboard input
		if isKey && i != len(m.overlays)-1 {
			continue
		}

		var cmd tea.Cmd
		ov, cmd = ov.Update(msg)
		m.overlays[i] = ov
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}

	// ----------------------------------------
	// Main view updates (non-key messages)
	// ----------------------------------------
	if len(m.overlays) == 0 || !isKey {
		var cmd tea.Cmd
		m.rootView, cmd = m.rootView.Update(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}

	return m, tea.Batch(cmds...)
}

func isBackKey(msg tea.KeyMsg) bool {
	switch msg.Type {
	case tea.KeyEsc, tea.KeyBackspace:
		return true
	}
	return msg.String() == "b"
}

func (m Model) handleBack() (Model, tea.Cmd) {
	m.PopView()
	return m, func() tea.Msg { return nav.OpenViewMsg{View: m.CurrentView()} }
}

