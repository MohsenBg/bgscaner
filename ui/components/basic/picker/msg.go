package picker

import (
	"bgscan/ui/shared/layout"
	"bgscan/ui/shared/overlay"

	tea "github.com/charmbracelet/bubbletea"
)

// OnSelect is called when the user selects a file.
type OnSelect func(path string) tea.Cmd

// NewOpenPickFileCmd returns a command that opens a file picker.
func NewOpenPickFileCmd(
	layout *layout.Layout,
	title string,
	baseDir string,
	allowType []string,
	onSelect OnSelect,
) tea.Cmd {
	return func() tea.Msg {
		fp := New(layout, title, baseDir, allowType, onSelect)
		return overlay.NewAddOverlay(fp, overlay.Center, overlay.Center, 0, 0)
	}
}
