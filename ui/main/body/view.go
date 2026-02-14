package body

import (
	overlay "github.com/rmhubbert/bubbletea-overlay"
)

func (m Model) View() string {
	view := containerStyle(
		m.layout.Body.Width,
		m.layout.Body.Height,
	).Render(m.rootView.View())

	return view
}

func (m Model) RenderOverlays(view string) string {
	for _, o := range m.overlays {
		p := m.GetOverlayPlacement(o.ID())
		view = overlay.Composite(
			WindowStyle().Render(o.View()),
			view,
			p.X,
			p.Y,
			p.XOff,
			p.YOff,
		)
	}
	return view
}
