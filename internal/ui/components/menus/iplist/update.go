package iplist

import (
	"bgscan/internal/core/iplist"
	"bgscan/internal/ui/components/basic/table"
	"bgscan/internal/ui/shared/ui"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/dustin/go-humanize"
)

// Update processes incoming BubbleTea messages and updates the component state.
//
// The method reacts to domain‑specific messages emitted by the IP list workflow:
//
//   - AddIPFileMsg: inserts a newly created file and triggers a list refresh
//   - IPFilesLoadedMsg: replaces the entire file list and rebuilds the table
//   - RequestDeleteIPFileMsg: opens the delete confirmation flow
//   - DeleteIPFileMsg: removes the file from disk and refreshes the list
//   - RequestRenameIPFileMsg: opens the rename input dialog
//   - SelectMsg: triggers the configured file selection callback
//
// Messages not handled by this component are forwarded to the internal
// table component so it can process navigation and key bindings.
func (m *Model) Update(msg tea.Msg) (ui.Component, tea.Cmd) {
	switch msg := msg.(type) {

	case AddIPFileMsg:
		// Insert the new file at the beginning of the list so that
		// newly added files appear at the top of the table.
		m.files = append([]iplist.IPFileInfo{msg.File}, m.files...)

		// Trigger a full list reload message to rebuild table state.
		return m, func() tea.Msg {
			return IPFilesLoadedMsg{Files: m.files}
		}

	case IPFilesLoadedMsg:
		// Replace the internal file list with the authoritative state.
		m.files = msg.Files

		var cursor int

		t, ok := m.table.(*table.Model)
		if !ok || t == nil {
			return m, m.errorCmd("Internal Error", "Cannot access table")
		}

		// Preserve the user's current cursor position.
		cursor = t.BubbleTable.Cursor()

		// Clear existing rows before rebuilding the table.
		t.BubbleTable.SetRows([]table.Row{})

		for i, f := range msg.Files {
			m.filesMap[f.Name] = &msg.Files[i]

			t.AppendRow(table.Row{
				f.Name,
				f.CreatedAt.Format("2006-01-02 15:04:05"),
				humanize.Bytes(uint64(f.Size)),
			})
		}

		// Restore cursor position.
		t.BubbleTable.SetCursor(cursor)

		m.table = t
		return m, nil

	case RequestDeleteIPFileMsg:
		return m, m.handleRequestDelete()

	case DeleteIPFileMsg:
		return m, m.deleteIPFile(msg.File)

	case RequestRenameIPFileMsg:
		return m, m.handleRequestRename()

	case SelectMsg:
		return m, m.handleIPFileSelect()
	}

	// Delegate unhandled messages to the table component.
	var cmd tea.Cmd
	m.table, cmd = m.table.Update(msg)

	return m, cmd
}
