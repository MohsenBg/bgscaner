package app

import (
	"bgscan/ui/main/body"
	"bgscan/ui/main/footer"
	"bgscan/ui/main/header"
	"bgscan/ui/shared/layout"

	tea "github.com/charmbracelet/bubbletea"
)

type model struct {
	layout *layout.Layout
	header header.Model
	body   body.Model
	footer footer.Model
}

func New() tea.Model {
	l := layout.New()
	return model{
		layout: l,
		header: header.New(l),
		body:   body.New(l),
		footer: footer.New(l),
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		m.header.Init(),
		m.body.Init(),
		m.footer.Init(),
	)
}
