package menu

import (
	"bgscan/ui/theme"

	"github.com/charmbracelet/lipgloss"
)

func itemTitleStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(theme.Current().Text).
		Padding(0, 0, 0, 0)
}

func selectedItemTitleStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(theme.Current().Primary).
		Padding(0, 1).
		Bold(true)
}

func shortcutStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(theme.Current().Text)
}

func iconStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(theme.Current().Text).
		Width(4).
		Bold(true)
}

func selectedIconStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(theme.Current().Primary).
		Width(4).
		Bold(true)
}

func titleStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Align(lipgloss.Center).
		Foreground(theme.Current().Info).
		Bold(true)
}

func PaddingCell() lipgloss.Style {
	return lipgloss.NewStyle().Padding(1, 0, 0, 0)
}
