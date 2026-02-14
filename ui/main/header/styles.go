package header

import (
	"bgscan/ui/theme"

	"github.com/charmbracelet/lipgloss"
)

func bannerStyle(width, height int) lipgloss.Style {
	return lipgloss.NewStyle().
		Align(lipgloss.Center, lipgloss.Bottom).
		Width(width).Height(height).
		Foreground(theme.Current().Success).
		Bold(true)
}
