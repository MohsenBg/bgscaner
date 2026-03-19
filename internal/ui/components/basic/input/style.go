package input

import (
	"bgscan/internal/ui/theme"

	"github.com/charmbracelet/lipgloss"
)

func containerStyle(width int) lipgloss.Style {
	return lipgloss.NewStyle().Width(width)
}

func messageStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(theme.Current().Text).
		Bold(true).
		MarginBottom(1)
}

func errorStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(theme.Current().Error).
		Bold(true).
		MarginTop(1)
}

func keyHintStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(theme.Current().Info).
		MarginTop(1)
}

