package outbounds

import (
	"bgscan/internal/core/xray"
	"bgscan/internal/ui/components/basic/table"
	"bgscan/internal/ui/shared/ui"

	tea "github.com/charmbracelet/bubbletea"
)

// Update processes incoming BubbleTea messages and updates the component state.
//
// The method reacts to domain‑specific messages emitted by the outbound
// templates workflow:
//
//   - AddOutboundMsg: inserts a newly created outbound and triggers a list refresh
//   - OutboundsLoadedMsg: replaces the entire outbound list and rebuilds the table
//   - RequestDeleteOutboundMsg: opens the delete confirmation flow
//   - DeleteOutboundMsg: deletes the outbound from disk and refreshes the list
//   - RequestRenameOutboundMsg: opens the rename input dialog
//   - SelectMsg: triggers the configured selection callback
//
// Messages not handled by this component are forwarded to the internal
// table component so it can process navigation and key bindings.
func (m *Model) Update(msg tea.Msg) (ui.Component, tea.Cmd) {
	switch msg := msg.(type) {

	case AddOutboundMsg:
		// Insert new outbound at the beginning (newest first)
		m.outbounds = append([]xray.XrayOutboundsFile{*msg.Outbound}, m.outbounds...)

		// Trigger a full refresh so the table rebuilds properly
		return m, func() tea.Msg {
			return OutboundsLoadedMsg{Outbounds: m.outbounds}
		}

	case OutboundsLoadedMsg:
		// Set authoritative state
		m.outbounds = msg.Outbounds

		t, ok := m.table.(*table.Model)
		if !ok || t == nil {
			return m, m.errorCmd("Internal Error", "Cannot access table")
		}

		// Store current cursor
		cursor := t.BubbleTable.Cursor()

		// Rebuild map and table rows
		t.BubbleTable.SetRows([]table.Row{})
		m.outboundsMap = make(map[string]*xray.XrayOutboundsFile)

		for i, o := range msg.Outbounds {
			m.outboundsMap[o.Name] = &msg.Outbounds[i]

			t.AppendRow(table.Row{
				o.Name,
				o.CreatedTime.Format("2006-01-02 15:04:05"),
			})
		}

		t.BubbleTable.SetCursor(cursor)
		m.table = t

		return m, nil

	case RequestDeleteOutboundMsg:
		return m, m.handleRequestDelete()

	case DeleteOutboundMsg:
		return m, m.deleteOutbound(*msg.Outbound)

	case RequestRenameOutboundMsg:
		return m, m.handleRequestRename()

	case SelectMsg:
		return m, m.handleOutboundSelect()
	}

	// Pass through unhandled messages to table component
	var cmd tea.Cmd
	m.table, cmd = m.table.Update(msg)

	return m, cmd
}
