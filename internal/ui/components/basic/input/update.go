package input

import (
	"bgscan/internal/ui/shared/ui"

	tea "github.com/charmbracelet/bubbletea"
)

// Update processes incoming Bubble Tea messages and updates the state
// of the input component.
//
// The method delegates message handling to the underlying textinput
// component first, then processes higher-level input dialog behavior
// such as submission and validation.
//
// Behavior:
//   - tea.WindowSizeMsg: adjusts the input width based on the layout.
//   - tea.KeyEnter: attempts to submit the input value.
//   - Other keys: may trigger dynamic validation if enabled.
//
// It returns the updated component and an optional command to execute.
func (m *Model) Update(msg tea.Msg) (ui.Component, tea.Cmd) {
	var cmd tea.Cmd

	// Always update the underlying text input first
	m.textinput, cmd = m.textinput.Update(msg)

	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.textinput.Width = m.Width()
		return m, cmd

	case tea.KeyMsg:
		value := m.textinput.Value()

		switch msg.Type {

		case tea.KeyEnter:
			return m.handleSubmit(value, cmd)

		default:
			if m.dynamicValidation && m.validationFunc != nil {
				_, m.errorMsg = m.validationFunc(value)
			}
		}
	}

	return m, cmd
}

// handleSubmit validates the provided input value and decides whether
// the input dialog should be submitted.
//
// If no validation function is defined, the input is submitted
// immediately. Otherwise the validation function is executed:
//
//   - If validation fails, the error message is stored and the dialog
//     remains open.
//   - If validation succeeds, the dialog is closed and the confirm
//     callback is triggered.
//
// Dynamic validation is enabled after the first submit attempt so that
// subsequent typing updates the validation state in real time.
func (m *Model) handleSubmit(value string, inputCmd tea.Cmd) (ui.Component, tea.Cmd) {

	if m.validationFunc == nil {
		return m, tea.Batch(inputCmd, m.submit(value))
	}

	valid, err := m.validationFunc(value)

	m.errorMsg = err
	m.dynamicValidation = true

	if !valid {
		return m, inputCmd
	}

	return m, tea.Batch(inputCmd, m.submit(value))
}

// submit finalizes the input process and produces the command that
// closes the dialog and optionally triggers the confirm callback.
//
// If a confirm callback is defined, it is executed with the submitted
// input value. Otherwise the dialog simply closes.
func (m *Model) submit(input string) tea.Cmd {
	if m.confirmFunc == nil {
		return m.CloseCmd()
	}

	return tea.Batch(
		m.CloseCmd(),
		m.confirmFunc(input),
	)
}
