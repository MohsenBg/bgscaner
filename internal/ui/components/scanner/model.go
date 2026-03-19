package scanner

import (
	"bgscan/internal/core/config"
	"bgscan/internal/core/result"
	"bgscan/internal/core/scanner"
	"bgscan/internal/core/scanner/engine"
	"bgscan/internal/ui/components/basic/progress"
	"bgscan/internal/ui/components/basic/table"
	"bgscan/internal/ui/components/menus/ipviewer"
	"bgscan/internal/ui/shared/env"
	"bgscan/internal/ui/shared/layout"
	"bgscan/internal/ui/shared/ui"

	"sort"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// tickMsg is sent periodically to update UI state.
type tickMsg struct{}

// ScannerStatus represents the lifecycle state of the scanner.
type ScannerStatus string

const (
	StatusPreProcess ScannerStatus = "PreProcess"
	StatusStart      ScannerStatus = "Start"
	StatusScanning   ScannerStatus = "Scanning"
	StatusEnded      ScannerStatus = "Ended"
	StatusError      ScannerStatus = "Error"
)

// Model implements the scanning UI component.
//
// Responsibilities:
//   - Run the background scanner engine
//   - Collect scan results safely from goroutines
//   - Periodically merge results into the UI state
//   - Maintain progress and scanner status
type Model struct {
	id     ui.ComponentID
	name   string
	layout *layout.Layout

	// Scanner engine
	scanner scanner.Scanner
	maxIPs  int

	// UI components
	progress  ui.Component
	ipViewer  ui.Component
	logViewer ui.Component

	// Results state
	ips []result.IPScanResult

	// ---- Concurrency (scanner → UI bridge) ----
	mu           sync.Mutex
	batch        []result.IPScanResult
	progressInfo engine.Progress

	status    ScannerStatus
	scanError error
}

// New creates a new Scanner component.
func New(layout *layout.Layout, maxIPs int, scannerObj scanner.Scanner) *Model {

	mode := ipviewer.ShortView
	if scanner.XRAY_SCAN == scannerObj.Mode() {
		mode = ipviewer.FullView
	}

	ipViewer := ipviewer.New(layout, "IP Result Scan", []result.IPScanResult{}, mode)

	ipViewer.Table().SetKeys(
		table.NewKey(
			[]string{"p"},
			"p pause/resume",
			"pause/resume scan",
			func() tea.Msg { return TogglePauseMsg{} },
		),
		table.NewKey(
			[]string{"l"},
			"l log",
			"view logs",
			nil,
		))

	return &Model{
		id:       ui.NewComponentID(),
		name:     "Scanner",
		layout:   layout,
		maxIPs:   maxIPs,
		scanner:  scannerObj,
		progress: progress.New(layout),
		ipViewer: ipViewer,

		ips:   make([]result.IPScanResult, 0, maxIPs),
		batch: make([]result.IPScanResult, 0, 50),

		status: StatusPreProcess,
	}
}

// --- ui.Component implementation ---

func (m *Model) ID() ui.ComponentID { return m.id }

func (m *Model) Name() string { return m.name }

func (m *Model) Mode() env.Mode { return env.ScanMode }

func (m *Model) OnClose() tea.Cmd { return nil }

// Init starts the scanner in a background goroutine.
func (m *Model) Init() tea.Cmd {

	go func() {

		// Step 1: preprocessing phase
		if err := m.scanner.PreProcess(); err != nil {
			m.onError(err)
			return
		}

		m.mu.Lock()
		m.status = StatusScanning
		m.mu.Unlock()

		// Step 2: start scanning
		m.scanner.Scan(
			engine.ScanHooks{
				OnProgress: m.onProgress,
				OnSuccess:  m.onSuccess,
				OnScanEnd:  m.onScanEnd,
				OnError:    m.onError,
			},
		)
	}()

	return m.tick()
}

// --- Tick Loop ---

// tick triggers periodic UI updates.
func (m *Model) tick() tea.Cmd {
	interval := config.Get().General.StatusInterval.Duration()
	return tea.Tick(interval, func(time.Time) tea.Msg {
		return tickMsg{}
	})
}

// --- Scanner Callbacks (Background Goroutine) ---

// onSuccess receives successful scan results from the engine.
func (m *Model) onSuccess(ip result.IPScanResult) {

	m.mu.Lock()
	m.batch = append(m.batch, ip)
	m.mu.Unlock()
}

// onProgress receives progress updates from the engine.
func (m *Model) onProgress(p engine.Progress) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.status == StatusScanning {
		m.progressInfo = p
	}
}

// onError records scanner errors.
func (m *Model) onError(err error) {

	m.mu.Lock()
	m.status = StatusError
	m.scanError = err
	m.mu.Unlock()
}

// onScanEnd marks the scanner as finished.
func (m *Model) onScanEnd() {
	m.mu.Lock()
	m.status = StatusEnded
	m.mu.Unlock()
}

// --- UI Thread Logic ---

// mergeBatch transfers results from the background buffer
// into the UI result list.
func (m *Model) mergeBatch() {

	m.mu.Lock()

	if len(m.batch) == 0 {
		m.mu.Unlock()
		return
	}

	newIPs := make([]result.IPScanResult, len(m.batch))
	copy(newIPs, m.batch)

	m.batch = m.batch[:0]

	m.mu.Unlock()

	// Merge results
	m.ips = append(m.ips, newIPs...)

	// Sort by latency (fastest first)
	sort.Slice(m.ips, func(i, j int) bool {
		return m.ips[i].Latency < m.ips[j].Latency
	})

	// Limit result list
	if len(m.ips) > m.maxIPs {
		m.ips = m.ips[:m.maxIPs]
	}
}
