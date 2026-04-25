package resultlist

import (
	"bgscan/internal/ui/components/basic/table"
	"bgscan/internal/ui/shared/ui"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/dustin/go-humanize"
)

// Update processes incoming BubbleTea messages and updates the component state.
//
// The method handles high-level actions such as:
//
//   - refreshing the table rows
//   - selecting a result file
//   - requesting rename/delete actions
//   - executing filesystem operations
//
// Messages that are not handled directly are delegated to the embedded
// table component.
func (m *Model) Update(msg tea.Msg) (ui.Component, tea.Cmd) {

	switch msg := msg.(type) {

	case UpdateTableMsg:
		m.updateTableRows()
		return m, nil

	case SelectResultFileMsg:
		return m, m.handleResultFileSelect()

	case RequestDeleteResultFileMsg:
		return m, m.handleRequestDelete()

	case DeleteResultFileMsg:
		return m, m.deleteResultFile(msg.File)

	case RequestRenameResultFileMsg:
		return m, m.handleRequestRename()
	}

	var cmd tea.Cmd
	m.table, cmd = m.table.Update(msg)

	return m, cmd
}

// updateTableRows rebuilds the table rows from the currently loaded result files.
//
// Each row displays:
//
//   - file name
//   - creation time
//   - result type
//   - file size (human readable)
func (m *Model) updateTableRows() {

	rows := make([]table.Row, 0, len(m.resultFiles))

	for _, res := range m.resultFiles {

		rows = append(rows, table.Row{
			res.Name,
			res.CreatedTime.Format("2006-01-02 15:04:05"),
			res.Type.String(),
			humanize.Bytes(uint64(res.SizeBytes)),
		})

	}

	if t, ok := m.table.(*table.Model); ok {
		t.BubbleTable.SetRows(rows)
	}
}
