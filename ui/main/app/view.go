package app

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

const (
	minWidth  = 75
	minHeight = 35
)

var ()

func (m model) View() string {
	termWidth := m.layout.Terminal.Width
	termHeight := m.layout.Terminal.Height

	if termWidth < minWidth || termHeight < minHeight {
		msg := fmt.Sprintf(
			"Terminal too small\nMinimum size is %dx%d\nPlease resize your terminal to have more space.",
			minWidth, minHeight,
		)

		return centerStyle().
			Width(termWidth).
			Height(termHeight).
			Render(warningStyle().Render(msg))
	}

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		m.header.View(),
		m.body.View(),
		m.footer.View(),
	)
	content = m.body.RenderOverlays(content)

	return containerStyle(termWidth, termHeight).
		Render(
			mainStyle(m.layout.Content.Width, m.layout.Content.Height).
				Render(content),
		)
}
