package input

import (
	"bgscan/internal/ui/theme"

	"github.com/charmbracelet/lipgloss"
)

//
// ────────────────────────────────────────────────────────────────────────────────
//  Styles for Input Prompt Components
// ────────────────────────────────────────────────────────────────────────────────
//

// containerStyle returns the wrapper style used to define the width of the
// entire input component.
//
// Parameters:
//   - width: the horizontal width in characters for the outer container.
//
// This style does not apply colors or padding; it is purely structural.
func containerStyle(width int) lipgloss.Style {
	return lipgloss.NewStyle().
		Width(width)
}

// messageStyle returns the style used to render the main prompt or label text
// above the input field.
//
// Visual properties:
//   - Bold, readable foreground using theme.Current().Text
//   - Bottom margin to separate it from the input box
func messageStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(theme.Current().Text).
		Bold(true).
		MarginBottom(1)
}

// errorStyle returns the style applied to validation or error messages
// displayed below the input field.
//
// Visual properties:
//   - Bold red-ish text from theme.Current().Error
//   - Top margin for spacing from the previous element (input box or label)
func errorStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(theme.Current().Error).
		Bold(true).
		MarginTop(1)
}

// keyHintStyle returns the style for helpful key hints, such as
// "Press Enter to confirm" or "Tab to switch fields".
//
// Visual properties:
//   - Soft informational color (theme.Current().Info)
//   - Top margin for spacing from the input or error message
func keyHintStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(theme.Current().Info).
		MarginTop(1)
}
