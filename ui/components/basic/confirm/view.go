package confirm

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

const maxButtonGap = 20

func (m Model) View() string {

	noBtn, yesBtn := m.renderButtons()
	buttons := m.layoutButtons(noBtn, yesBtn)

	message := MessageStyle(lipgloss.Width(buttons)).Render(m.Message)

	return lipgloss.JoinVertical(
		lipgloss.Left,
		message,
		buttons,
	)
}

func (m Model) renderButtons() (string, string) {
	no := ButtonStyle().Render("No")
	yes := ButtonStyle().Render("Yes")

	if m.confirm {
		yes = SelectButtonStyle().Render("Yes")
	} else {
		no = SelectButtonStyle().Render("No")
	}

	return no, yes
}

func (m Model) layoutButtons(noBtn, yesBtn string) string {
	available := m.Layout.Terminal.Width -
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
