package logs

import (
	"bgscan/internal/logger"
	"bgscan/internal/ui/components/basic/logview"
	"bgscan/internal/ui/components/basic/menu"
	"bgscan/internal/ui/shared/env"
	"bgscan/internal/ui/shared/layout"
	"bgscan/internal/ui/shared/ui"

	tea "github.com/charmbracelet/bubbletea"
)

// Model represents the main logs menu component.
type Model struct {
	id     ui.ComponentID
	name   string
	menu   ui.Component
	layout *layout.Layout
}

// ID returns the component's unique identifier.
func (m *Model) ID() ui.ComponentID {
	return m.id
}

// Name returns the component's display name.
func (m *Model) Name() string {
	return m.name
}

// OnClose is called when the component is closed.
func (m *Model) OnClose() tea.Cmd {
	return nil
}

// New creates and returns a new logs menu model.
func New(l *layout.Layout) *Model {
	return &Model{
		id:     ui.NewComponentID(),
		name:   "Logs Menu",
		layout: l,
		menu:   newLogsMenu(l),
	}
}

func (m *Model) Init() tea.Cmd {
	return nil
}

// newLogsMenu creates the menu with log category options.
func newLogsMenu(l *layout.Layout) *menu.Model {
	items := []menu.MenuItem{
		menu.NewMenuItem("▶", "Core Logs", "c",
			func() tea.Msg {
				return ui.OpenComponentMsg{
					Component: logview.New(l, logger.Core(), "Core Logs"),
				}
			},
		),
		menu.NewMenuItem("⚙", "UI Logs", "u",
			func() tea.Msg {
				return ui.OpenComponentMsg{
					Component: logview.New(l, logger.UI(), "UI Logs"),
				}
			},
		),
		menu.NewMenuItem("☰", "Debug Logs", "d",
			func() tea.Msg {
				return ui.OpenComponentMsg{
					Component: logview.New(l, logger.Debug(), "Debug Logs"),
				}
			},
		),
	}
	return menu.New(items, "Logs Menu", l)
}

func (m *Model) Mode() env.Mode {
	return env.NormalMode
}
