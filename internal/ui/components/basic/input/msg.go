package input

import (
	"bgscan/internal/ui/shared/layout"
	"bgscan/internal/ui/shared/ui"

	tea "github.com/charmbracelet/bubbletea"
)

// ShowInputCmd creates a BubbleTea command that opens an input dialog
// as a centered overlay.
//
// Parameters:
//   - layout: shared layout reference used for sizing
//   - message: text displayed above the input field
//   - placeholder: placeholder shown when the input is empty
//   - value: initial value of the input field
//   - validationFunc: optional validation function for the input
//   - onCancel: callback executed when the dialog is cancelled
//   - onConfirm: callback executed when the input is confirmed
func ShowInputCmd(
	layout *layout.Layout,
	message string,
	placeholder string,
	value string,
	validationFunc func(string) (bool, string),
	onCancel func(string) tea.Cmd,
	onConfirm func(string) tea.Cmd,
) tea.Cmd {
	return func() tea.Msg {

		model := New(
			layout,
			message,
			placeholder,
			validationFunc,
			onCancel,
			onConfirm,
		)

		// Set initial input value
		model.textinput.SetValue(value)

		return ui.AddNewOverlay(
			model,
			ui.Center,
			ui.Center,
			0,
			0,
		)
	}
}
