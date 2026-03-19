package theme

import (
	"os"

	"github.com/charmbracelet/lipgloss"
)

// Package theme provides a centralized color palette and theme management
// system for the terminal UI.
//
// It supports three modes:
//
//   - ModeDark  – forces the dark color palette
//   - ModeLight – forces the light color palette
//   - ModeAuto  – automatically selects a palette based on terminal background
//
// The package exposes a minimal API used by UI components to retrieve
// the active theme and react to theme changes.
//
// Example:
//
//	th := theme.Current()
//
//	title := lipgloss.NewStyle().
//		Foreground(th.Primary).
//		Bold(true)
//
//	fmt.Println(title.Render("BGScan"))

type ThemeMode int

const (
	// ModeAuto selects the theme automatically based on terminal detection.
	ModeAuto ThemeMode = iota

	// ModeDark forces the dark color palette.
	ModeDark

	// ModeLight forces the light color palette.
	ModeLight
)

// Theme represents the complete color palette used by the UI.
//
// All UI components should retrieve colors through this struct instead
// of defining colors directly. This keeps the UI visually consistent
// and allows the palette to be swapped dynamically.
type Theme struct {
	Primary   lipgloss.TerminalColor
	Secondary lipgloss.TerminalColor

	Border       lipgloss.TerminalColor
	BorderActive lipgloss.TerminalColor

	Text  lipgloss.TerminalColor
	Muted lipgloss.TerminalColor

	Info    lipgloss.TerminalColor
	Error   lipgloss.TerminalColor
	Success lipgloss.TerminalColor

	Orange lipgloss.TerminalColor
	Yellow lipgloss.TerminalColor
	Purple lipgloss.TerminalColor
}

// Dark defines the dark terminal color palette.
var Dark = Theme{
	Primary:      lipgloss.Color("170"),
	Secondary:    lipgloss.Color("245"),
	Border:       lipgloss.Color("240"),
	BorderActive: lipgloss.Color("62"),
	Text:         lipgloss.Color("252"),
	Muted:        lipgloss.Color("241"),
	Error:        lipgloss.Color("196"),
	Success:      lipgloss.Color("42"),
	Info:         lipgloss.Color("39"),
	Orange:       lipgloss.Color("208"),
	Yellow:       lipgloss.Color("220"),
	Purple:       lipgloss.Color("62"),
}

// Light defines the light terminal color palette.
var Light = Theme{
	Primary:      lipgloss.Color("170"),
	Secondary:    lipgloss.Color("248"),
	Border:       lipgloss.Color("244"),
	BorderActive: lipgloss.Color("62"),
	Text:         lipgloss.Color("234"),
	Muted:        lipgloss.Color("246"),
	Info:         lipgloss.Color("27"),
	Error:        lipgloss.Color("160"),
	Success:      lipgloss.Color("28"),
	Orange:       lipgloss.Color("208"),
	Yellow:       lipgloss.Color("220"),
	Purple:       lipgloss.Color("62"),
}

var (
	current Theme
	mode    = ModeAuto
)

// Current returns the active theme palette.
func Current() Theme {
	return current
}

// Mode returns the currently configured ThemeMode.
func Mode() ThemeMode {
	return mode
}

// SetMode changes the active theme mode and resolves
// the appropriate palette.
func SetMode(m ThemeMode) {
	mode = m
	resolve()
}

func resolve() {
	switch mode {

	case ModeDark:
		current = Dark

	case ModeLight:
		current = Light

	case ModeAuto:
		if terminalLooksDark() {
			current = Dark
		} else {
			current = Light
		}

	}
}

// terminalLooksDark attempts to detect whether the terminal
// background is dark using the COLORFGBG environment variable.
func terminalLooksDark() bool {

	bg := os.Getenv("COLORFGBG")

	if bg == "" {
		return true
	}

	for i := len(bg) - 1; i >= 0; i-- {
		if bg[i] == ';' {
			bg = bg[i+1:]
			break
		}
	}

	switch bg {
	case "0", "1", "2", "3", "4", "5", "6", "7":
		return true
	default:
		return false
	}
}

// Init initializes the theme system.
// Call this once during application startup.
func Init() {
	resolve()
}
