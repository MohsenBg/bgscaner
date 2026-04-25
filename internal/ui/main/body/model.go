package body

import (
	mainMenu "bgscan/internal/ui/components/menus/entry"
	"bgscan/internal/ui/shared/env"
	"bgscan/internal/ui/shared/layout"
	"bgscan/internal/ui/shared/ui"

	tea "github.com/charmbracelet/bubbletea"
)

type Model struct {
	id         ui.ComponentID
	name       string
	layout     *layout.Layout
	components []ui.Component
}

func New(layout *layout.Layout) *Model {
	return &Model{
		id:         ui.NewComponentID(),
		name:       "body",
		layout:     layout,
		components: make([]ui.Component, 0, 4),
	}
}

func (m *Model) Init() tea.Cmd {
	m.components = append(m.components, mainMenu.New(m.layout))
	return m.components[0].Init()
}

func (m *Model) ID() ui.ComponentID {
	return m.id
}

func (m *Model) Name() string {
	return m.name
}

func (m *Model) OnClose() tea.Cmd {
	return nil
}

func (m *Model) Mode() env.Mode {
	return env.NormalMode
}
