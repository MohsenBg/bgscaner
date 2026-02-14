package table

import (
	"bgscan/ui/shared/layout"
	"slices"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
)

type Column = table.Column
type Row = table.Row

// Model represents a table with help and key actions
type Model struct {
	Title       string
	Layout      *layout.Layout
	Help        help.Model
	FullHelp    bool
	BubbleTable table.Model
	Keys        KeyMap
	colsWidth   []int
}

func (m Model) Init() tea.Cmd {
	return nil
}

// New creates a new table model
func New(title string, cols []table.Column, rows []table.Row, layout *layout.Layout) Model {
	m := Model{
		Title:  title,
		Layout: layout,
		Help:   help.New(),
		Keys:   DefaultKeys(),
	}
	m.BubbleTable = m.createTable(rows, cols)
	return m
}

//
// ----------------------
// ActionKey
// ----------------------
//

type ActionKey struct {
	Keys      []string
	ShortHelp string
	FullHelp  string
	Cmd       tea.Cmd
}

func NewActionKey(keys []string, shortHelp, fullHelp string, cmd tea.Cmd) ActionKey {
	return ActionKey{
		Keys:      keys,
		ShortHelp: shortHelp,
		FullHelp:  fullHelp,
		Cmd:       cmd,
	}
}

//
// ----------------------
// KeyMap
// ----------------------
//

type KeyMap struct {
	Actions []ActionKey
}

func (k *KeyMap) Add(a ActionKey) {
	k.Actions = append(k.Actions, a)
}

func (k KeyMap) Check(msg tea.Msg) tea.Cmd {
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return nil
	}

	for _, a := range k.Actions {
		if slices.Contains(a.Keys, keyMsg.String()) {
			return a.Cmd
		}
	}

	return nil
}

//
// ----------------------
// Bubble Tea help integration
// ----------------------
//

func (k KeyMap) ShortHelp() []key.Binding {
	bindings := make([]key.Binding, 0, len(k.Actions))

	for _, a := range k.Actions {
		// Fallback short help to first key if empty
		short := a.ShortHelp
		if short == "" {
			continue
		}

		bindings = append(bindings, key.NewBinding(
			key.WithKeys(a.Keys...),
			key.WithHelp(short, ""),
		))
	}

	return bindings
}

func (k KeyMap) FullHelp() [][]key.Binding {
	if len(k.Actions) == 0 {
		return nil
	}

	colCount := 4
	cols := make([][]key.Binding, colCount)

	for i, a := range k.Actions {
		full := a.FullHelp
		if full == "" {
			continue
		}

		binding := key.NewBinding(
			key.WithKeys(a.Keys...),
			key.WithHelp("", full),
		)

		col := i % colCount
		cols[col] = append(cols[col], binding)
	}

	return cols
}

//
// ----------------------
// Defaults
// ----------------------
//

func DefaultKeys() KeyMap {
	const spacebar = " "

	km := KeyMap{}

	km.Add(NewActionKey(
		[]string{"up", "k"},
		"↑/k up",
		"↑/k Move up",
		nil,
	))

	km.Add(NewActionKey(
		[]string{"down", "j"},
		"↓/j down",
		"↓/j Move down",
		nil,
	))

	km.Add(NewActionKey(
		[]string{"b", "pgup"},
		"",
		"b/pgup Page up",
		nil,
	))

	km.Add(NewActionKey(
		[]string{"f", "pgdown", spacebar},
		"",
		"f/pgdn Page down",
		nil,
	))

	km.Add(NewActionKey(
		[]string{"u", "ctrl+u"},
		"",
		"u ½ page up",
		nil,
	))

	km.Add(NewActionKey(
		[]string{"d", "ctrl+d"},
		"",
		"d ½ page down",
		nil,
	))

	km.Add(NewActionKey(
		[]string{"home", "g"},
		"",
		"g/home Go to start",
		nil,
	))

	km.Add(NewActionKey(
		[]string{"end", "G"},
		"",
		"G/end Go to end",
		nil,
	))

	km.Add(NewActionKey(
		[]string{"?"},
		"? more",
		"? Toggle help",
		nil,
	))

	km.Add(NewActionKey(
		[]string{"q", "esc"},
		"q quit",
		"q Quit",
		nil,
	))

	return km
}

//
// ----------------------
// Table helpers
// ----------------------
//

func NewRowTime(t time.Time) string {
	if t.IsZero() {
		return "-"
	}
	return t.Format("2006-01-02 15:04:05")
}

func NewRowBool(v bool) string {
	if v {
		return "yes"
	}
	return "no"
}

func NewTimeDurationRow(from time.Time) string {
	if from.IsZero() {
		return "-"
	}

	d := time.Since(from)

	switch {
	case d < time.Second:
		return "just now"
	case d < time.Minute:
		return d.Truncate(time.Second).String()
	case d < time.Hour:
		return d.Truncate(time.Minute).String()
	default:
		return d.Truncate(time.Hour).String()
	}
}

//
// ----------------------
// Internal helpers
// ----------------------
//

func (m *Model) createTable(rows []table.Row, cols []table.Column) table.Model {
	width := m.tableWidth()
	total := 0
	for _, col := range cols {
		total += col.Width
	}

	if total > 0 {
		ratio := width / total
		for i := range cols {
			m.colsWidth = append(m.colsWidth, cols[i].Width)
			cols[i].Width *= ratio
		}
	}

	t := table.New(
		table.WithColumns(cols),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(len(rows)),
	)

	t.SetStyles(tableStyles())
	return t
}

func (m Model) updateTableSize() Model {
	width := m.tableWidth()
	height := m.Layout.Content.Height - 20
	cols := m.BubbleTable.Columns()
	if len(cols) == 0 {
		return m
	}

	total := 0
	for _, w := range m.colsWidth {
		total += w
	}

	if total <= 0 {
		return m
	}

	var ratio float64 = float64(width) / float64(total)
	for i := range cols {
		cols[i].Width = int(ratio * float64(m.colsWidth[i]))
	}

	m.BubbleTable.SetColumns(cols)
	m.BubbleTable.SetHeight(height)
	return m
}

func (m Model) tableWidth() int {
	return min(80, m.Layout.Body.Width-10)
}

func (m *Model) AppendRow(row table.Row) {
	rows := m.BubbleTable.Rows()
	rows = append(rows, row)
	m.BubbleTable.SetRows(rows)
}
