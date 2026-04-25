package scanner

// tickMsg is an internal message used to drive the
// scanner's periodic update loop.
//
// It is typically triggered by a BubbleTea tick command and
// allows the scanner component to refresh progress, poll
// results, or update UI state while a scan is running.
type tickMsg struct{}

// TogglePauseMsg is emitted when the user requests to
// pause or resume the active scan.
//
// The scanner component listens for this message and
// toggles its paused state accordingly.
type TogglePauseMsg struct{}

type immediateTickMsg struct{}
