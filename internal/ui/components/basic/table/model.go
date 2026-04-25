package table

import (
	"bgscan/internal/ui/shared/env"
	"bgscan/internal/ui/shared/layout"
	"bgscan/internal/ui/shared/ui"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Column is an alias to bubbles/table Column for convenience.
type Column = table.Column

// Row is an alias to bubbles/table Row for convenience.
type Row = table.Row

// Model represents a reusable BubbleTea table component.
//
// It wraps the bubbles/table component while adding:
//
//   - integrated keymap action system
//   - help panel support
//   - layout‑aware resizing
//   - dynamic column scaling
//
// The component is designed to integrate with the BGScan UI framework.
type Model struct {
	id   ui.ComponentID
	name string

	// Title displayed above the table
	Title string

	// Layout reference for responsive sizing
	Layout *layout.Layout

	// Help model used to render key help
	Help help.Model

	// FullHelp toggles between short and full help modes
	FullHelp bool

	// Underlying bubbles table component
	BubbleTable table.Model

	// Key bindings
	Keys KeyMap

	// Original column widths used for resizing
	colsWidth []int

	paddingY int
}

// Init implements the BubbleTea initialization interface.
func (m *Model) Init() tea.Cmd {
	return nil
}

// ID returns the component unique identifier.
func (m *Model) ID() ui.ComponentID {
	return m.id
}

// Name returns the component name.
func (m *Model) Name() string {
	return m.name
}

// OnClose is called when the component is removed from the UI.
func (m *Model) OnClose() tea.Cmd {
	return nil
}

// New creates a new table component.
//
// Parameters:
//   - title: table title
//   - cols: table columns
//   - rows: initial rows
//   - layout: UI layout reference
func New(title string, cols []table.Column, rows []table.Row, layout *layout.Layout) *Model {
	m := &Model{
		id:       ui.NewComponentID(),
		name:     "table",
		Title:    title,
		Layout:   layout,
		Help:     help.New(),
		Keys:     defaultKeys(),
		paddingY: 10,
	}

	m.BubbleTable = m.createTable(rows, cols)
	m.updateTableSize()

	return m
}

func (m *Model) SetPaddingY(padding int) {
	m.paddingY = padding
	m.updateTableSize()
}

//
// ----------------------
// ActionKey
// ----------------------
//

// ActionKey defines a keyboard action bound to one or more keys.
type ActionKey struct {
	Keys      []string
	ShortHelp string
	FullHelp  string
	Cmd       tea.Cmd
}

// NewKey creates a new ActionKey definition.
func NewKey(keys []string, shortHelp, fullHelp string, cmd tea.Cmd) ActionKey {
	ks := make([]string, len(keys))

	for i, key := range keys {
		switch key {
		case "up", "↑":
			ks[i] = "↑"
		case "down", "↓":
			ks[i] = "↓"
		case "left", "←":
			ks[i] = "←"
		case "right", "→":
			ks[i] = "→"
		default:
			ks[i] = key
		}
	}

	// Join nicely (e.g., "↑/↓")
	kstr := strings.Join(ks, "/")

	if kstr == "" {
		kstr = "?"
	}

	if shortHelp != "" {
		shortHelp = fmt.Sprintf("%s %s", kstr, shortHelp)
	}

	if fullHelp != "" {
		fullHelp = fmt.Sprintf("%s %s", kstr, fullHelp)
	}

	return ActionKey{
		Keys:      keys,
		ShortHelp: shortHelp,
		FullHelp:  fullHelp,
		Cmd:       cmd,
	}
}

func (m *Model) SetKeys(keys ...ActionKey) {
	m.Keys = defaultKeys(keys...)
}

//
// ----------------------
// KeyMap
// ----------------------
//

// KeyMap represents a collection of keyboard actions.
type KeyMap struct {
	Actions []ActionKey
}

// Add registers a new keyboard action.
func (k *KeyMap) Add(a ActionKey) {
	k.Actions = append(k.Actions, a)
}

// Check checks if a key press matches a registered action and
// returns its command if found.
func (k KeyMap) Check(msg tea.Msg) tea.Cmd {
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return nil
	}

	keyStr := keyMsg.String()

	for _, a := range k.Actions {
		if slices.Contains(a.Keys, keyStr) {
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

// ShortHelp implements help.KeyMap and returns condensed help bindings.
func (k KeyMap) ShortHelp() []key.Binding {
	bindings := make([]key.Binding, 0, len(k.Actions))

	for _, a := range k.Actions {
		if a.ShortHelp == "" {
			continue
		}

		bindings = append(bindings, key.NewBinding(
			key.WithKeys(a.Keys...),
			key.WithHelp(a.ShortHelp, ""),
		))
	}

	return bindings
}

// FullHelp implements help.KeyMap and returns expanded help bindings.
func (k KeyMap) FullHelp() [][]key.Binding {
	if len(k.Actions) == 0 {
		return nil
	}

	colCount := 4
	cols := make([][]key.Binding, colCount)

	for i, a := range k.Actions {
		if a.FullHelp == "" {
			continue
		}

		binding := key.NewBinding(
			key.WithKeys(a.Keys...),
			key.WithHelp("", a.FullHelp),
		)

		col := i % colCount
		cols[col] = append(cols[col], binding)
	}

	return cols
}

//
// ----------------------
// Default Keys
// ----------------------
//

// defaultKeys returns the default key bindings used by the table.
func defaultKeys(keys ...ActionKey) KeyMap {
	const spacebar = " "

	km := KeyMap{}

	km.Add(NewKey(
		[]string{"up", "k"},
		"up",
		"Move up",
		nil,
	))

	km.Add(NewKey(
		[]string{"down", "j"},
		"down",
		"Move down",
		nil,
	))

	km.Add(NewKey(
		[]string{"b", "pgup"},
		"",
		"Page up",
		nil,
	))

	km.Add(NewKey(
		[]string{"f", "pgdown", spacebar},
		"",
		"Page down",
		nil,
	))

	km.Add(NewKey(
		[]string{"u", "ctrl+u"},
		"",
		"½ page up",
		nil,
	))

	km.Add(NewKey(
		[]string{"d", "ctrl+d"},
		"",
		"½ page down",
		nil,
	))

	km.Add(NewKey(
		[]string{"home", "g"},
		"",
		"Go to start",
		nil,
	))

	km.Add(NewKey(
		[]string{"end", "G"},
		"",
		"Go to end",
		nil,
	))

	for _, key := range keys {
		km.Add(key)
	}

	km.Add(NewKey(
		[]string{"?"},
		"help",
		"Toggle help",
		nil,
	))

	km.Add(NewKey(
		[]string{"q", "esc"},
		"quit",
		"Quit",
		nil,
	))

	return km
}

//
// ----------------------
// Row Helpers
// ----------------------
//

// NewRowTime formats a time value for table display.
// Zero values are rendered as "-".
func NewRowTime(t time.Time) string {
	if t.IsZero() {
		return "-"
	}
	return t.Format("2006-01-02 15:04:05")
}

// NewRowBool formats a boolean value as "yes" or "no".
func NewRowBool(v bool) string {
	if v {
		return "yes"
	}
	return "no"
}

// NewTimeDurationRow returns a human-readable duration
// since the given time.
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
// Internal Helpers
// ----------------------
//

func (m *Model) createTable(rows []table.Row, cols []table.Column) table.Model {
	width := m.tableWidth()

	total := 0
	for _, col := range cols {
		total += col.Width
	}

	if total > 0 {
		ratio := float64(width) / float64(total)

		m.colsWidth = make([]int, len(cols))

		for i := range cols {
			m.colsWidth[i] = cols[i].Width
			cols[i].Width = int(float64(cols[i].Width) * ratio)
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

func (m *Model) updateTableSize() {
	if m.Layout == nil {
		return
	}

	width := m.tableWidth()

	helpHeight := lipgloss.Height(m.renderHelpView())
	titleHeight := lipgloss.Height(m.renderTitle())
	padding := m.paddingY

	height := max(1, m.Layout.Content.Height-helpHeight-titleHeight-padding)

	cols := m.BubbleTable.Columns()
	if len(cols) == 0 {
		return
	}

	total := 0
	for _, w := range m.colsWidth {
		total += w
	}

	if total <= 0 {
		return
	}

	ratio := float64(width) / float64(total)

	for i := range cols {
		cols[i].Width = int(ratio * float64(m.colsWidth[i]))
	}

	m.BubbleTable.SetColumns(cols)
	m.BubbleTable.SetHeight(height)
}

func (m *Model) tableWidth() int {
	if m.Layout == nil {
		return 80
	}

	return min(80, m.Layout.Body.Width-10)
}

// AppendRow appends a new row to the table.
func (m *Model) AppendRow(row table.Row) {
	rows := append([]table.Row(nil), m.BubbleTable.Rows()...)
	rows = append(rows, row)
	m.BubbleTable.SetRows(rows)
}

// Mode returns the environment mode required by the component.
func (m *Model) Mode() env.Mode {
	return env.NormalMode
}
