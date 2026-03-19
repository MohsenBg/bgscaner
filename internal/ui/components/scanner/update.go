package scanner

import (
	"bgscan/internal/logger"
	"bgscan/internal/ui/components/basic/confirm"
	logview "bgscan/internal/ui/components/basic/logview"
	"bgscan/internal/ui/components/basic/notice"
	"bgscan/internal/ui/components/basic/progress"
	"bgscan/internal/ui/components/menus/ipviewer"
	"bgscan/internal/ui/shared/ui"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

// Update handles incoming BubbleTea messages and updates the scanner state.
//
// Responsibilities:
//   - Handle periodic tick updates
//   - Process user keyboard input
//   - Forward messages to child components
//   - Manage scanner lifecycle events
//   - Synchronize UI with background scan progress
func (m *Model) Update(msg tea.Msg) (ui.Component, tea.Cmd) {

	var cmds []tea.Cmd

	switch msg := msg.(type) {

	// Periodic UI refresh
	case tickMsg:
		m.mergeBatch()
		m.updateTableRows()
		return m, m.handleTick()

	// Pause / resume scan
	case TogglePauseMsg:
		m.togglePause()

	// Keyboard input
	case tea.KeyMsg:

		switch msg.String() {

		// Exit scan
		case "q", "b":
			return m, confirm.ConfirmCmd(
				m.layout,
				"Do you want to exit the scan?",
				func() tea.Msg {
					if m.scanner != nil {
						m.scanner.Close()
					}
					return ui.ResetComponentStacksMsg{}
				},
				false,
			)

		// Open log viewer
		case "l":
			return m, m.openLogViewer()
		}
	}

	// Forward message to child components
	var tCmd, pCmd tea.Cmd

	m.ipViewer, tCmd = m.ipViewer.Update(msg)
	m.progress, pCmd = m.progress.Update(msg)

	cmds = append(cmds, tCmd, pCmd)

	return m, tea.Batch(cmds...)
}

// togglePause pauses or resumes the scanner depending on its current state.
func (m *Model) togglePause() {

	if m.scanner.IsPaused() {
		m.scanner.Resume()
		return
	}

	m.scanner.Pause()
}

// openLogViewer opens an overlay containing the core application logs.
func (m *Model) openLogViewer() tea.Cmd {

	return func() tea.Msg {

		l := logview.New(m.layout, logger.Core(), "core logs")

		l.SetContainerWidth(min(80, m.layout.Body.Width))
		l.SetShowBorder(false)

		m.logViewer = l

		return ui.AddNewOverlay(l, ui.Center, ui.Center, 0, 0)
	}
}

// updateTableRows synchronizes the IP viewer table with the latest scan results.
func (m *Model) updateTableRows() {

	if viewer, ok := m.ipViewer.(*ipviewer.Model); ok {
		viewer.SetRows(m.ips)
	}
}

// errorCmd creates a UI command that displays an error notice.
func (m *Model) errorCmd(title, message string) tea.Cmd {
	return notice.NewNoticeCmd(m.layout, title, message, notice.NOTICE_ERROR)
}

// infoCmd creates a UI command that displays an informational notice.
func (m *Model) infoCmd(title, message string) tea.Cmd {
	return notice.NewNoticeCmd(m.layout, title, message, notice.NOTICE_INFO)
}

// handleTick processes periodic scanner state updates.
//
// It updates the progress bar, handles scan completion,
// and reacts to scanner errors.
func (m *Model) handleTick() tea.Cmd {

	var cmds []tea.Cmd

	switch m.status {

	case StatusScanning:

		cmds = append(cmds, func() tea.Msg {
			return progress.UpdateProgressMsg{
				Progress: m.progressInfo.Percent / 100,
			}
		})

	case StatusEnded:

		cmds = append(cmds, func() tea.Msg {
			return progress.UpdateProgressMsg{Progress: 1}
		})

	case StatusError:

		if m.scanError != nil {

			cmds = append(cmds, m.errorCmd(
				"Error while scanning",
				fmt.Sprintf("%v", m.scanError),
			))
		}
	}

	// Schedule next tick
	cmds = append(cmds, m.tick())

	return tea.Batch(cmds...)
}
