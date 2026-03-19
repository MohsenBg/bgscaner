package settings

import (
	"bgscan/internal/ui/components/basic/menu"
	"bgscan/internal/ui/components/basic/notice"
	"bgscan/internal/ui/shared/env"
	"bgscan/internal/ui/shared/layout"
	"bgscan/internal/ui/shared/ui"

	tea "github.com/charmbracelet/bubbletea"
)

// ═══ Model ═══
type Model struct {
	id     ui.ComponentID
	name   string
	Layout *layout.Layout
	menu   ui.Component
}

// ═══ Constructor ═══
func New(layout *layout.Layout) *Model {
	items := []menu.MenuItem{
		menu.NewMenuItem("📡", "ICMP Config", "i", notice.NoticeUnderDevelopment(layout)),
		menu.NewMenuItem("🔌", "TCP Config", "t", notice.NoticeUnderDevelopment(layout)),
		menu.NewMenuItem("🌐", "HTTP Config", "h", notice.NoticeUnderDevelopment(layout)),
		menu.NewMenuItem("⚡", "XRay Config", "x", notice.NoticeUnderDevelopment(layout)),
	}

	return &Model{
		menu:   menu.New(items, "settings", layout),
		id:     ui.NewComponentID(),
		name:   "settings",
		Layout: layout,
	}
}

// ═══ Init ═══
func (m *Model) Init() tea.Cmd {
	return nil
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
