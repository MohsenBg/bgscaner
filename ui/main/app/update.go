package app

import (
	"bgscan/ui/components/basic/confirm"

	tea "github.com/charmbracelet/bubbletea"
)

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.layout.Update(msg.Width, msg.Height)
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "q":
			return m, confirm.ExitConfirmCmd(m.layout)
		}
	}

	var headerCmd tea.Cmd
	m.header, headerCmd = m.header.Update(msg)
	cmds = append(cmds, headerCmd)

	var bodyCmd tea.Cmd
	m.body, bodyCmd = m.body.Update(msg)
	cmds = append(cmds, bodyCmd)

	var footerCmd tea.Cmd
	m.footer, footerCmd = m.footer.Update(msg)
	cmds = append(cmds, footerCmd)

	return m, tea.Batch(cmds...)
}
