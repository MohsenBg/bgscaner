package layout

import (
	"bgscan/logger"
)

// Layout
type Layout struct {
	Terminal TerminalSize
	Content  ContentSize
	Header   ComponentSize
	Body     ComponentSize
	Footer   ComponentSize
}

type TerminalSize struct {
	Width  int
	Height int
}

type ContentSize struct {
	Width   int
	Height  int
	Padding int
}

type ComponentSize struct {
	Width  int
	Height int
	X      int
	Y      int
}

func New() *Layout {
	return &Layout{
		Terminal: TerminalSize{Width: 80, Height: 24},
	}
}

// Update
func (l *Layout) Update(termWidth, termHeight int) {
	l.Terminal.Width = termWidth
	l.Terminal.Height = termHeight

	logger.Log("w:%d,h:%d", termWidth, termHeight)

	l.Content = ContentSize{
		Width:   termWidth - 4,
		Height:  termHeight - 2,
		Padding: 1,
	}

	// ═══ Header ═══
	l.Header = ComponentSize{
		Width:  l.Content.Width,
		Height: 8,
		X:      0,
		Y:      0,
	}

	// ═══ Footer ═══
	l.Footer = ComponentSize{
		Width:  l.Content.Width,
		Height: 2,
		X:      0,
		Y:      termHeight,
	}

	// ═══ Body ═══
	l.Body = ComponentSize{
		Width:  l.Content.Width,
		Height: l.Content.Height - (l.Header.Height + l.Footer.Height),
		X:      0,
		Y:      l.Header.Height,
	}
}

// Helper methods
func (l *Layout) BodyContentWidth() int {
	return l.Body.Width
}

func (l *Layout) BodyContentHeight() int {
	return l.Body.Height
}
