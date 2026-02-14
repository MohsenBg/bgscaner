package entry

import (
	"bgscan/ui/shared/nav"

	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case nav.OpenViewMsg:
		m.openView(msg.View)
	}

	return m.updateCurrentView(msg)
}

func (m *Model) openView(view nav.ViewName) {
	if v, exists := m.Views[view]; exists {
		m.CurrentViewName = v.Name
	}
}

func (m Model) updateCurrentView(msg tea.Msg) (tea.Model, tea.Cmd) {
	currentView := m.CurrentView()
	if currentView == nil {
		return m, nil
	}

	var cmd tea.Cmd
	model, cmd := currentView.Update(msg)
	m.Views[m.CurrentViewName] = nav.View{Name: m.CurrentViewName, Model: model}
	return m, cmd
}
