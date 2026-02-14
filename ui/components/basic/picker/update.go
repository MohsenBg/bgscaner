package picker

import (
	"bgscan/ui/shared/overlay"

	tea "github.com/charmbracelet/bubbletea"
)

// Update handles messages for the file picker overlay.
func (m Model) Update(msg tea.Msg) (overlay.Overlay, tea.Cmd) {
	var cmds []tea.Cmd

	// Handle resize
	if _, ok := msg.(tea.WindowSizeMsg); ok {
		m.FilePicker.SetHeight(pickerHeight(m.Layout))
	}

	// Normal update
	model, cmd := m.FilePicker.Update(msg)
	m.FilePicker = model
	if cmd != nil {
		cmds = append(cmds, cmd)
	}

	// Selection
	if didSelect, path := m.FilePicker.DidSelectFile(msg); didSelect && m.OnSelect != nil {
		cmds = append(cmds, m.OnSelect(path))
		cmds = append(cmds, m.CloseCmd())
	}

	return m, tea.Batch(cmds...)
}
