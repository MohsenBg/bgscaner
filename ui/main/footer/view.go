package footer

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

func (m Model) View() string {
	width := m.layout.Footer.Width
	height := m.layout.Footer.Height

	// Calculate section widths
	leftWidth := width / 3
	centerWidth := width / 3
	rightWidth := width - leftWidth - centerWidth

	// Left section: App info
	leftSection := leftSectionStyle(leftWidth).Render(
		fmt.Sprintf("%s %s %s",
			iconStyle().Render("⚡"),
			appNameStyle().Render("BGScan"),
			versionStyle().Render("v"+m.appVersion),
		),
	)

	// Center section: Status
	centerSection := centerSectionStyle(centerWidth).Render(
		statusTextStyle().Render(m.status),
	)

	// Right section: Time and extras
	currentTime := time.Now().Format("15:04:05")
	rightSection := rightSectionStyle(rightWidth).Render(
		fmt.Sprintf("%s %s | %s %s",
			iconStyle().Render("🕐"),
			timeStyle().Render(currentTime),
			iconStyle().Render("⌨"),
			helpStyle().Render("? help"),
		),
	)

	// Combine sections
	footerContent := lipgloss.JoinHorizontal(
		lipgloss.Top,
		leftSection,
		centerSection,
		rightSection,
	)

	// Add separator line
	separator := separatorStyle(width).Render(strings.Repeat("─", width))

	footer := lipgloss.JoinVertical(
		lipgloss.Left,
		separator,
		footerContent,
	)

	return containerStyle(width, height).Render(footer)
}
