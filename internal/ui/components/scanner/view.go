package scanner

import (
	"bgscan/internal/core/scanner/engine"
	"fmt"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// View renders the scanner UI.
//
// Layout:
//
//	Tabs
//	Progress Panel
//	IP Results Table
func (m *Model) View() string {
	idx := m.currentTab

	return lipgloss.JoinVertical(
		lipgloss.Left,
		m.tabs.View(),
		m.renderProgress(idx),
		m.ipViewers[idx].View(),
	)
}

// renderProgress renders the progress panel shown above the IP results.
//
// Panel sections:
//
//  1. Scan statistics
//  2. Current scanner status / ETA
//  3. Progress bar
func (m *Model) renderProgress(idx int) string {
	p := m.progressInfo[idx]
	width := m.layout.Body.Width

	stats := m.renderStatsRow(p)
	status := m.renderStatusRow()

	container := lipgloss.NewStyle().
		Width(width).
		Align(lipgloss.Center)

	return lipgloss.JoinVertical(
		lipgloss.Center,
		container.Render(stats),
		container.Padding(1, 0).Render(status),
		container.Render(m.progress[idx].View()),
	)
}

// renderStatsRow builds the statistics row.
func (m *Model) renderStatsRow(p engine.Progress) string {

	left := p.Total - p.Processed
	if left < 0 {
		left = 0
	}

	return lipgloss.JoinHorizontal(
		lipgloss.Left,
		scannedStyle().Render(fmt.Sprintf("scanned: %d", p.Processed)),
		separatorStyle().Render(" | "),
		leftStyle().Render(fmt.Sprintf("left: %d", left)),
		separatorStyle().Render(" | "),
		foundStyle().Render(fmt.Sprintf("found: %d", p.Succeed)),
		separatorStyle().Render(" | "),
		elapsedStyle().Render(fmt.Sprintf(
			"elapsed: %s",
			p.Elapsed.Truncate(time.Second),
		)),
	)
}

// renderStatusRow returns the styled status line.
func (m *Model) renderStatusRow() string {
	return elapsedEndStyle().Render(m.statusText())
}

// statusText returns the human‑readable scanner state.
func (m *Model) statusText() string {
	idx := m.currentTab
	status := m.status[idx]
	p := m.progressInfo[idx]

	switch status {

	case StatusPreProcess:
		return "preparing scan..."

	case StatusScanning:

		if m.scn.IsPaused() {
			return "scan paused..."
		}

		return m.estimateRemaining(p)

	case StatusEnded:
		return "scan completed"

	case StatusError:
		if m.scanError != nil {
			return fmt.Sprintf("scan error: %v", m.scanError)
		}
		return "scan error"

	default:
		return "starting scan..."
	}
}

// estimateRemaining formats the ETA text.
func (m *Model) estimateRemaining(p engine.Progress) string {

	left := time.Until(p.EstimatedEnd).Truncate(time.Second)

	if left <= 0 {
		return "estimating remaining time..."
	}

	return fmt.Sprintf("estimated remaining: %v", left)
}

