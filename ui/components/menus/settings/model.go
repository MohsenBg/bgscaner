package settings

import (
	"bgscan/ui/components/basic/menu"
	"bgscan/ui/shared/layout"
	"bgscan/ui/shared/nav"

	tea "github.com/charmbracelet/bubbletea"
)

// ═══ Model ═══
type Model struct {
	Views           map[nav.ViewName]nav.View
	Layout          *layout.Layout
	CurrentViewName nav.ViewName
}

// ═══ Constructor ═══
func New(layout *layout.Layout) Model {
	items := []menu.MenuItem{
		menu.NewMenuItemWithMsg("📡", "ICMP Config", "i", nav.ViewICMPConfigMenu),
		menu.NewMenuItemWithMsg("🔌", "TCP Config", "t", nav.ViewTCPConfigMenu),
		menu.NewMenuItemWithMsg("🌐", "HTTP Config", "h", nav.ViewHTTPConfigMenu),
		menu.NewMenuItemWithMsg("⚡", "XRay Config", "x", nav.ViewXRayConfigMenu),
	}

	return Model{
		CurrentViewName: nav.ViewSettingsMenu,
		Layout:          layout,
		Views: map[nav.ViewName]nav.View{
			nav.ViewSettingsMenu: {
				Name:  nav.ViewSettingsMenu,
				Model: menu.New(items, "Settings", layout),
			},
		},
	}
}

// ═══ Init ═══
func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) CurrentView() tea.Model {
	view, exist := m.Views[m.CurrentViewName]
	if !exist {
		return nil
	}
	return view.Model
}
