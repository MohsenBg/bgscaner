package body

import (
	"bgscan/ui/shared/layout"
	"bgscan/ui/shared/nav"
	"bgscan/ui/shared/overlay"

	entryMenu "bgscan/ui/components/menus/entry"

	tea "github.com/charmbracelet/bubbletea"
)

/*
	====================
	Overlay Placement
	====================
*/

type OverlayPlacement struct {
	X    overlay.Position
	Y    overlay.Position
	XOff int
	YOff int
}

var defaultOverlayPlacement = OverlayPlacement{
	X:    overlay.Center,
	Y:    overlay.Center,
	XOff: 0,
	YOff: 0,
}

/*
	====================
	Application Model
	====================
*/

type Model struct {
	layout *layout.Layout

	// Navigation
	viewStack []nav.ViewName

	// Overlays
	overlays          []overlay.Overlay
	overlayPlacements map[overlay.OverlayID]OverlayPlacement

	// Root view
	rootView tea.Model
}

/*
	====================
	Constructor
	====================
*/

func New(layout *layout.Layout) Model {
	return Model{
		layout:            layout,
		rootView:          entryMenu.New(layout),
		viewStack:         []nav.ViewName{nav.ViewMainMenu},
		overlayPlacements: make(map[overlay.OverlayID]OverlayPlacement),
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

/*
	====================
	Navigation Helpers
	====================
*/

// CurrentView returns the active view
func (m Model) CurrentView() nav.ViewName {
	return m.viewStack[len(m.viewStack)-1]
}

// PushView pushes a new view onto the stack
func (m *Model) PushView(v nav.ViewName) {
	m.viewStack = append(m.viewStack, v)
}

// PopView removes the current view
func (m *Model) PopView() {
	if len(m.viewStack) > 1 {
		m.viewStack = m.viewStack[:len(m.viewStack)-1]
	}
}

/*
	====================
	Overlay Helpers
	====================
*/

// GetOverlayPlacement returns the placement for an overlay,
// or a default placement if none is registered.
func (m Model) GetOverlayPlacement(id overlay.OverlayID) OverlayPlacement {
	if p, ok := m.overlayPlacements[id]; ok {
		return p
	}
	return defaultOverlayPlacement
}

// SetOverlayPlacement registers placement for an overlay.
func (m *Model) SetOverlayPlacement(id overlay.OverlayID, placement OverlayPlacement) {
	m.overlayPlacements[id] = placement
}
