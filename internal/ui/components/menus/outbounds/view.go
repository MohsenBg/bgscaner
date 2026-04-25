package outbounds

// View renders the outbound templates component.
//
// The actual UI is fully delegated to the underlying table component,
// which handles rendering, layout, styling, and key bindings.
//
// This method satisfies the ui.Component interface and keeps the
// Outbound Templates component as a thin orchestration layer that
// simply forwards rendering to the table.
func (m *Model) View() string {
	if m.table == nil {
		return ""
	}
	return m.table.View()
}
