package progress

import (
	"bgscan/internal/ui/shared/ui"

	tea "github.com/charmbracelet/bubbletea"
)

// UpdateProgressMsg is sent to update the progress bar state.
//
// Progress must be a normalized value in the range [0.0, 1.0],
// where:
//   - 0.0 represents no progress
//   - 1.0 represents completion
//
// Values outside this range may be clamped by the receiver.
type UpdateProgressMsg struct {
	ID       ui.ComponentID
	Progress float64
}

//
// ────────────────────────────────────────────────────────────
//   Helper: Convert msg struct to Cmd
// ────────────────────────────────────────────────────────────
//

func (m UpdateProgressMsg) Cmd() tea.Cmd {
	return func() tea.Msg { return m }
}
