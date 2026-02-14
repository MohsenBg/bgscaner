package entry

func (m Model) View() string {
	return m.CurrentView().View()
}
