package notice

import (
	"bgscan/ui/shared/layout"
	"bgscan/ui/shared/overlay"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/google/uuid"
)

type LEVEL = int

const (
	NOTICE_ERROR LEVEL = iota
	NOTICE_INFO
	NOTICE_SUCCESS
)

type Model struct {
	id         string
	Layout     *layout.Layout
	NoticeType LEVEL
	Message    string
	Title      string
	viewport   viewport.Model

	titleHeight    int
	viewportHeight int
	footerHeight   int
}

func (m Model) Init() tea.Cmd {
	return m.viewport.Init()
}

func (m Model) ID() overlay.OverlayID {
	return m.id
}

func New(layout *layout.Layout, title, message string, level LEVEL) Model {
	v := viewport.New(0, 0)
	m := Model{
		id:         uuid.NewString(),
		Layout:     layout,
		NoticeType: level,
		Message:    message,
		Title:      title,
		viewport:   v,
	}

	wrapped := lipgloss.NewStyle().
		Width(m.Width()).
		Render(m.Message)

	m.viewport.SetContent(wrapped)
	return m.UpdateSize()
}

func (m *Model) Height() int {
	height := min(50, m.Layout.Body.Height)
	return height
}

func (m Model) UpdateSize() Model {
	m.titleHeight = lipgloss.Height(m.headerView())
	m.footerHeight = lipgloss.Height(m.footerView())
	m.viewportHeight = max(m.Height()-m.titleHeight-m.footerHeight, 1)
	wrappedMsgHeight := lipgloss.Height(lipgloss.NewStyle().Width(m.Width()).Render(m.Message))

	m.viewportHeight = min(wrappedMsgHeight, m.viewportHeight)

	m.viewport.Width = m.Width()
	m.viewport.Height = m.viewportHeight

	return m
}

func (m *Model) Width() int {
	width := min(50, m.Layout.Body.Width)
	return width
}

func (m Model) CloseCmd() tea.Cmd {
	return func() tea.Msg {
		return overlay.CloseOverlayMsg{ID: m.ID()}
	}
}
