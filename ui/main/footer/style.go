package footer

import (
	"bgscan/ui/theme"

	"github.com/charmbracelet/lipgloss"
)

var (
	// ----------------------------------------
	// Container
	// ----------------------------------------

	containerStyle = func(width, height int) lipgloss.Style {
		t := theme.Current()

		return lipgloss.NewStyle().
			Width(width).
			Height(height).
			Foreground(t.Text)
	}

	// ----------------------------------------
	// Separator
	// ----------------------------------------

	separatorStyle = func(width int) lipgloss.Style {
		t := theme.Current()

		return lipgloss.NewStyle().
			Width(width).
			Foreground(t.Border)
	}

	// ----------------------------------------
	// Sections
	// ----------------------------------------

	leftSectionStyle = func(width int) lipgloss.Style {
		return lipgloss.NewStyle().
			Width(width).
			Padding(0, 1).
			Align(lipgloss.Left)
	}

	centerSectionStyle = func(width int) lipgloss.Style {
		return lipgloss.NewStyle().
			Width(width).
			Padding(0, 1).
			Align(lipgloss.Center)
	}

	rightSectionStyle = func(width int) lipgloss.Style {
		return lipgloss.NewStyle().
			Width(width).
			Padding(0, 1).
			Align(lipgloss.Right)
	}

	// ----------------------------------------
	// Text styles
	// ----------------------------------------

	appNameStyle = func() lipgloss.Style {
		t := theme.Current()

		return lipgloss.NewStyle().
			Foreground(t.AccentYellow).
			Bold(true)
	}

	versionStyle = func() lipgloss.Style {
		t := theme.Current()

		return lipgloss.NewStyle().
			Foreground(t.Success).
			Faint(true)
	}

	statusTextStyle = func() lipgloss.Style {
		t := theme.Current()

		return lipgloss.NewStyle().
			Foreground(t.Primary).
			Bold(true)
	}

	timeStyle = func() lipgloss.Style {
		t := theme.Current()

		return lipgloss.NewStyle().
			Foreground(t.Secondary)
	}

	helpStyle = func() lipgloss.Style {
		t := theme.Current()

		return lipgloss.NewStyle().
			Foreground(t.Muted).
			Italic(true)
	}

	iconStyle = func() lipgloss.Style {
		t := theme.Current()

		return lipgloss.NewStyle().
			Foreground(t.AccentOrange)
	}
)
