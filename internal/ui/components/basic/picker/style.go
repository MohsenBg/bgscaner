package picker

import (
	"bgscan/internal/ui/shared/layout"
	"bgscan/internal/ui/theme"

	"github.com/charmbracelet/lipgloss"
)

// -------------------- Styling --------------------
func containerStyle(width, height int) lipgloss.Style {
	return lipgloss.NewStyle().
		Align(lipgloss.Left, lipgloss.Top).
		Width(width).
		Height(height).
		Padding(0, 1).
		Margin(1, 0)
}

func currentDirStyle(width int) lipgloss.Style {
	return lipgloss.NewStyle().
		Align(lipgloss.Left).
		Width(width-5).Border(lipgloss.NormalBorder(), false, false, true, false).
		Bold(true).
		Foreground(theme.Current().Yellow)
}

// pickerHeight computes the height for the overlay
func pickerHeight(layout *layout.Layout) int {
	return layout.Body.Height - 10
}

func TitleStyle(width int) lipgloss.Style {
	return lipgloss.NewStyle().
		Width(width-5).
		Align(lipgloss.Center).
		Bold(true).
		Foreground(theme.Current().Info).
		Padding(0, 0, 2, 0).
		BorderForeground(lipgloss.Color("240"))
}

func helpStyle(width int) lipgloss.Style {
	return lipgloss.NewStyle().Width(width - 5).
		Foreground(theme.Current().Muted).
		Align(lipgloss.Center).
		PaddingTop(1)
}

func helpKeyStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(theme.Current().Secondary).
		Bold(true)
}
