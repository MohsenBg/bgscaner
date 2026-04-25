package confirm

import (
	"bgscan/internal/ui/theme"

	"github.com/charmbracelet/lipgloss"
)

//
// ────────────────────────────────────────────────────────────────────────────────
//  Styles for Confirm Dialog
// ────────────────────────────────────────────────────────────────────────────────
//

// MessageStyle returns the base lipgloss style used for the question text
// displayed in the confirm dialog.
//
// Width:
//
//	The provided width is applied using lipgloss.Style.Width() to ensure the
//	message is properly centered and horizontally aligned.
//
// Appearance:
//   - Bold text
//   - Foreground color from the active theme (theme.Current().Text)
//   - Padding around the message to give spacing within the dialog box
//   - Horizontal alignment centered for a symmetrical layout
func MessageStyle(width int) lipgloss.Style {
	return lipgloss.NewStyle().
		Width(width).
		Bold(true).
		Foreground(theme.Current().Text).
		Padding(1, 2).
		Align(lipgloss.Center)
}

// ButtonStyle returns the default style for an **inactive** confirm dialog button.
//
// Visual properties:
//   - Muted foreground color (theme.Current().Muted)
//   - Rounded border styled with the inactive border color
//   - Horizontal padding for UI consistency
//
// This style is used for buttons that are not currently focused/selected.
func ButtonStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(theme.Current().Muted).
		Padding(0, 2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.Current().Border)
}

// SelectButtonStyle returns the style for the **active/selected** confirm dialog
// button. It represents the current user focus.
//
// Visual properties:
//   - Bold text
//   - Foreground rendered using the primary theme color
//   - Rounded border with active border color
//   - Padding matching inactive buttons for consistent layout
//
// This style is typically applied when navigating between "Yes"/"No" options
// using keyboard focus.
func SelectButtonStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Bold(true).
		Foreground(theme.Current().Primary).
		Padding(0, 2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.Current().BorderActive)
}
