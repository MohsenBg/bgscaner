package progress

// View renders the progress bar.
//
// Rendering is delegated to the underlying BubbleTea progress model.
// The returned string represents the current visual state of the
// progress bar and is intended to be placed inside the application
// layout body.
func (m *Model) View() string {
	return m.progress.View()
}

