package header

import (
	"bgscan/internal/ui/theme"

	"github.com/charmbracelet/lipgloss"
)

func bannerStyle(width, height int) lipgloss.Style {
	return lipgloss.NewStyle().
		Align(lipgloss.Center, lipgloss.Bottom).
		Width(width).Height(height).
		Foreground(theme.Current().Success).
		Bold(true)
}
