package progress

import (
	"bgscan/internal/ui/shared/env"
	"bgscan/internal/ui/shared/layout"
	"bgscan/internal/ui/shared/ui"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
)

const (
	// padding defines the horizontal padding applied around the progress bar.
	padding = 1

	// maxWidth limits the maximum width of the progress bar to prevent
	// overly wide rendering on large terminals.
	maxWidth = 90
)


// Model represents a reusable progress bar UI component.
//
// It wraps the BubbleTea progress model and integrates it with the
// application's component system. The component automatically adapts
// its width to the available layout space.
type Model struct {
	id   ui.ComponentID
	name string

	layout *layout.Layout
	progress progress.Model

	// percent represents the current progress value (0.0 → 1.0).
	percent float64
}


// New creates a new progress bar component.
//
// The progress bar uses the default BubbleTea gradient style and
// automatically adapts its width based on the available layout body width.
func New(layout *layout.Layout) *Model {

	p := progress.New(progress.WithDefaultGradient())

	m := &Model{
		id:       ui.NewComponentID(),
		name:     "Progress",
		progress: p,
		layout:   layout,
		percent:  0,
	}

	m.progress.Width = m.Width()
	m.progress.PercentFormat = " %0.2f%%"

	return m
}


// Init initializes the component.
//
// The progress component does not require any startup commands.
func (m *Model) Init() tea.Cmd {
	return nil
}


// Width calculates the progress bar width based on the layout body.
//
// The width is constrained by:
//   - the layout body width
//   - the defined maximum width
//   - internal horizontal padding
func (m *Model) Width() int {
	width := min(m.layout.Body.Width, maxWidth)
	return width - padding*2
}


// ID returns the unique component identifier.
func (m *Model) ID() ui.ComponentID {
	return m.id
}


// Name returns the human‑readable component name.
func (m *Model) Name() string {
	return m.name
}


// OnClose is called when the component is removed from the UI stack.
func (m *Model) OnClose() tea.Cmd {
	return nil
}


// Mode returns the interaction mode of the component.
//
// The progress component operates in NormalMode and does not
// capture exclusive input.
func (m *Model) Mode() env.Mode {
	return env.NormalMode
}

