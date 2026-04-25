package logview

// LogUpdateTickMsg is emitted periodically to trigger
// log viewport refreshes.
//
// The log viewer uses this message to safely transfer
// buffered log messages collected by the background
// logger subscription goroutine into the BubbleTea UI
// update loop.
type LogUpdateTickMsg struct{}
