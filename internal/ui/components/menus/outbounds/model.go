package outbounds

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"bgscan/internal/core/xray"
	"bgscan/internal/logger"
	"bgscan/internal/ui/components/basic/confirm"
	"bgscan/internal/ui/components/basic/input"
	"bgscan/internal/ui/components/basic/notice"
	"bgscan/internal/ui/components/basic/picker"
	"bgscan/internal/ui/components/basic/table"
	"bgscan/internal/ui/shared/env"
	"bgscan/internal/ui/shared/layout"
	"bgscan/internal/ui/shared/ui"
	"bgscan/internal/ui/shared/validation"
)

// Model represents the Outbound Templates view.
//
// The component displays outbound template files in a table and provides
// operations such as:
//
//   - Adding new outbound template
//   - Deleting outbound template
//   - Renaming outbound template
//   - Selecting outbound template for further actions
//
// A map is internally maintained for fast lookup.
type Model struct {
	id           ui.ComponentID
	name         string
	layout       *layout.Layout
	table        ui.Component
	columns      []table.Column
	rows         []table.Row
	outbounds    []xray.XrayOutboundsFile
	outboundsMap map[string]*xray.XrayOutboundsFile
	onSelect     func(*xray.XrayOutboundsFile) tea.Cmd
}

// Default table columns for listing outbound templates.
var outboundColumns = []table.Column{
	{Title: "Name", Width: 50},
	{Title: "Created Time", Width: 50},
}

// New creates a new outbound template list component.
func New(l *layout.Layout, title string, onSelect func(*xray.XrayOutboundsFile) tea.Cmd) *Model {
	m := &Model{
		layout:       l,
		id:           ui.NewComponentID(),
		name:         "outbounds",
		columns:      outboundColumns,
		rows:         []table.Row{},
		outboundsMap: make(map[string]*xray.XrayOutboundsFile),
		onSelect:     onSelect,
	}

	t := table.New(title, outboundColumns, m.rows, l)
	keys := make([]table.ActionKey, 0, 4)

	// Select
	if onSelect != nil {
		keys = append(keys,
			table.NewKey(
				[]string{tea.KeyEnter.String()},
				"select",
				"select outbound",
				func() tea.Msg { return SelectMsg{} },
			))
	}

	// Add
	keys = append(keys,
		table.NewKey(
			[]string{"a"},
			"add",
			"add outbound template",
			picker.NewOpenPickFileCmd(
				l,
				"Select outbound .json",
				"",
				[]string{".json"},
				m.handleFileSelect,
			),
		),

		// Delete
		table.NewKey(
			[]string{"x"},
			"remove",
			"remove outbound",
			func() tea.Msg { return RequestDeleteOutboundMsg{} },
		),

		// Rename
		table.NewKey(
			[]string{"r"},
			"rename",
			"rename outbound",
			func() tea.Msg { return RequestRenameOutboundMsg{} },
		),
	)

	t.SetKeys(keys...)
	m.table = t
	return m
}

// Init loads outbound templates from disk.
func (m *Model) Init() tea.Cmd {
	return loadOutboundsTemplatesCmd(m.layout)
}

func (m *Model) ID() ui.ComponentID { return m.id }
func (m *Model) Name() string       { return m.name }
func (m *Model) OnClose() tea.Cmd   { return nil }
func (m *Model) Mode() env.Mode     { return env.NormalMode }

// loadOutboundsTemplatesCmd loads outbound templates.
func loadOutboundsTemplatesCmd(l *layout.Layout) tea.Cmd {
	return func() tea.Msg {
		outbounds, err := xray.GetOutboundsTemplates()
		if err != nil {
			logger.UIError("Failed to load outbounds: %s", err.Error())
			return notice.NewNoticeCmd(
				l,
				"Error Loading Outbounds",
				fmt.Sprintf("Failed to load outbound templates: %v", err),
				notice.NOTICE_ERROR,
			)
		}

		logger.UIInfo("Loaded %d Outbounds", len(outbounds))
		return OutboundsLoadedMsg{Outbounds: outbounds}
	}
}

// handleFileSelect receives selected file from picker.
func (m *Model) handleFileSelect(path string) tea.Cmd {
	if path == "" {
		logger.UIInfo("[%s]:File selection cancelled", m.name)
		return nil
	}

	logger.UIInfo("[%s]:outbound selected: %s", m.name, path)

	return input.ShowInputCmd(
		m.layout,
		"What do you want to call this outbound?",
		"outbound name",
		"",
		validation.ValidateFilename,
		func(_ string) tea.Cmd { return nil },
		func(filename string) tea.Cmd { return m.saveOutboundCmd(path, filename) },
	)
}

