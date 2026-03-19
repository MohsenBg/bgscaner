package body

import (
	"github.com/charmbracelet/lipgloss"
)

func containerStyle(width, height int) lipgloss.Style {
	return lipgloss.NewStyle().Align(lipgloss.Center, lipgloss.Center).
		Width(width).Height(height)
}
