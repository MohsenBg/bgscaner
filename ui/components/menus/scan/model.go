package scan

import (
	"bgscan/ui/components/basic/menu"
	"bgscan/ui/shared/layout"
	"bgscan/ui/shared/nav"

	tea "github.com/charmbracelet/bubbletea"
)

const (
	SELECT_IP_FROM_LIST_VIEW   = "SELECT_IP_FROM_LIST_VIEW"
	SELECT_IP_FROM_RESULT_VIEW = "SELECT_IP_FROM_RESULT_VIEW"
)

type Model struct {
	view   nav.ViewName
	layout *layout.Layout
	menu   menu.Model
}

func (m Model) Init() tea.Cmd {
	return nil
}

func New(layout *layout.Layout) Model {
	items := []menu.MenuItem{
		menu.NewMenuItem(
			"≡",
			"Scanned IPs",
			"l",
			func() tea.Cmd {
				return func() tea.Msg { return ShowScannedIPsMsg{} }
			},
		),
		menu.NewMenuItem(
			"▦",
			"Scan Results",
			"r",
			func() tea.Cmd {
				return func() tea.Msg { return ShowScanResultsMsg{} }
			},
		),
	}

	return Model{
		view: SELECT_IP_FROM_LIST_VIEW,
		menu: menu.New(items, "Run Scan", layout),
	}
}
