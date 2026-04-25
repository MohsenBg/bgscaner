package outbounds

import "bgscan/internal/core/xray"

// OutboundsLoadedMsg is emitted whenever the outbound template list
// is (re)loaded.
//
// This message is dispatched:
//
//   - after initial directory scan
//   - after adding a new outbound template
//   - after deleting an outbound template
//   - after renaming an outbound template
//
// The Outbounds slice represents the full, authoritative list of
// outbound templates and should replace any existing state.
type OutboundsLoadedMsg struct {
	Outbounds []xray.XrayOutboundsFile
}

// AddOutboundMsg is emitted when a new outbound template has been
// successfully saved into the managed templates directory.
//
// The Outbound field contains the fully populated metadata object.
type AddOutboundMsg struct {
	Outbound *xray.XrayOutboundsFile
}

// RequestDeleteOutboundMsg signals the user's intent to delete the
// currently selected outbound template.
//
// This triggers the delete confirmation flow.
type RequestDeleteOutboundMsg struct{}

// DeleteOutboundMsg triggers the actual deletion of an outbound template.
//
// This message is typically emitted after the user confirms the
// delete action via the confirmation dialog.
type DeleteOutboundMsg struct {
	Outbound *xray.XrayOutboundsFile
}

// RequestRenameOutboundMsg signals the user's intent to rename the
// currently selected outbound template.
//
// This transitions into the rename input flow.
type RequestRenameOutboundMsg struct{}

// SelectMsg signals that the user has confirmed selection of the
// currently highlighted outbound template.
//
// This is usually bound to the Enter key and is only handled when
// an onSelect callback is configured.
type SelectMsg struct{}
