package iplist

// View renders the IP list component.
//
// The view is entirely delegated to the underlying table component,
// which is responsible for layout, styling, and key help rendering.
//
// This method satisfies the ui.Component interface and ensures the
// IP list remains a thin orchestration layer over the table UI.
func (m *Model) View() string {
	if m.table == nil {
		return ""
	}

	return m.table.View()
}
