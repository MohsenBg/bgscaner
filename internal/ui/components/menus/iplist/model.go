package iplist

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"bgscan/internal/core/iplist"
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

// Model represents the IP list view component.
//
// The component displays available IP list files in a table and allows
// the user to perform file operations such as:
//
//   - adding new IP files
//   - deleting existing files
//   - renaming files
//   - selecting a file for further processing
//
// Internally the component maintains a map of filenames to file metadata
// for fast lookup when handling table selections.
type Model struct {
	id       ui.ComponentID
	name     string
	layout   *layout.Layout
	table    ui.Component
	columns  []table.Column
	rows     []table.Row
	files    []iplist.IPFileInfo
	filesMap map[string]*iplist.IPFileInfo
	onSelect func(*iplist.IPFileInfo) tea.Cmd
}

// Default table columns used to render the IP list.
var ipListColumns = []table.Column{
	{Title: "Name", Width: 30},
	{Title: "Created Time", Width: 35},
	{Title: "Size", Width: 30},
}

// New creates a new IP list component.
//
// The component renders a table listing IP files located in the IP list
// directory. It registers several keybindings for file operations:
//
//	a — add a new IP file
//	x — delete the selected file
//	r — rename the selected file
//	enter — select the file (only if onSelect is provided)
//
// The onSelect callback is triggered when the user confirms a file selection.
func New(l *layout.Layout, title string, onSelect func(*iplist.IPFileInfo) tea.Cmd) *Model {
	m := &Model{
		layout:   l,
		id:       ui.NewComponentID(),
		name:     "IP Files",
		columns:  ipListColumns,
		rows:     []table.Row{},
		filesMap: make(map[string]*iplist.IPFileInfo),
		onSelect: onSelect,
	}

	t := table.New(title, ipListColumns, m.rows, l)

	keys := make([]table.ActionKey, 0, 4)

	if onSelect != nil {
		keys = append(
			keys,
			table.NewKey(
				[]string{tea.KeyEnter.String()},
				"select",
				"select ip File",
				func() tea.Msg { return SelectMsg{} },
			))
	}

	keys = append(keys,
		table.NewKey(
			[]string{"a"},
			"add file",
			"add ip file",
			picker.NewOpenPickFileCmd(
				l,
				"Select IP File .txt",
				"",
				[]string{".txt"},
				m.handleFileSelect,
			),
		),
		table.NewKey(
			[]string{"x"},
			"remove file",
			"remove ip file",
			func() tea.Msg { return RequestDeleteIPFileMsg{} },
		),
		table.NewKey(
			[]string{"r"},
			"rename file",
			"rename ip file",
			func() tea.Msg { return RequestRenameIPFileMsg{} },
		))

	t.SetKeys(keys...)

	m.table = t
	return m
}

// Init initializes the component and loads available IP files from disk.
func (m *Model) Init() tea.Cmd {
	return loadIPFilesCmd(m.layout)
}

// ID returns the unique identifier of the component.
func (m *Model) ID() ui.ComponentID {
	return m.id
}

// Name returns the display name of the component.
func (m *Model) Name() string {
	return m.name
}

// OnClose is invoked when the component is removed from the UI stack.
//
// The IP list component does not require cleanup, so it returns nil.
func (m *Model) OnClose() tea.Cmd {
	return nil
}

// Mode returns the interaction mode used by this component.
func (m *Model) Mode() env.Mode {
	return env.NormalMode
}

// loadIPFilesCmd loads all IP list files from disk and returns a message
// containing the discovered files.
//
// If loading fails, an error notice is displayed instead.
func loadIPFilesCmd(l *layout.Layout) tea.Cmd {
	return func() tea.Msg {
		files, err := iplist.ListIPFiles()
		if err != nil {
			logger.UIError("Failed to load IP files: %v", err)
			return notice.NewNoticeCmd(
				l,
				"Error Loading Files",
				fmt.Sprintf("Failed to load IP files: %v", err),
				notice.NOTICE_ERROR,
			)
		}

		logger.UIInfo("Loaded %d IP files", len(files))
		return IPFilesLoadedMsg{Files: files}
	}
}

// handleFileSelect is called when the file picker returns a selected path.
//
// The selected path is captured in a closure and passed to an input component
// which asks the user to provide a logical filename.
func (m *Model) handleFileSelect(path string) tea.Cmd {
	if path == "" {
		logger.UIInfo("File selection cancelled")
		return nil
	}

	logger.UIInfo("IP file selected: %s", path)

	return input.ShowInputCmd(
		m.layout,
		"What do you want to call this IP file?",
		"filename",
		"",
		validation.ValidateFilename,
		func(_ string) tea.Cmd { return nil },
		func(filename string) tea.Cmd { return m.saveIPFileCmd(path, filename) },
	)
}

// saveIPFileCmd copies the selected file into the managed IP list directory
// and registers it in the UI.
func (m *Model) saveIPFileCmd(srcPath, filename string) tea.Cmd {
	dstPath, err := iplist.GetIPFilePath(filename)
	if err != nil {
		logger.UIError("Failed to resolve destination path: %v", err)
		return m.errorCmd("Copy Failed", fmt.Sprintf("Failed to resolve destination path: %v", err))
	}

	if _, exists := m.filesMap[filename]; exists {
		return m.infoCmd("Duplicate File Name", "A file with this name already exists.")
	}

	if err := iplist.CopyIPFile(srcPath, dstPath); err != nil {
		logger.UIError("Failed to copy IP file: %v", err)
		return m.errorCmd("Copy Failed", fmt.Sprintf("Failed to copy IP file: %v", err))
	}

	fileInfo, err := iplist.GetIPFileInfo(dstPath)
	if err != nil {
		logger.UIError("Failed to read new IP file: %v", err)
		return m.errorCmd("Read Failed", fmt.Sprintf("Failed to read IP file info: %v", err))
	}

	return func() tea.Msg {
		return AddIPFileMsg{File: fileInfo}
	}
}

