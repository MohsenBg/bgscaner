package app

import (
	"bgscan/ui/theme"

	"github.com/charmbracelet/lipgloss"
)

func containerStyle(termWidth, termHeight int) lipgloss.Style {
	return lipgloss.NewStyle().
		Align(lipgloss.Center, lipgloss.Center).
		Width(termWidth).
		Height(termHeight)
}

func mainStyle(contentWidth, contentHeight int) lipgloss.Style {
	return lipgloss.NewStyle().Align(lipgloss.Center, lipgloss.Center).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.Current().BorderActive).
		Width(contentWidth).
		Height(contentHeight)
}

func warningStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(theme.Current().AccentYellow).
		Bold(true).
		Padding(1, 2)
}

func iconStyle() lipgloss.Style {
	return lipgloss.NewStyle().Width(10)
}

func centerStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Align(lipgloss.Center, lipgloss.Center)
}
