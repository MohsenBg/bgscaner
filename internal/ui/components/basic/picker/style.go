package picker

import (
	"bgscan/internal/ui/shared/layout"
	"bgscan/internal/ui/theme"

	"github.com/charmbracelet/lipgloss"
)

//
// ────────────────────────────────────────────────────────────────────────────────
//  Layout Helpers
// ────────────────────────────────────────────────────────────────────────────────
//

// pickerHeight calculates the vertical height available for the picker overlay
// based on the current application layout.
//
// The picker reserves space for:
//
//   - header / title
//   - footer help text
//   - margins and padding
//
// This ensures the picker fits inside the main body region without overlapping
// other UI components.
func pickerHeight(layout *layout.Layout) int {
	return layout.Body.Height - 10
}

//
// ────────────────────────────────────────────────────────────────────────────────
//  Container Styles
// ────────────────────────────────────────────────────────────────────────────────
//

// containerStyle returns the base style used for the picker container.
//
// Behavior:
//
//   - Defines the overall width and height of the picker
//   - Aligns content to the top-left
//   - Adds horizontal padding for readability
//   - Adds vertical margin to visually separate the picker from surrounding UI
func containerStyle(width, height int) lipgloss.Style {
	return lipgloss.NewStyle().
		Width(width).
		Height(height).
		Align(lipgloss.Left, lipgloss.Top).
		Padding(0, 1).
		Margin(1, 0)
}

//
// ────────────────────────────────────────────────────────────────────────────────
//  Component Styles
// ────────────────────────────────────────────────────────────────────────────────
//

// TitleStyle returns the style used to render the picker title.
//
// Visual characteristics:
//
//   - Center-aligned text
//   - Bold emphasis
//   - Informational theme color
//   - Bottom padding to create spacing between the title and list content
func TitleStyle(width int) lipgloss.Style {
	return lipgloss.NewStyle().
		Width(width-5).
		Align(lipgloss.Center).
		Bold(true).
		Foreground(theme.Current().Info).
		Padding(0, 0, 2, 0).
		BorderForeground(lipgloss.Color("240"))
}

// currentDirStyle returns the style used to display the current directory
// when the picker is operating in filesystem navigation mode.
//
// Visual characteristics:
//
//   - Left-aligned path text
//   - Bottom border for separation from file list
//   - Highlighted color (yellow) to emphasize the active directory
//   - Bold text for improved visibility
func currentDirStyle(width int) lipgloss.Style {
	return lipgloss.NewStyle().
		Width(width-5).
		Align(lipgloss.Left).
		Bold(true).
		Foreground(theme.Current().Yellow).
		Border(lipgloss.NormalBorder(), false, false, true, false)
}

//
// ────────────────────────────────────────────────────────────────────────────────
//  Help Footer Styles
// ────────────────────────────────────────────────────────────────────────────────
//

// helpStyle returns the style used for the picker footer help text.
//
// This area typically displays navigation hints such as:
//
//	↑ ↓ navigate
//	Enter select
//	Esc cancel
//
// Visual characteristics:
//
//   - Muted theme color
//   - Center-aligned text
//   - Top padding to separate from picker content
func helpStyle(width int) lipgloss.Style {
	return lipgloss.NewStyle().
		Width(width - 5).
		Foreground(theme.Current().Muted).
		Align(lipgloss.Center).
		PaddingTop(1)
}

// helpKeyStyle returns the style used to highlight keyboard keys
// inside the help footer (e.g., "Enter", "Esc", "↑", "↓").
//
// Visual characteristics:
//
//   - Secondary theme color
//   - Bold weight for emphasis
func helpKeyStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(theme.Current().Secondary).
		Bold(true)
}
