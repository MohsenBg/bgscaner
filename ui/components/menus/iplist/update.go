package iplist

import (
	"bgscan/ui/components/basic/table"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/dustin/go-humanize"
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case ResultFilesLoadedMsg:
		t, ok := m.Table.(table.Model)
		if !ok {
			break
		}
		t.BubbleTable.SetRows([]table.Row{})

		for _, f := range msg.Files {
			t.AppendRow(table.Row{
				f.Name,
				f.CreatedTime.Format("2006-01-02 15:04:05"),
				humanize.Bytes(uint64(f.SizeBytes)),
				"-",
			})
		}
	}

	var cmd tea.Cmd
	m.Table, cmd = m.Table.Update(msg)
	return m, cmd
}
