package resultlist

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"sort"

	"bgscan/internal/core/result"
	"bgscan/internal/logger"
	"bgscan/internal/ui/components/basic/confirm"
	"bgscan/internal/ui/components/basic/input"
	"bgscan/internal/ui/components/basic/notice"
	"bgscan/internal/ui/components/basic/table"
	ipviewer "bgscan/internal/ui/components/menus/ipviewer"
	"bgscan/internal/ui/shared/env"
	"bgscan/internal/ui/shared/layout"
	"bgscan/internal/ui/shared/ui"
	"bgscan/internal/ui/shared/validation"

	tea "github.com/charmbracelet/bubbletea"
)

// Model represents the Result Files menu component.
//
// It displays previously generated scan result files in a table and allows
// the user to:
//
//   - open and inspect result IPs
//   - rename result files
//   - delete result files
//
// The table is responsible for rendering and navigation, while this model
// orchestrates filesystem operations and viewer navigation.
type Model struct {
	id     ui.ComponentID
	name   string
	layout *layout.Layout

	table ui.Component

	// maxIPs defines the maximum number of IPs loaded when
	// opening a result file inside the IP viewer.
	maxIPs uint32

	// onSelect allows external components to override the default
	// behavior when a result file is selected.
	onSelect func(*result.ResultFile) tea.Cmd

	// filesMap maps file names to their metadata for quick lookup.
	filesMap map[string]*result.ResultFile

	// resultFiles stores the currently loaded result file metadata.
	resultFiles []result.ResultFile
}

// resultColumns defines the table layout used to display result files.
var resultColumns = []table.Column{
	{Title: "Name", Width: 50},
	{Title: "Created At", Width: 30},
	{Title: "Type", Width: 10},
	{Title: "Size", Width: 10},
}

// New creates and initializes a ResultList component.
func New(layout *layout.Layout, title string, onSelect func(*result.ResultFile) tea.Cmd) *Model {
	t := table.New(title, resultColumns, []table.Row{}, layout)

	keys := make([]table.ActionKey, 0, 3)

	keys = append(keys,
		table.NewKey(
			[]string{tea.KeyEnter.String()},
			"enter open",
			"enter view result IPs",
			func() tea.Msg { return SelectResultFileMsg{} },
		),
	)

	keys = append(keys,
		table.NewKey(
			[]string{"r"},
			"r rename",
			"r rename result file",
			func() tea.Msg { return RequestRenameResultFileMsg{} },
		),
		table.NewKey(
			[]string{"x"},
			"x delete",
			"x delete result file",
			func() tea.Msg { return RequestDeleteResultFileMsg{} },
		),
	)

	t.SetKeys(keys...)

	return &Model{
		id:          ui.NewComponentID(),
		name:        "Result Files",
		layout:      layout,
		table:       t,
		maxIPs:      10_000,
		onSelect:    onSelect,
		resultFiles: []result.ResultFile{},
		filesMap:    make(map[string]*result.ResultFile),
	}
}

// Mode returns the UI mode used by this component.
func (m *Model) Mode() env.Mode {
	return env.NormalMode
}

// Init loads result files from disk and populates the table.
//
// Files are sorted by creation time (newest first).
func (m *Model) Init() tea.Cmd {
	files, err := result.ListResultFiles(result.ResultAll)
	if err != nil {
		return m.errorCmd(
			"Result File Error",
			fmt.Sprintf("Error while reading result files: %v", err),
		)
	}

	sort.Slice(files, func(i, j int) bool {
		return files[i].CreatedTime.After(files[j].CreatedTime)
	})

	m.resultFiles = files

	// rebuild map
	for i := range m.resultFiles {
		f := &m.resultFiles[i]
		m.filesMap[result.NormalizeResultFileName(f.Name)] = f
	}

	return func() tea.Msg { return UpdateTableMsg{} }
}

// ID returns the unique component identifier.
func (m *Model) ID() ui.ComponentID {
	return m.id
}

// Name returns the display name of the component.
func (m *Model) Name() string {
	return m.name
}

// OnClose executes when the component is removed from the UI stack.
func (m *Model) OnClose() tea.Cmd {
	return nil
}

// OpenResultIP loads IP results from a result file and opens the IP viewer.
func (m *Model) OpenResultIP(file result.ResultFile) tea.Cmd {
	ips, err := result.LoadAll(file.Path, int64(m.maxIPs))
	if err != nil {
		return m.errorCmd(
			"Result File Error",
			fmt.Sprintf("Error while reading result file: %v", err),
		)
	}

	return func() tea.Msg {
		mode := ipviewer.ShortView

		if file.Type == result.ResultXRAY {
			mode = ipviewer.FullView
		}

		return ui.OpenComponentMsg{
			Component: ipviewer.New(
				m.layout,
				fmt.Sprintf("IP Scan [%s]", file.Type.String()),
				ips,
				mode,
			),
		}
	}
}

