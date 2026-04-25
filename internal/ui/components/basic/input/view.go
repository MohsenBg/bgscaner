package input

import "github.com/charmbracelet/lipgloss"

// View renders the input dialog UI.
//
// The view consists of:
//   - An optional message displayed above the input field
//   - The text input component
//   - An optional validation error message
//   - Key hints for user interaction
//
// The content is vertically stacked and wrapped inside a styled
// container whose width is determined by the layout.
func (m *Model) View() string {
	content := make([]string, 0, 4)

	// Optional message
	if m.message != "" {
		content = append(content, messageStyle().Render(m.message))
	}

	// Input field
	content = append(content, m.textinput.View())

	// Validation error (if present)
	if m.errorMsg != "" {
		content = append(content, errorStyle().Render("✗ "+m.errorMsg))
	}

	// Key hints
	hints := keyHintStyle().Render("Enter to confirm • Esc to cancel")
	content = append(content, hints)

	body := lipgloss.JoinVertical(
		lipgloss.Top,
		content...,
	)

	return containerStyle(m.Width()).Render(body)
}
