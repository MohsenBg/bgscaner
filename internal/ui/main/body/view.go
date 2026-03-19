package body

func (m *Model) View() string {
	view := containerStyle(
		m.layout.Body.Width,
		m.layout.Body.Height,
	).Render(m.components[len(m.components)-1].View())

	return view
}
