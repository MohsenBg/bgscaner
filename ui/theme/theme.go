package theme

import (
	"os"

	"github.com/charmbracelet/lipgloss"
)

/*
	====================
	Theme Mode
	====================
*/

type ThemeMode int

const (
	ModeAuto ThemeMode = iota
	ModeDark
	ModeLight
)

/*
	====================
	Theme Definition
	====================
*/

type Theme struct {
	Primary   lipgloss.TerminalColor
	Secondary lipgloss.TerminalColor

	Border       lipgloss.TerminalColor
	BorderActive lipgloss.TerminalColor

	Text    lipgloss.TerminalColor
	Muted   lipgloss.TerminalColor
	Info    lipgloss.TerminalColor
	Error   lipgloss.TerminalColor
	Success lipgloss.TerminalColor

	AccentOrange lipgloss.TerminalColor
	AccentYellow lipgloss.TerminalColor
	AccentPurple lipgloss.TerminalColor
}

/*
	====================
	Color Palettes
	====================
	Purple primary, gray secondary
	Accent colors: orange & yellow
*/

var Dark = Theme{
	Primary:      lipgloss.Color("170"), // purple
	Secondary:    lipgloss.Color("245"), // gray
	Border:       lipgloss.Color("240"),
	BorderActive: lipgloss.Color("62"), // purple highlight
	Text:         lipgloss.Color("252"),
	Muted:        lipgloss.Color("241"),
	Error:        lipgloss.Color("196"),
	Success:      lipgloss.Color("42"),
	Info:         lipgloss.Color("39"),
	AccentOrange: lipgloss.Color("208"), // orange
	AccentYellow: lipgloss.Color("220"), // yellow
	AccentPurple: lipgloss.Color("62"),
}

var Light = Theme{
	Primary:      lipgloss.Color("170"), // purple
	Secondary:    lipgloss.Color("248"), // light gray
	Border:       lipgloss.Color("244"),
	BorderActive: lipgloss.Color("62"),
	Text:         lipgloss.Color("234"),
	Muted:        lipgloss.Color("246"),
	Info:         lipgloss.Color("27"),
	Error:        lipgloss.Color("160"),
	Success:      lipgloss.Color("28"),
	AccentOrange: lipgloss.Color("208"), // orange
	AccentYellow: lipgloss.Color("220"), // yellow
	AccentPurple: lipgloss.Color("62"),
}

/*
	====================
	State
	====================
*/

var (
	current Theme
	mode    = ModeAuto
)

/*
	====================
	Public API
	====================
*/

func Current() Theme {
	return current
}

func Mode() ThemeMode {
	return mode
}

func SetMode(m ThemeMode) {
	mode = m
	resolve()
}

/*
	====================
	Theme Resolver
	====================
*/

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

/*
	====================
	Terminal Detection
	====================
	Best-effort heuristic.
*/

func terminalLooksDark() bool {
	bg := os.Getenv("COLORFGBG")
	if bg == "" {
		return true // safe default
	}

	// COLORFGBG is usually "15;0" or "0;15"
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

/*
	====================
	Init
	====================
*/

func Init() {
	resolve()
}

