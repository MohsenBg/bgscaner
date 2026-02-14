package body

import (
	"bgscan/ui/theme"

	"github.com/charmbracelet/lipgloss"
)

func containerStyle(width, height int) lipgloss.Style {
	return lipgloss.NewStyle().Align(lipgloss.Center, lipgloss.Center).
		Width(width).Height(height)
}

func WindowStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.Current().BorderActive).
		Padding(1, 2)
}
