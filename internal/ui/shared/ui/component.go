package ui

import (
	"bgscan/internal/ui/shared/env"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/google/uuid"
)

// ComponentID uniquely identifies a UI component instance.
type ComponentID string

// NewComponentID generates a new unique ComponentID.
func NewComponentID() ComponentID {
	return ComponentID(uuid.NewString())
}

// Component represents a UI module that can be mounted
// and managed by the application.
//
// Components encapsulate their own state, update logic,
// and rendering behavior.
type Component interface {

	// ID returns the unique identifier of the component instance.
	ID() ComponentID

	// Name returns a human readable component name.
	Name() string

	// Init is called when the component is first mounted.
	Init() tea.Cmd

	// Update handles incoming BubbleTea messages and updates
	// the component state.
	Update(tea.Msg) (Component, tea.Cmd)

	// View renders the component UI.
	View() string

	// OnClose is executed when the component is removed
	// from the component stack.
	OnClose() tea.Cmd

	// Mode returns the input mode the component operates in.
	Mode() env.Mode
}

// CloseComponentMsg signals that a component should be closed.
type CloseComponentMsg struct {
	ID ComponentID
}

// OpenComponentMsg requests opening a new component.
type OpenComponentMsg struct {
	Component Component
}

// ResetComponentStacksMsg clears all component stacks.
type ResetComponentStacksMsg struct{}

// OpenComponentCmd creates a BubbleTea command that
// emits an OpenComponentMsg for mounting a new component.
func OpenComponentCmd(component Component) tea.Cmd {
	return func() tea.Msg {
		return OpenComponentMsg{
			Component: component,
		}
	}
}
