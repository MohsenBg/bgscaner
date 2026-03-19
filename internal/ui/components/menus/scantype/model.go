package scantype

import (
	"bgscan/internal/core/scanner"
	"bgscan/internal/ui/components/basic/menu"
	"bgscan/internal/ui/components/basic/notice"
	scannerUi "bgscan/internal/ui/components/scanner"
	"bgscan/internal/ui/shared/env"
	"bgscan/internal/ui/shared/layout"
	"bgscan/internal/ui/shared/ui"
	"context"

	tea "github.com/charmbracelet/bubbletea"
)

type Model struct {
	id      ui.ComponentID
	name    string
	layout  *layout.Layout
	input   string
	overlay ui.Component
	menu    ui.Component
}

func (m *Model) Init() tea.Cmd { return nil }

func (m *Model) ID() ui.ComponentID { return m.id }

func (m *Model) Name() string { return m.name }

func (m *Model) OnClose() tea.Cmd { return nil }

func New(layout *layout.Layout, input string) *Model {
	m := &Model{
		id:     ui.NewComponentID(),
		name:   "Scan Menu",
		layout: layout,
		input:  input,
	}

	items := []menu.MenuItem{
		menu.NewMenuItem(
			"▦",
			"ICMP Scan",
			"i",
			m.openScanner(scanner.ICMP_SCAN, input),
		),
		menu.NewMenuItem(
			"≡",
			"TCP Scan",
			"t",
			m.openScanner(scanner.TCP_SCAN, input),
		),
		menu.NewMenuItem(
			"▦",
			"HTTP Scan",
			"h",
			m.openScanner(scanner.HTTP_SCAN, input),
		),
	}

	m.menu = menu.New(items, "Select Scan Type", layout)
	return m
}

func (m *Model) Mode() env.Mode {
	return env.NormalMode
}

// openScanner is the shared implementation for both TCP and ICMP scanners.
// It closes the overlay, creates the correct scanner type, and opens the scanner UI.
func (m *Model) openScanner(mode scanner.ScanMode, input string) tea.Cmd {
	cmds := make([]tea.Cmd, 0, 2)

	// close the overlay first regardless of scanner outcome
	if m.overlay != nil {
		id := m.overlay.ID()
		m.overlay = nil
		cmds = append(cmds, func() tea.Msg {
			return ui.CloseComponentMsg{ID: id}
		})
	}

	s, err := newScanner(mode, input)
	if err != nil {
		cmds = append(cmds, m.errorCmd("Error Creating Scanner", err.Error()))
		return tea.Batch(cmds...)
	}

	cmds = append(cmds,
		ui.OpenComponentCmd(
			scannerUi.New(m.layout, 10_000, s),
		),
	)

	return tea.Batch(cmds...)
}

// newScanner constructs the correct scanner based on mode.
func newScanner(mode scanner.ScanMode, input string) (scanner.Scanner, error) {
	ctx := context.Background()
	switch mode {
	case scanner.TCP_SCAN:
		return scanner.NewTCPScanner(ctx, input)
	case scanner.ICMP_SCAN:
		return scanner.NewICMPScanner(ctx, input)
	case scanner.HTTP_SCAN:
		return scanner.NewHTTPScanner(ctx, input)
	default:
		return scanner.NewTCPScanner(ctx, input)
	}
}

// errorCmd returns a notice command styled as an error.
func (m *Model) errorCmd(title, message string) tea.Cmd {
	return notice.NewNoticeCmd(m.layout, title, message, notice.NOTICE_ERROR)
}
