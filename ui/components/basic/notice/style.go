package notice

import (
	"bgscan/ui/theme"

	"github.com/charmbracelet/lipgloss"
)

// -------------------- Level Palette --------------------
type levelStyle struct {
	TitleColor  lipgloss.TerminalColor
	BorderColor lipgloss.TerminalColor
	AccentColor lipgloss.TerminalColor
	Background  lipgloss.TerminalColor
	Icon        string
	FooterText  string
}

func levelPalette(level LEVEL) levelStyle {
	switch level {
	case NOTICE_ERROR:
		return levelStyle{
			TitleColor:  theme.Current().Error,
			BorderColor: theme.Current().Error,
			AccentColor: theme.Current().Error,
			Icon:        "[×]",
			FooterText:  "Continue",
		}
	case NOTICE_SUCCESS:
		return levelStyle{
			TitleColor:  theme.Current().Success,
			BorderColor: theme.Current().Success,
			AccentColor: theme.Current().Success,
			Icon:        "[✓]",
			FooterText:  "Done",
		}
	case NOTICE_INFO:
		fallthrough
	default:
		return levelStyle{
			TitleColor:  theme.Current().Info,
			BorderColor: theme.Current().Info,
			AccentColor: theme.Current().Info,
			Icon:        "[i]",
			FooterText:  "Continue",
		}
	}
}

// -------------------- Styling --------------------
func containerStyle(width int) lipgloss.Style {
	return lipgloss.NewStyle().
		Align(lipgloss.Left, lipgloss.Top).
		Width(width)
}

func titleStyle(width int, level LEVEL) lipgloss.Style {
	p := levelPalette(level)
	return lipgloss.NewStyle().
		Width(width).
		Align(lipgloss.Center).
		Bold(true).
		Foreground(p.TitleColor).MarginBottom(1)
}

func CenterStyle(width int) lipgloss.Style {
	return lipgloss.NewStyle().Width(width - 2).Align(lipgloss.Center)
}

// ButtonStyle styles inactive buttons
func ButtonStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(theme.Current().Primary).
		Align(lipgloss.Center).
		Padding(0, 2).
		Border(lipgloss.RoundedBorder()).MarginTop(1).
		BorderForeground(theme.Current().BorderActive)
}
