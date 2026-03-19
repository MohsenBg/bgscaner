package confirm

import (
	"bgscan/internal/ui/shared/env"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

const maxButtonGap = 20

// View renders the confirmation dialog.
//
// The view consists of:
//   - The confirmation message
//   - Two action buttons: "No" and "Yes"
//
// The currently selected option is highlighted. Buttons are spaced
// dynamically based on the available terminal width to keep the dialog
// visually balanced.
func (m *Model) View() string {
	noBtn, yesBtn := m.renderButtons()
	buttons := m.layoutButtons(noBtn, yesBtn)

	message := MessageStyle(lipgloss.Width(buttons)).Render(m.message)

	return lipgloss.JoinVertical(
		lipgloss.Left,
		message,
		buttons,
	)
}

// renderButtons builds the styled "No" and "Yes" buttons.
//
// The button corresponding to the current selection state is rendered
// using SelectButtonStyle while the other uses the normal ButtonStyle.
func (m *Model) renderButtons() (string, string) {
	no := ButtonStyle().Render("No")
	yes := ButtonStyle().Render("Yes")

	if m.confirm {
		yes = SelectButtonStyle().Render("Yes")
	} else {
		no = SelectButtonStyle().Render("No")
	}

	return no, yes
}

// layoutButtons arranges the confirmation buttons horizontally.
//
// The spacing between buttons is dynamically calculated based on the
// terminal width while being clamped between 1 and maxButtonGap to
// prevent excessive spacing on large terminals.
func (m *Model) layoutButtons(noBtn, yesBtn string) string {
	width := 80
	if m.layout != nil && m.layout.Terminal.Width > 0 {
		width = m.layout.Terminal.Width
	}

	available := width -
		lipgloss.Width(noBtn) -
		lipgloss.Width(yesBtn)

	gap := min(max(available, 1), maxButtonGap)

	return lipgloss.JoinHorizontal(
		lipgloss.Top,
		noBtn,
		strings.Repeat(" ", gap),
		yesBtn,
	)
}

// Mode returns the UI mode in which the confirmation dialog operates.
//
// Confirmation dialogs run in NormalMode because they handle their
// own keyboard interactions and do not require special input modes.
func (m *Model) Mode() env.Mode {
	return env.NormalMode
}
