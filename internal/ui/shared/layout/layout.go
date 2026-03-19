package layout

// Layout describes the complete screen geometry of the TUI.
//
// It is recalculated whenever the terminal size changes and
// provides precomputed regions for each major UI component
// (header, body, footer).
//
// The Layout type is intentionally stateful to avoid
// repeated layout calculations inside render paths.
type Layout struct {
	Terminal TerminalSize
	Content  ContentSize

	Header ComponentSize
	Body   ComponentSize
	Footer ComponentSize
}

// TerminalSize represents the raw terminal dimensions.
type TerminalSize struct {
	Width  int
	Height int
}

// ContentSize represents the drawable content area inside
// the terminal, excluding outer margins and padding.
type ContentSize struct {
	Width   int
	Height  int
	Padding int
}

// ComponentSize represents the position and size of a UI
// component in terminal coordinates.
type ComponentSize struct {
	Width  int
	Height int
	X      int
	Y      int
}

// New creates a new Layout with sensible default dimensions.
// The layout must be updated with the actual terminal size
// using Update() before rendering.
func New() *Layout {
	return &Layout{
		Terminal: TerminalSize{
			Width:  80,
			Height: 24,
		},
	}
}

// Update recalculates the entire layout based on the current
// terminal dimensions.
//
// This method should be called whenever a tea.WindowSizeMsg
// is received.
func (l *Layout) Update(termWidth, termHeight int) {
	l.Terminal.Width = termWidth
	l.Terminal.Height = termHeight

	// Content area inside terminal borders
	l.Content = ContentSize{
		Width:   termWidth - 4,
		Height:  termHeight - 2,
		Padding: 1,
	}

	// ─── Header ───────────────────────────
	l.Header = ComponentSize{
		Width:  l.Content.Width,
		Height: 8,
		X:      0,
		Y:      0,
	}

	// ─── Body ─────────────────────────────
	l.Body = ComponentSize{
		Width:  l.Content.Width,
		Height: l.Content.Height - (l.Header.Height + 2),
		X:      0,
		Y:      l.Header.Height,
	}

	// ─── Footer ───────────────────────────
	l.Footer = ComponentSize{
		Width:  l.Content.Width,
		Height: 2,
		X:      0,
		Y:      l.Body.Y + l.Body.Height,
	}
}

// BodyContentWidth returns the drawable width of the body.
func (l *Layout) BodyContentWidth() int {
	return l.Body.Width
}

// BodyContentHeight returns the drawable height of the body.
func (l *Layout) BodyContentHeight() int {
	return l.Body.Height
}
