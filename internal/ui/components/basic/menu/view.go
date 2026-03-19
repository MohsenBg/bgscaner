package menu

// View renders the menu to a string.
func (m *Model) View() string {
	return m.List.View()
}
