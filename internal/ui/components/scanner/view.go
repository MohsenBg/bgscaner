package scanner

import (
	"fmt"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// View renders the scanner UI.
//
// The view consists of two main parts:
//  1. The progress section (statistics, status, progress bar)
//  2. The IP results viewer
//
// Layout structure:
//
//	┌ Progress Information ┐
//	│  Stats Row           │
//	│  Status / ETA Row    │
//	│  Progress Bar        │
//	└──────────────────────┘
//	┌ IP Results Table ┐
func (m *Model) View() string {
	return lipgloss.JoinVertical(
		lipgloss.Left,
		m.renderProgress(),
		m.ipViewer.View(),
	)
}

// renderProgress builds the progress panel shown above the results table.
//
// It displays:
//   - Scan statistics (processed, remaining, found, elapsed)
//   - Current scanner state (preprocessing, scanning, paused, error, etc.)
//   - Estimated remaining time when available
//   - Progress bar
func (m *Model) renderProgress() string {

	p := m.progressInfo
	width := m.layout.Body.Width

	// --- Stats Row ---

	stats := lipgloss.JoinHorizontal(
		lipgloss.Left,
		scannedStyle().Render(fmt.Sprintf("scanned: %d", p.Processed)),
		separatorStyle().Render(" | "),
		leftStyle().Render(fmt.Sprintf("left: %d", p.Total-p.Processed)),
		separatorStyle().Render(" | "),
		foundStyle().Render(fmt.Sprintf("found: %d", p.Succeed)),
		separatorStyle().Render(" | "),
		elapsedStyle().Render(fmt.Sprintf("elapsed: %v", p.Elapsed.Truncate(time.Second))),
	)

	// --- Status / ETA Row ---

	statusText := m.statusText()

	estRow := elapsedEndStyle().Render(statusText)

	container := lipgloss.NewStyle().
		Width(width).
		Align(lipgloss.Center)

	return lipgloss.JoinVertical(
		lipgloss.Center,
		container.Render(stats),
		container.Padding(1, 0).Render(estRow),
		container.Render(m.progress.View()),
	)
}

// statusText returns the human‑readable scanner status message.
//
// The message reflects the current lifecycle state of the scanner:
//
//	PreProcess → "preparing scan..."
//	Scanning   → shows estimated remaining time
//	Paused     → "scan paused..."
//	Ended      → "scan completed"
//	Error      → error message
func (m *Model) statusText() string {

	p := m.progressInfo

	switch m.status {

	case StatusPreProcess:
		return "preparing scan..."

	case StatusScanning:

		if m.scanner.IsPaused() {
			return "scan paused..."
		}

		left := time.Until(p.EstimatedEnd).Truncate(time.Second)

		if left <= 0 {
			return "estimated remaining..."
		}

		return fmt.Sprintf("estimated remaining: %v", left)

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