// deleteIPFile removes an IP file from disk and updates the file list.
func (m *Model) deleteIPFile(file iplist.IPFileInfo) tea.Cmd {
	if err := os.Remove(file.Path); err != nil && !os.IsNotExist(err) {
		logger.UIError("Failed to delete file: %v", err)
		return m.errorCmd("Delete File", fmt.Sprintf("Failed to delete file: %v", err))
	}

	return func() tea.Msg {
		files := make([]iplist.IPFileInfo, 0, len(m.files)-1)

		for _, f := range m.files {
			if f.Name != file.Name {
				files = append(files, f)
			}
		}

		delete(m.filesMap, file.Name)

		return IPFilesLoadedMsg{Files: files}
	}
}

// handleRequestDelete shows a confirmation dialog for deleting the
// currently selected IP file.
func (m *Model) handleRequestDelete() tea.Cmd {
	t, ok := m.table.(*table.Model)
	if !ok || t == nil {
		return m.errorCmd("Internal Error", "Cannot access table")
	}

	row := t.BubbleTable.SelectedRow()
	if row == nil {
		return m.infoCmd("No Selection", "Please select a row first")
	}

	filename := row[0]
	ipFile, exists := m.filesMap[filename]
	if !exists {
		return m.errorCmd("Not Found", "Cannot find IP file")
	}

	return confirm.ConfirmCmd(
		m.layout,
		fmt.Sprintf("Delete IP file '%s'?", ipFile.Name),
		func() tea.Msg {
			return DeleteIPFileMsg{File: *ipFile}
		},
		false,
	)
}

// handleIPFileSelect executes the onSelect callback for the currently
// selected IP file.
func (m *Model) handleIPFileSelect() tea.Cmd {
	t, ok := m.table.(*table.Model)
	if !ok || t == nil {
		return m.errorCmd("Internal Error", "Cannot access table")
	}

	row := t.BubbleTable.SelectedRow()
	if row == nil {
		return m.infoCmd("No Selection", "Please select a row first")
	}

	filename := row[0]
	ipFile, exists := m.filesMap[filename]
	if !exists {
		return m.errorCmd("Not Found", "Cannot find IP file")
	}

	return m.onSelect(ipFile)
}

// handleRequestRename prompts the user to provide a new name for
// the currently selected IP file.
func (m *Model) handleRequestRename() tea.Cmd {
	t, ok := m.table.(*table.Model)
	if !ok || t == nil {
		return m.errorCmd("Internal Error", "Cannot access table")
	}

	row := t.BubbleTable.SelectedRow()
	if row == nil {
		return m.infoCmd("No Selection", "Please select a row first")
	}

	filename := row[0]
	ipFile, exists := m.filesMap[filename]
	if !exists {
		return m.errorCmd("Not Found", "Cannot find IP file")
	}

	return input.ShowInputCmd(
		m.layout,
		"Enter new name for file:",
		"filename",
		ipFile.Name,
		validation.ValidateFilename,
		nil,
		func(newName string) tea.Cmd {
			return m.renameIPFile(*ipFile, newName)
		},
	)
}

// renameIPFile renames an IP file on disk and updates the list.
//
// The renamed file is reinserted into the list while preserving the
// sorting order (newest files first).
func (m *Model) renameIPFile(file iplist.IPFileInfo, newName string) tea.Cmd {
	dstPath, err := iplist.GetIPFilePath(newName)
	if err != nil {
		logger.UIError("Failed to resolve destination path: %v", err)
		return m.errorCmd("Rename Failed", fmt.Sprintf("Failed to resolve destination path: %v", err))
	}

	if _, exists := m.filesMap[newName]; exists {
		return m.infoCmd("Duplicate File Name", "A file with this name already exists.")
	}

	if err := os.Rename(file.Path, dstPath); err != nil {
		return m.errorCmd("Rename Failed", fmt.Sprintf("Failed to rename IP file: %v", err))
	}

	fileInfo, err := iplist.GetIPFileInfo(dstPath)
	if err != nil {
		logger.UIError("Failed to read renamed IP file: %v", err)
		return m.errorCmd("Read Failed", fmt.Sprintf("Failed to read IP file info: %v", err))
	}

	files := make([]iplist.IPFileInfo, 0, len(m.files))
	inserted := false

	for _, f := range m.files {
		if f.Name == file.Name {
			continue
		}

		if !inserted && fileInfo.CreatedAt.After(f.CreatedAt) {
			files = append(files, fileInfo)
			inserted = true
		}

		files = append(files, f)
	}

	if !inserted {
		files = append(files, fileInfo)
	}

	delete(m.filesMap, file.Name)

	return func() tea.Msg {
		return IPFilesLoadedMsg{Files: files}
	}
}

// errorCmd returns a notice command styled as an error message.
func (m *Model) errorCmd(title, message string) tea.Cmd {
	return notice.NewNoticeCmd(m.layout, title, message, notice.NOTICE_ERROR)
}

// infoCmd returns a notice command styled as an informational message.
func (m *Model) infoCmd(title, message string) tea.Cmd {
	return notice.NewNoticeCmd(m.layout, title, message, notice.NOTICE_INFO)
}
