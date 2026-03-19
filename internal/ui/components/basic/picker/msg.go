package picker

import (
	"bgscan/internal/ui/shared/layout"
	"bgscan/internal/ui/shared/ui"

	tea "github.com/charmbracelet/bubbletea"
)

// OnSelect defines the callback executed when a file is selected.
//
// The selected file path is passed to the callback, which may return
// a BubbleTea command to trigger additional actions in the application.
type OnSelect func(path string) tea.Cmd

// NewOpenPickFileCmd returns a BubbleTea command that opens a file picker overlay.
//
// When executed, the command creates a new picker component and emits a
// message instructing the application to add it to the overlay stack.
//
// Parameters:
//
//	layout    — UI layout manager used for sizing the picker
//	title     — title displayed at the top of the picker overlay
//	baseDir   — initial directory to open (defaults to home directory)
//	allowType — optional list of allowed file extensions (e.g. ".csv", ".txt")
//	onSelect  — callback executed when the user selects a file
//
// The picker is positioned in the center of the screen using the overlay system.
func NewOpenPickFileCmd(
	layout *layout.Layout,
	title string,
	baseDir string,
	allowType []string,
	onSelect OnSelect,
) tea.Cmd {

	return func() tea.Msg {

		fp := New(layout, title, baseDir, allowType, onSelect)

		return ui.AddNewOverlay(
			fp,
			ui.Center,
			ui.Center,
			0,
			0,
		)
	}
}
