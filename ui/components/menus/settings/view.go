package settings

func (m Model) View() string {
	return m.CurrentView().View()
}
