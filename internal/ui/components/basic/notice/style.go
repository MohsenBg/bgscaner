package notice

import (
	"bgscan/internal/ui/theme"

	"github.com/charmbracelet/lipgloss"
)

//
// ────────────────────────────────────────────────────────────────────────────────
//  Notice Level Palette
// ────────────────────────────────────────────────────────────────────────────────
//

// levelStyle defines the visual palette used for a specific notice level.
// It encapsulates all color and text elements required to render a notice
// consistently across the UI.
type levelStyle struct {
	TitleColor  lipgloss.TerminalColor
	BorderColor lipgloss.TerminalColor
	AccentColor lipgloss.TerminalColor
	Background  lipgloss.TerminalColor

	Icon       string
	FooterText string
}

// levelPalette returns the style palette associated with a given notice level.
//
// The palette controls:
//
//   - Title color
//   - Border color
//   - Accent elements
//   - Icon prefix
//   - Footer button label
//
// This centralizes all visual decisions for notice severity levels,
// ensuring consistent styling across the application.
func levelPalette(level LEVEL) levelStyle {
	switch level {

	case NOTICE_ERROR:
		return levelStyle{
			TitleColor:  theme.Current().Error,
			BorderColor: theme.Current().Error,
			AccentColor: theme.Current().Error,
			Icon:        "[×] ",
			FooterText:  "Continue",
		}

	case NOTICE_SUCCESS:
		return levelStyle{
			TitleColor:  theme.Current().Success,
			BorderColor: theme.Current().Success,
			AccentColor: theme.Current().Success,
			Icon:        "[✓] ",
			FooterText:  "Done",
		}

	case NOTICE_INFO:
		fallthrough

	default:
		return levelStyle{
			TitleColor:  theme.Current().Info,
			BorderColor: theme.Current().Info,
			AccentColor: theme.Current().Info,
			Icon:        "[i] ",
			FooterText:  "Continue",
		}
	}
}

//
// ────────────────────────────────────────────────────────────────────────────────
//  Layout Styles
// ────────────────────────────────────────────────────────────────────────────────
//

// containerStyle returns the base layout style used for the notice container.
//
// Behavior:
//
//   - Sets the component width
//   - Aligns content to the top-left
//   - Acts as the structural wrapper for the notice UI
func containerStyle(width int) lipgloss.Style {
	return lipgloss.NewStyle().
		Width(width).
		Align(lipgloss.Left, lipgloss.Top)
}

// CenterStyle returns a helper style used to horizontally center
// content inside the notice body.
//
// Behavior:
//
//   - Reduces width slightly to prevent border overlap
//   - Centers text horizontally
func CenterStyle(width int) lipgloss.Style {
	return lipgloss.NewStyle().
		Width(width - 2).
		Align(lipgloss.Center)
}

//
// ────────────────────────────────────────────────────────────────────────────────
//  Component Styles
// ────────────────────────────────────────────────────────────────────────────────
//

// titleStyle returns the style used to render the notice title.
//
// Behavior:
//
//   - Width constrained to container width
//   - Center-aligned text
//   - Bold emphasis
//   - Color determined by notice level palette
//   - Bottom margin to separate title from body text
func titleStyle(width int, level LEVEL) lipgloss.Style {
	p := levelPalette(level)

	return lipgloss.NewStyle().
		Width(width).
		Align(lipgloss.Center).
		Bold(true).
		Foreground(p.TitleColor).
		MarginBottom(1)
}

// ButtonStyle returns the style used for notice action buttons
// such as "Continue", "Done", or confirmation controls.
//
// Visual characteristics:
//
//   - Rounded border
//   - Center-aligned label
//   - Horizontal padding for click-area feel
//   - Active border color from theme
//   - Top margin to separate it from notice content
func ButtonStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(theme.Current().Primary).
		Align(lipgloss.Center).
		Padding(0, 2).
		MarginTop(1).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.Current().BorderActive)
}
