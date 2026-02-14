package table

import (
	"bgscan/ui/theme"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
)

func tableStyles() table.Styles {
	s := table.DefaultStyles()

	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(theme.Current().Info).
		BorderBottom(true).
		Padding(0, 1)

	s.Selected = s.Selected.
		Foreground(theme.Current().Text).
		Background(theme.Current().AccentPurple).
		Height(1).
		Bold(true)

	return s
}

func titleStyles(width int) lipgloss.Style {
	return lipgloss.NewStyle().
		Width(width).
		Align(lipgloss.Center).
		Foreground(theme.Current().Info).
		Bold(true).
		Padding(1, 0)
}

func tableViewStyle(width int) lipgloss.Style {
	return lipgloss.NewStyle().
		Width(width).
		Align(lipgloss.Center).
		Foreground(theme.Current().Secondary).
		Padding(0, 0)
}

func helpViewStyle(width int) lipgloss.Style {
	return lipgloss.NewStyle().
		Width(width).
		Align(lipgloss.Center).
		Foreground(theme.Current().Secondary).
		MarginTop(1)
}
