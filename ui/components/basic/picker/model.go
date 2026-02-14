package picker

import (
	"bgscan/ui/shared/layout"
	"bgscan/ui/shared/overlay"
	"os"

	"github.com/charmbracelet/bubbles/filepicker"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/google/uuid"
)

// Model implements a file picker overlay
type Model struct {
	id         overlay.OverlayID
	Title      string
	Layout     *layout.Layout
	FilePicker filepicker.Model
	OnSelect   OnSelect
}

func arrowKeyCmd(k tea.KeyType) tea.Cmd {
	return func() tea.Msg {
		return tea.KeyMsg{Type: k}
	}
}

func (m Model) Init() tea.Cmd {
	return m.FilePicker.Init()
}

// ID returns the overlay ID
func (m Model) ID() overlay.OverlayID {
	return m.id
}

// CloseCmd returns a command to close this overlay
func (m Model) CloseCmd() tea.Cmd {
	return func() tea.Msg {
		return overlay.CloseOverlayMsg{ID: m.ID()}
	}
}

// New creates a new file picker overlay
func New(layout *layout.Layout, title string, baseDir string, allowType []string, onSelect OnSelect) Model {
	p := filepicker.New()
	if baseDir != "" {
		p.CurrentDirectory = baseDir
	} else {
		p.CurrentDirectory, _ = os.UserHomeDir()
	}

	if len(allowType) > 0 {
		p.AllowedTypes = allowType
	}

	if onSelect == nil {
		onSelect = func(path string) tea.Cmd { return nil }
	}

	p.ShowPermissions = true
	p.AutoHeight = false
	p.SetHeight(pickerHeight(layout))

	return Model{
		id:         overlay.OverlayID(uuid.NewString()),
		Title:      title,
		Layout:     layout,
		FilePicker: p,
		OnSelect:   onSelect,
	}
}
