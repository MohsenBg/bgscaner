package entry

import (
	"bgscan/ui/components/basic/menu"
	"bgscan/ui/components/menus/iplist"
	scanMenu "bgscan/ui/components/menus/scan"
	settingsMenu "bgscan/ui/components/menus/settings"
	"bgscan/ui/shared/layout"
	"bgscan/ui/shared/nav"

	tea "github.com/charmbracelet/bubbletea"
)

// Model is the main app model with a menu stack
type Model struct {
	Views           map[nav.ViewName]nav.View
	Layout          *layout.Layout
	CurrentViewName nav.ViewName
}

// New creates a new entry model with main menu and subviews
func New(layout *layout.Layout) Model {
	entry := Model{
		Layout:          layout,
		CurrentViewName: nav.ViewMainMenu,
		Views: map[nav.ViewName]nav.View{
			nav.ViewMainMenu: {
				Name:  nav.ViewMainMenu,
				Model: newMainMenu(layout),
			},
			nav.ViewScanMenu: {
				Name:  nav.ViewScanMenu,
				Model: scanMenu.New(layout),
			},
			nav.ViewSettingsMenu: {
				Name:  nav.ViewSettingsMenu,
				Model: settingsMenu.New(layout),
			},
			nav.ViewIPList: {
				Name:  nav.ViewIPList,
				Model: iplist.New(layout),
			},
		},
	}

	return entry
}

// Init satisfies Bubble Tea interface
func (m Model) Init() tea.Cmd {
	return nil
}

// CurrentView returns the active Bubble Tea model
func (m Model) CurrentView() tea.Model {
	if view, exists := m.Views[m.CurrentViewName]; exists {
		return view.Model
	}
	return nil
}

// ----------------------
// helpers
// ----------------------

func newMainMenu(layout *layout.Layout) menu.Model {

	items := []menu.MenuItem{
		menu.NewMenuItemWithMsg("▶", "Run Scan", "s", nav.ViewScanMenu),
		menu.NewMenuItemWithMsg("⚙", "Settings", "c", nav.ViewSettingsMenu),
		menu.NewMenuItemWithMsg("☰", "IP Lists", "i", nav.ViewIPList),
		menu.NewMenuItemWithMsg("▲", "Results", "r", nav.ViewResultsMenu),
		menu.NewMenuItemWithMsg("ⓘ", "About", "a", nav.ViewAboutMenu),
	}

	return menu.New(items, "Main Menu", layout)
}
