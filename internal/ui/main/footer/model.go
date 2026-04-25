package footer

import (
	"bgscan/internal/core/config"
	"bgscan/internal/ui/shared/env"
	"bgscan/internal/ui/shared/layout"
	"bgscan/internal/ui/shared/ui"
	"runtime"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// Model represents the footer component responsible for
// displaying runtime information and application status.
type Model struct {
	layout *layout.Layout

	id   ui.ComponentID
	name string

	appVersion string
	status     string

	goroutines  int
	memoryBytes uint64
	sys         uint64
}

// RuntimeStats contains runtime metrics collected
// from the Go runtime.
type RuntimeStats struct {
	Goroutines  int
	MemoryBytes uint64
	Sys         uint64
}

// timesTickMsg is emitted every second to update
// runtime metrics displayed in the footer.
type timesTickMsg time.Time

// New creates a new footer component.
func New(l *layout.Layout) *Model {
	return &Model{
		id:         ui.NewComponentID(),
		name:       "footer",
		layout:     l,
		appVersion: config.AppVersion,
		status:     "Main Menu",
	}
}

// ID returns the unique component identifier.
func (m *Model) ID() ui.ComponentID {
	return m.id
}

// Name returns the component name.
func (m *Model) Name() string {
	return m.name
}

// Mode returns the interaction mode of the component.
func (m *Model) Mode() env.Mode {
	return env.NormalMode
}

// Init starts the periodic runtime stats ticker.
func (m *Model) Init() tea.Cmd {
	return tickCmd()
}

// OnClose runs cleanup logic when the component is removed.
func (m *Model) OnClose() tea.Cmd {
	return nil
}

// tickCmd schedules a runtime stats refresh every second.
func tickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return timesTickMsg(t)
	})
}

// getRuntimeStats collects runtime metrics such as
// goroutine count and memory usage.
func getRuntimeStats() RuntimeStats {
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)

	return RuntimeStats{
		Goroutines:  runtime.NumGoroutine(),
		MemoryBytes: mem.Alloc,
		Sys:         mem.Sys,
	}
}
