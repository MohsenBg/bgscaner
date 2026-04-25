package input

import (
	"bgscan/internal/ui/shared/env"
	"bgscan/internal/ui/shared/layout"
	"bgscan/internal/ui/shared/ui"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// Model represents an input dialog component that collects
// user text using a BubbleTea textinput.
//
// It supports optional validation and provides callbacks for
// cancel and confirm actions.
type Model struct {
	// Component identity
	id   ui.ComponentID
	name string

	// Layout reference
	layout *layout.Layout

	// UI content
	message     string
	placeholder string

	// Input field
	textinput textinput.Model

	// Validation
	validationFunc    func(input string) (bool, string)
	dynamicValidation bool
	errorMsg          string

	// Callbacks
	cancelFunc  func(input string) tea.Cmd
	confirmFunc func(input string) tea.Cmd
}

// New creates a new input component.
//
// Parameters:
//   - layout: shared layout reference for sizing
//   - message: message displayed above the input field
//   - placeholder: placeholder text inside the input
//   - validationFunc: optional validation function
//   - cancel: callback executed when the user cancels input
//   - confirm: callback executed when the user confirms input
func New(
	layout *layout.Layout,
	message string,
	placeholder string,
	validationFunc func(input string) (bool, string),
	cancel func(input string) tea.Cmd,
	confirm func(input string) tea.Cmd,
) *Model {

	ti := textinput.New()
	ti.Placeholder = placeholder
	ti.CharLimit = 256
	ti.Focus()

	m := &Model{
		id:                ui.NewComponentID(),
		name:              "input",
		layout:            layout,
		message:           message,
		placeholder:       placeholder,
		textinput:         ti,
		validationFunc:    validationFunc,
		cancelFunc:        cancel,
		confirmFunc:       confirm,
		dynamicValidation: false,
	}

	m.textinput.Width = m.Width()

	return m
}

// Init initializes the component.
func (m *Model) Init() tea.Cmd {
	return nil
}

// ID returns the component identifier.
func (m *Model) ID() ui.ComponentID {
	return m.id
}

// Name returns the component name.
func (m *Model) Name() string {
	return m.name
}

// Width calculates the maximum width of the input field.
func (m *Model) Width() int {
	if m.layout == nil {
		return 50
	}

	return min(50, m.layout.Body.Width)
}

// CloseCmd returns a command that closes this component.
func (m *Model) CloseCmd() tea.Cmd {
	return func() tea.Msg {
		return ui.CloseComponentMsg{ID: m.ID()}
	}
}

// OnClose is called when the component is removed.
func (m *Model) OnClose() tea.Cmd {
	return nil
}

// Mode returns the input mode used by this component.
func (m *Model) Mode() env.Mode {
	return env.InputMode
}
