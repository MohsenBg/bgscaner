package tabs

import (
	"bgscan/internal/ui/shared/env"
	"bgscan/internal/ui/shared/layout"
	"bgscan/internal/ui/shared/ui"

	tea "github.com/charmbracelet/bubbletea"
)

type Tab struct {
	Label string
	Value any
}

func NewTab(label string, value any) Tab {
	return Tab{
		Label: label,
		Value: value,
	}
}

type Model struct {
	layout      *layout.Layout
	id          ui.ComponentID
	name        string
	tabs        []Tab
	onSelectTab func(idx int, tab Tab) tea.Cmd
	idx         int
}

func New(layout *layout.Layout, tabs []Tab, onSelectTab func(idx int, tab Tab) tea.Cmd) *Model {
	return &Model{
		layout:      layout,
		id:          ui.NewComponentID(),
		name:        "tabs",
		tabs:        tabs,
		idx:         0,
		onSelectTab: onSelectTab,
	}
}

// Mode implements
func (m *Model) Mode() env.Mode {
	return env.NormalMode
}

// Init implements the BubbleTea initialization interface.
func (m *Model) Init() tea.Cmd {
	return nil
}

// ID returns the component unique identifier.
func (m *Model) ID() ui.ComponentID {
	return m.id
}

// Name returns the component name.
func (m *Model) Name() string {
	return m.name
}

// OnClose is called when the component is removed from the UI.
func (m *Model) OnClose() tea.Cmd {
	return nil
}

func (m *Model) SelectTab(idx int) *Tab {
	if idx >= 0 && len(m.tabs) > idx {
		m.idx = idx
		return &m.tabs[idx]
	}
	return nil
}

func (m *Model) CurrentTab() *Tab {
	if m.idx >= 0 && len(m.tabs) > m.idx {
		return &m.tabs[m.idx]
	}
	return nil
}

func (m *Model) selectTabCmd() tea.Cmd {
	tab := m.CurrentTab()
	if tab != nil {
		return m.onSelectTab(m.idx, *tab)
	}
	return nil
}

func (m *Model) NextTab() {
	if m.idx+1 < len(m.tabs) {
		m.idx++
		return
	}
	m.idx = 0
}

func (m *Model) BackTab() {
	if m.idx-1 < 0 {
		m.idx = len(m.tabs) - 1
		return
	}
	m.idx--
}
