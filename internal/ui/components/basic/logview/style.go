package logview

import (
	"bgscan/internal/ui/theme"

	"github.com/charmbracelet/lipgloss"
)

func TitleStyle(width int) lipgloss.Style {
	return lipgloss.NewStyle().Width(width).
		Align(lipgloss.Center, lipgloss.Center).
		Bold(true).Border(lipgloss.NormalBorder(), false, false, true, false).
		BorderForeground(theme.Current().BorderActive)
}

func ContainerStyle(width int) lipgloss.Style {
	return lipgloss.NewStyle().Width(width).Align(lipgloss.Center, lipgloss.Center)
}

func BorderStyle(width int) lipgloss.Style {
	width = min(80, width)
	return lipgloss.NewStyle().Width(width).
		Align(lipgloss.Center, lipgloss.Center).
		Border(lipgloss.DoubleBorder()).BorderForeground(theme.Current().BorderActive)
}

func helpStyle(width int) lipgloss.Style {
	return lipgloss.NewStyle().Width(width-5).
		Foreground(theme.Current().Muted).Align(lipgloss.Center, lipgloss.Center).
		Padding(1)
}

func helpKeyStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(theme.Current().Secondary).
		Bold(true)
}
