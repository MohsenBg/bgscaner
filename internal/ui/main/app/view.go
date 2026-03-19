package app

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	overlay "github.com/rmhubbert/bubbletea-overlay"
)

const (
	minWidth  = 75
	minHeight = 35
)

// View renders the entire application UI including
// base components and overlay layers.
func (m model) View() string {
	termWidth := m.layout.Terminal.Width
	termHeight := m.layout.Terminal.Height

	// Prevent rendering when terminal is too small
	if termWidth < minWidth || termHeight < minHeight {
		return m.renderLimitSize(termWidth, termHeight)
	}

	// Render main layout (header, body, footer)
	content := lipgloss.JoinVertical(
		lipgloss.Left,
		m.header.View(),
		m.body.View(),
		m.footer.View(),
	)

	// Render overlays on top of the base content
	content = m.renderOverlays(content)

	return containerStyle(termWidth, termHeight).Render(
		mainStyle(m.layout.Content.Width, m.layout.Content.Height).
			Render(content),
	)
}

// renderLimitSize displays a warning when the terminal
// is smaller than the minimum supported size.
func (m *model) renderLimitSize(termWidth, termHeight int) string {
	msg := fmt.Sprintf(
		"Terminal too small\nMinimum size is %dx%d\nPlease resize your terminal to have more space.",
		minWidth,
		minHeight,
	)

	return centerStyle().
		Width(termWidth).
		Height(termHeight).
		Render(
			warningStyle().Render(msg),
		)
}

// renderOverlays composites overlay components on top
// of the base UI view using the configured placement.
func (m *model) renderOverlays(baseView string) string {
	view := baseView

	for _, layer := range m.layers {
		placement := m.getOverlayPlacement(layer.ID())

		view = overlay.Composite(
			WindowStyle(m.layout.Body.Width).Render(layer.View()),
			view,
			placement.XPos,
			placement.YPos,
			placement.XOffset,
			placement.YOffset,
		)
	}

	return view
}
