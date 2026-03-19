package confirm

import (
	"bgscan/internal/ui/shared/layout"
	"bgscan/internal/ui/shared/ui"

	tea "github.com/charmbracelet/bubbletea"
)

// ExitConfirmCmd returns a BubbleTea command that opens a confirmation dialog
// asking the user whether they want to exit the application.
//
// If the user confirms, the dialog executes tea.Quit which terminates the
// BubbleTea program.
//
// The dialog is displayed as an overlay positioned at the top center
// of the screen.
func ExitConfirmCmd(layout *layout.Layout) tea.Cmd {
	return func() tea.Msg {
		return ui.AddNewOverlay(
			New(layout, "Are you sure you want to exit?", func() tea.Cmd { return tea.Quit }, false),
			ui.Center,
			ui.Top,
			0,
			0,
		)
	}
}

// ConfirmCmd returns a BubbleTea command that opens a generic confirmation
// dialog overlay.
//
// Parameters:
//
//	layout     - shared layout manager used to determine component sizing
//	message    - confirmation message displayed in the dialog
//	confirm    - command executed if the user confirms the action
//	defaultYes - determines the initial selection state (true = Yes, false = No)
//
// The dialog is positioned at the top center of the screen and managed by
// the application's overlay system.
func ConfirmCmd(
	layout *layout.Layout,
	message string,
	confirm tea.Cmd,
	defaultYes bool,
) tea.Cmd {
	return func() tea.Msg {
		return ui.AddNewOverlay(
			New(layout, message, func() tea.Cmd { return confirm }, defaultYes),
			ui.Center,
			ui.Top,
			0,
			0,
		)
	}
}

// CloseCmd returns a command that closes this confirmation dialog.
//
// It emits a ui.CloseComponentMsg containing the component's ID, which is
// handled by the UI overlay manager to remove the dialog from the component
// stack.
func (m *Model) CloseCmd() tea.Cmd {
	return func() tea.Msg {
		return ui.CloseComponentMsg{ID: m.ID()}
	}
}
