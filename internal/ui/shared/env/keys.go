package env

import (
	"slices"

	tea "github.com/charmbracelet/bubbletea"
)

// backKeys defines which keys trigger the "back" action depending on the
// current UI mode.
//
// Each mode can override the keys that should navigate back in the UI.
// For example:
//
//   - NormalMode allows multiple keys for convenience.
//   - InputMode restricts the action to Escape.
//   - ScanMode disables it because cancellation is handled internally
//     by scanning goroutines.
var backKeys = map[Mode][]string{
	NormalMode: {"b", tea.KeyBackspace.String(), tea.KeyEscape.String()},
	InputMode:  {tea.KeyEscape.String()},
	ScanMode:   {}, // Scan handles back internally via goroutines
}

// quitKeys defines which keys terminate the application depending on
// the current UI mode.
var quitKeys = map[Mode][]string{
	NormalMode: {"q", tea.KeyCtrlC.String()},
	InputMode:  {tea.KeyCtrlC.String()},
	ScanMode:   {tea.KeyCtrlC.String()},
}

// IsBackKey reports whether the provided key should trigger a "back"
// navigation action for the given UI mode.
func IsBackKey(key tea.KeyMsg, mode Mode) bool {
	if keys, ok := backKeys[mode]; ok {
		return slices.Contains(keys, key.String())
	}
	return false
}

// IsQuitKey reports whether the provided key should terminate the
// application in the current UI mode.
func IsQuitKey(key tea.KeyMsg, mode Mode) bool {
	if keys, ok := quitKeys[mode]; ok {
		return slices.Contains(keys, key.String())
	}
	return false
}