// deleteResultFile removes a result file from disk and updates the UI list.
func (m *Model) deleteResultFile(file result.ResultFile) tea.Cmd {
	if err := os.Remove(file.Path); err != nil && !os.IsNotExist(err) {
		logger.UIError("Failed to delete file: %v", err)
		return m.errorCmd("Delete Failed", fmt.Sprintf("Failed to delete file: %v", err))
	}

	files := make([]result.ResultFile, 0, len(m.resultFiles)-1)

	for _, f := range m.resultFiles {
		if f.Name != file.Name {
			files = append(files, f)
		}
	}

	delete(m.filesMap, file.Name)
	m.resultFiles = files

	return func() tea.Msg { return UpdateTableMsg{} }
}

// handleRequestDelete shows a confirmation dialog for deleting
// the currently selected result file.
func (m *Model) handleRequestDelete() tea.Cmd {
	t, ok := m.table.(*table.Model)
	if !ok || t == nil {
		return m.errorCmd("Internal Error", "Cannot access table")
	}

	row := t.BubbleTable.SelectedRow()
	if row == nil {
		return m.infoCmd("No Selection", "Please select a row first")
	}

	filename := result.NormalizeResultFileName(row[0])

	file, exists := m.filesMap[filename]
	if !exists {
		return m.errorCmd("Not Found", "Cannot find result file")
	}

	return confirm.ConfirmCmd(
		m.layout,
		fmt.Sprintf("Delete result file '%s'?", file.Name),
		func() tea.Msg {
			return DeleteResultFileMsg{File: *file}
		},
		false,
	)
}

// handleResultFileSelect executes the selection action for the
// currently selected result file.
func (m *Model) handleResultFileSelect() tea.Cmd {
	t, ok := m.table.(*table.Model)
	if !ok || t == nil {
		return m.errorCmd("Internal Error", "Cannot access table")
	}

	row := t.BubbleTable.SelectedRow()
	if row == nil {
		return m.infoCmd("No Selection", "Please select a row first")
	}

	filename := result.NormalizeResultFileName(row[0])

	file, exists := m.filesMap[filename]
	if !exists {
		return m.errorCmd("Not Found", "Cannot find result file")
	}

	if m.onSelect != nil {
		return m.onSelect(file)
	}

	return m.OpenResultIP(*file)
}

// handleRequestRename prompts the user to provide a new name
// for the currently selected result file.
func (m *Model) handleRequestRename() tea.Cmd {
	t, ok := m.table.(*table.Model)
	if !ok || t == nil {
		return m.errorCmd("Internal Error", "Cannot access table")
	}

	row := t.BubbleTable.SelectedRow()
	if row == nil {
		return m.infoCmd("No Selection", "Please select a row first")
	}

	filename := result.NormalizeResultFileName(row[0])

	file, exists := m.filesMap[filename]
	if !exists {
		return m.errorCmd("Not Found", "Cannot find result file")
	}

	return input.ShowInputCmd(
		m.layout,
		"Enter new name for file:",
		"filename",
		file.Name,
		validation.ValidateFilename,
		nil,
		func(newName string) tea.Cmd {
			return m.renameResultFile(*file, newName)
		},
	)
}

// renameResultFile renames a result file on disk and updates the list.
func (m *Model) renameResultFile(file result.ResultFile, newName string) tea.Cmd {

	newName = result.NormalizeResultFileName(newName)

	if _, exists := m.filesMap[newName]; exists {
		return m.infoCmd("Duplicate File Name", "A file with this name already exists.")
	}

	dstPath := path.Join(filepath.Dir(file.Path), newName)

	if err := os.Rename(file.Path, dstPath); err != nil {
		return m.errorCmd("Rename Failed", fmt.Sprintf("Failed to rename file: %v", err))
	}

	newFile, err := result.GetResultFileInfo(dstPath)
	if err != nil {
		logger.UIError("Failed to read renamed file: %v", err)
		return m.errorCmd("Read Failed", fmt.Sprintf("Failed to read file info: %v", err))
	}

	files := make([]result.ResultFile, 0, len(m.resultFiles))
	inserted := false

	for _, f := range m.resultFiles {
		if f.Name == file.Name {
			continue
		}

		if !inserted && newFile.CreatedTime.After(f.CreatedTime) {
			files = append(files, newFile)

			idx := len(files) - 1
			m.filesMap[result.NormalizeResultFileName(files[idx].Name)] = &files[idx]

			inserted = true
		}

		files = append(files, f)
	}

	if !inserted {
		files = append(files, newFile)
	}

	delete(m.filesMap, result.NormalizeResultFileName(file.Name))
	m.resultFiles = files

	return func() tea.Msg { return UpdateTableMsg{} }
}

// errorCmd returns an error notice command.
func (m *Model) errorCmd(title, message string) tea.Cmd {
	return notice.NewNoticeCmd(m.layout, title, message, notice.NOTICE_ERROR)
}

// infoCmd returns an informational notice command.
func (m *Model) infoCmd(title, message string) tea.Cmd {
	return notice.NewNoticeCmd(m.layout, title, message, notice.NOTICE_INFO)
}
