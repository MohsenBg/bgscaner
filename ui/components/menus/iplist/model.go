package iplist

import (
	"bgscan/core/ipmanager"
	"bgscan/logger"
	"bgscan/ui/components/basic/notice"
	"bgscan/ui/components/basic/picker"
	"bgscan/ui/components/basic/table"
	"bgscan/ui/shared/layout"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

type Model struct {
	Layout  *layout.Layout
	Table   tea.Model
	columns []table.Column
	rows    []table.Row
}

func (m Model) Init() tea.Cmd {
	return loadResultFilesCmd(m.Layout, ipmanager.ResultAll)
}

// New creates a new IP list model with randomized rows.
func New(layout *layout.Layout) Model {
	columns := []table.Column{
		{Title: "Name", Width: 30},
		{Title: "Created Time", Width: 30},
		{Title: "Size", Width: 15},
		{Title: "IP Count", Width: 25},
	}

	rows := make([]table.Row, 0)

	t := table.New("IP List", columns, rows, layout)
	t.Keys.Add(table.NewActionKey(
		[]string{"a"},
		"a add file",
		"a add ip file",
		picker.NewOpenPickFileCmd(layout, "Select IP File .txt", "", []string{".txt"}, OnSelect),
	))

	return Model{
		Layout:  layout,
		columns: columns,
		rows:    rows,
		Table:   t,
	}
}

func OnSelect(path string) tea.Cmd {
	logger.Log("path:%s", path)
	return nil
}

func loadResultFilesCmd(layout *layout.Layout, searchType ipmanager.ResultType) tea.Cmd {
	return func() tea.Msg {
		files, err := ipmanager.ListResultFiles(searchType)
		if err != nil {
			msg := fmt.Sprintf("error message:%s", err.Error())
			return notice.NewNoticeCmd(layout, "Error Loading Files", msg, notice.NOTICE_ERROR)
		}
		return ResultFilesLoadedMsg{Files: files}
	}
}
