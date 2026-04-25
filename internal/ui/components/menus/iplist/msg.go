package iplist

import "bgscan/internal/core/iplist"

// IPFilesLoadedMsg is emitted whenever the IP file list is (re)loaded.
//
// This message is used as a synchronization point after operations that
// mutate the file set, including:
//
//   - initial directory scan
//   - adding a new IP file
//   - deleting an existing file
//   - renaming a file
//
// The Files slice represents the full, authoritative list of IP files
// and should replace any existing state.
type IPFilesLoadedMsg struct {
	Files []iplist.IPFileInfo
}

// AddIPFileMsg is emitted when a new IP file has been successfully copied
// into the managed IP list directory.
//
// The message contains the fully populated file metadata, ready to be
// merged into the current state.
type AddIPFileMsg struct {
	File iplist.IPFileInfo
}

// RequestDeleteIPFileMsg signals the user's intent to delete the
// currently selected IP file.
//
// This message does not perform deletion directly; it triggers the
// delete confirmation flow.
type RequestDeleteIPFileMsg struct{}

// DeleteIPFileMsg triggers the actual deletion of an IP file.
//
// This message is typically emitted after the user confirms the
// delete action in a confirmation dialog.
type DeleteIPFileMsg struct {
	File iplist.IPFileInfo
}

// RequestRenameIPFileMsg signals the user's intent to rename the
// currently selected IP file.
//
// It triggers the rename input flow, allowing the user to provide
// a new filename.
type RequestRenameIPFileMsg struct{}

// SelectMsg signals that the user has confirmed selection of the
// currently highlighted IP file.
//
// This message is usually bound to the Enter key and is handled
// only when a selection callback is configured.
type SelectMsg struct{}