// saveOutboundCmd imports a new outbound template.
func (m *Model) saveOutboundCmd(srcPath, filename string) tea.Cmd {
	return func() tea.Msg {
		out, err := xray.SaveOutbound(srcPath, filename)
		if err != nil {
			logger.UIError("Failed to save outbound: %v", err)
			return m.errorCmd("Save Failed", err.Error())()
		}
		return AddOutboundMsg{Outbound: out}
	}
}

// deleteOutbound deletes a template from disk.
func (m *Model) deleteOutbound(outbound xray.XrayOutboundsFile) tea.Cmd {
	return func() tea.Msg {
		if err := os.Remove(outbound.Path); err != nil && !os.IsNotExist(err) {
			logger.UIError("Failed to delete outbound: %s", err.Error())
			return m.errorCmd("Delete Failed", err.Error())()
		}

		updated := make([]xray.XrayOutboundsFile, 0, len(m.outbounds)-1)
		for _, o := range m.outbounds {
			if o.Name != outbound.Name {
				updated = append(updated, o)
			}
		}

		delete(m.outboundsMap, outbound.Name)

		return OutboundsLoadedMsg{Outbounds: updated}
	}
}

// handleRequestDelete shows confirmation dialog.
func (m *Model) handleRequestDelete() tea.Cmd {
	t := m.table.(*table.Model)

	row := t.BubbleTable.SelectedRow()
	if row == nil {
		return m.infoCmd("No Selection", "Please select a row first")
	}

	name := row[0]
	out, ok := m.outboundsMap[name]
	if !ok {
		return m.errorCmd("Not Found", "Cannot find outbound")
	}

	return confirm.ConfirmCmd(
		m.layout,
		fmt.Sprintf("Delete outbound '%s'?", out.Name),
		func() tea.Msg { return DeleteOutboundMsg{Outbound: out} },
		false,
	)
}

// handleOutboundSelect executes onSelect callback.
func (m *Model) handleOutboundSelect() tea.Cmd {
	if m.onSelect == nil {
		return nil
	}

	t := m.table.(*table.Model)
	row := t.BubbleTable.SelectedRow()
	if row == nil {
		return m.infoCmd("No Selection", "Please select a row first")
	}

	name := row[0]
	out, ok := m.outboundsMap[name]
	if !ok {
		return m.errorCmd("Not Found", "Cannot find outbound")
	}

	return m.onSelect(out)
}

// handleRequestRename prompts user for new name.
func (m *Model) handleRequestRename() tea.Cmd {
	t := m.table.(*table.Model)

	row := t.BubbleTable.SelectedRow()
	if row == nil {
		return m.infoCmd("No Selection", "Please select a row first")
	}

	name := row[0]
	out, ok := m.outboundsMap[name]
	if !ok {
		return m.errorCmd("Not Found", "Cannot find outbound")
	}

	return input.ShowInputCmd(
		m.layout,
		"Enter new name for outbound:",
		"new name",
		out.Name,
		validation.ValidateFilename,
		nil,
		func(newName string) tea.Cmd { return m.renameOutbound(out, newName) },
	)
}

// renameOutbound renames a template on disk and updates list.
func (m *Model) renameOutbound(out *xray.XrayOutboundsFile, newName string) tea.Cmd {
	return func() tea.Msg {
		newMeta, err := xray.RenameOutboundTemplate(out.Name, newName)
		if err != nil {
			logger.UIError("Rename failed: %v", err)
			return m.errorCmd("Rename Failed", err.Error())()
		}

		updated := make([]xray.XrayOutboundsFile, 0, len(m.outbounds))
		inserted := false

		for _, o := range m.outbounds {
			if o.Name == out.Name {
				continue
			}

			if !inserted && newMeta.CreatedTime.After(o.CreatedTime) {
				updated = append(updated, *newMeta)
				inserted = true
			}

			updated = append(updated, o)
		}

		if !inserted {
			updated = append(updated, *newMeta)
		}

		delete(m.outboundsMap, out.Name)

		return OutboundsLoadedMsg{Outbounds: updated}
	}
}

func (m *Model) errorCmd(title, message string) tea.Cmd {
	return notice.NewNoticeCmd(m.layout, title, message, notice.NOTICE_ERROR)
}

func (m *Model) infoCmd(title, message string) tea.Cmd {
	return notice.NewNoticeCmd(m.layout, title, message, notice.NOTICE_INFO)
}

