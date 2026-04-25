package table

import (
	"bgscan/internal/ui/theme"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
)

//
// ────────────────────────────────────────────────────────────────────────────────
//  Table Core Styles
// ────────────────────────────────────────────────────────────────────────────────
//

// tableStyles configures the visual styling used by the Bubble Tea table
// component across the application.
//
// Customizations:
//
//   - Header border styling and color
//   - Header bottom border for column separation
//   - Row selection highlighting
//
// This function extends the default Bubble table styles while applying
// the application's theme palette.
func tableStyles() table.Styles {
	s := table.DefaultStyles()

	// Header styling
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(theme.Current().Info).
		BorderBottom(true).
		Padding(0, 1)

	// Selected row styling
	s.Selected = s.Selected.
		Foreground(theme.Current().Text).
		Background(theme.Current().Purple).
		Height(1).
		Bold(true)

	return s
}

//
// ────────────────────────────────────────────────────────────────────────────────
//  Layout Styles
// ────────────────────────────────────────────────────────────────────────────────
//

// titleStyles returns the style used to render the table title.
//
// Visual characteristics:
//
//   - Full-width title container
//   - Center-aligned text
//   - Bold emphasis
//   - Informational theme color
//   - Vertical padding to separate it from the table body
func titleStyles(width int) lipgloss.Style {
	return lipgloss.NewStyle().
		Width(width).
		Align(lipgloss.Center).
		Foreground(theme.Current().Info).
		Bold(true).
		Padding(1, 0)
}

// tableViewStyle returns the style used for the main table container.
//
// Behavior:
//
//   - Centers the table horizontally
//   - Applies the secondary theme color
//   - Maintains consistent width alignment with surrounding components
func tableViewStyle(width int) lipgloss.Style {
	return lipgloss.NewStyle().
		Width(width).
		Align(lipgloss.Center).
		Foreground(theme.Current().Secondary).
		Padding(0, 0)
}

//
// ────────────────────────────────────────────────────────────────────────────────
//  Help / Footer Styles
// ────────────────────────────────────────────────────────────────────────────────
//

// helpViewStyle returns the style used for the table help footer.
//
// This footer typically contains navigation hints such as:
//
//	↑ ↓ navigate rows
//	Enter select
//	Esc return
//
// Visual characteristics:
//
//   - Center-aligned text
//   - Secondary theme color
//   - Top margin to separate it from the table content
func helpViewStyle(width int) lipgloss.Style {
	return lipgloss.NewStyle().
		Width(width).
		Align(lipgloss.Center).
		Foreground(theme.Current().Secondary).
		MarginTop(1)
}
