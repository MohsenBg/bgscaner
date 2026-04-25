package logview

import (
	"bgscan/internal/ui/shared/ui"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Update handles BubbleTea messages for the log viewer.
//
// Responsibilities:
//   - Resize the viewport when the terminal size changes
//   - Refresh log content when new logs arrive (LogUpdateTickMsg)
//   - Forward messages to the viewport for scrolling and navigation
func (m *Model) Update(msg tea.Msg) (ui.Component, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg.(type) {

	case tea.WindowSizeMsg:
		m.setSize()
		m.viewport.SetContent(m.renderContent())

	case LogUpdateTickMsg:
		m.mu.Lock()

		if m.needUpdate {
			m.viewport.SetContent(m.renderContent())
			m.viewport.GotoBottom()
			m.needUpdate = false
		}

		m.mu.Unlock()

		cmds = append(cmds, m.tick())
	}

	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)

	if cmd != nil {
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

// renderContent builds the viewport content string from buffered messages.
func (m *Model) renderContent() string {
	if len(m.messages) == 0 {
		return ""
	}

	width := max(0, m.viewport.Width-5)

	return lipgloss.NewStyle().
		Width(width).
		Render(strings.Join(m.messages, "\n"))
}
