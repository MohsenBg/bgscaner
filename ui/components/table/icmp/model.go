package icmpconfig

import (
	"bgscan/core/config"
	"bgscan/ui/shared/layout"
	"fmt"
	"strconv"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ═══════════════════════════════════════════════════════════
// Constants & Config Descriptions
// ═══════════════════════════════════════════════════════════

var descriptions = map[string]string{
	"Workers":       "Number of concurrent ping workers",
	"Timeout":       "Maximum wait time for ICMP response",
	"Shuffle IPs":   "Randomize IP test order to avoid patterns",
	"Output Prefix": "File name prefix for results storage",
}

// ═══════════════════════════════════════════════════════════
// Key Bindings
// ═══════════════════════════════════════════════════════════

type keyMap struct {
	Up    key.Binding
	Down  key.Binding
	Enter key.Binding
	Back  key.Binding
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Up, k.Down, k.Enter, k.Back}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Enter, k.Back},
	}
}

var keys = keyMap{
	Up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "move up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "move down"),
	),
	Enter: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "edit field"),
	),
	Back: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "back"),
	),
}

// ═══════════════════════════════════════════════════════════
// Model
// ═══════════════════════════════════════════════════════════

type Model struct {
	table  table.Model
	layout *layout.Layout
	help   help.Model
	keys   keyMap
}

// ═══════════════════════════════════════════════════════════
// Constructor
// ═══════════════════════════════════════════════════════════

func New(layout *layout.Layout) Model {
	m := Model{
		layout: layout,
		help:   help.New(),
		keys:   keys,
	}

	// Initialize table with current config
	m.table = m.createTable()
	m.help.Width = layout.BodyContentWidth()

	return m
}

func (m Model) Init() tea.Cmd {
	return nil
}

// ═══════════════════════════════════════════════════════════
// Helper Methods
// ═══════════════════════════════════════════════════════════

func (m *Model) createTable() table.Model {
	width := m.calculateTableWidth()
	columns := m.calculateColumns(width)
	rows := m.buildRows(columns)

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(len(rows)),
	)

	t.SetStyles(m.tableStyles())
	return t
}

func (m *Model) calculateTableWidth() int {
	return min(80, m.layout.BodyContentWidth()-20)
}

func (m *Model) calculateColumns(totalWidth int) []table.Column {
	return []table.Column{
		{Title: "Setting", Width: int(float64(totalWidth) * 0.25)},
		{Title: "Value", Width: int(float64(totalWidth) * 0.20)},
		{Title: "Description", Width: int(float64(totalWidth) * 0.55)},
	}
}

func (m *Model) buildRows(columns []table.Column) []table.Row {
	icmpConfig := config.GetICMP()

	settingWidth := columns[0].Width
	valueWidth := columns[1].Width
	descWidth := columns[2].Width

	return []table.Row{
		{
			padCell("Workers", settingWidth),
			padCell(strconv.Itoa(icmpConfig.Workers), valueWidth),
			padCell(descriptions["Workers"], descWidth),
		},
		{
			padCell("Timeout", settingWidth),
			padCell(formatDuration(icmpConfig.Timeout), valueWidth),
			padCell(descriptions["Timeout"], descWidth),
		},
		{
			padCell("Shuffle IPs", settingWidth),
			padCell(formatBool(icmpConfig.ShuffleIPs), valueWidth),
			padCell(descriptions["Shuffle IPs"], descWidth),
		},
		{
			padCell("Output Prefix", settingWidth),
			padCell(icmpConfig.PrefixOutput, valueWidth),
			padCell(descriptions["Output Prefix"], descWidth),
		},
	}
}

func (m *Model) tableStyles() table.Styles {
	s := table.DefaultStyles()

	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(true).
		Padding(0, 1)

	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false).
		Padding(0, 1)

	s.Cell = s.Cell.
		Padding(0, 1)

	return s
}

func (m *Model) updateTableSize() {
	width := m.calculateTableWidth()
	columns := m.calculateColumns(width)
	rows := m.buildRows(columns)

	m.table.SetColumns(columns)
	m.table.SetRows(rows)
}

// ═══════════════════════════════════════════════════════════
// Formatting Helpers
// ═══════════════════════════════════════════════════════════

func padCell(content string, width int) string {
	return lipgloss.NewStyle().
		Width(width).
		Padding(0, 1).
		Render(content)
}

func formatDuration(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}
	return fmt.Sprintf("%ds", int(d.Seconds()))
}

func formatBool(b bool) string {
	if b {
		return "✓ Yes"
	}
	return "✗ No"
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
