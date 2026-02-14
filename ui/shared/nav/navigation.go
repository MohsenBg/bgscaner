package nav

import tea "github.com/charmbracelet/bubbletea"

type ViewName string

const (
	// Root
	ViewMainMenu ViewName = "main_menu"

	// Main menu views
	ViewScanMenu     ViewName = "scan_menu"
	ViewSettingsMenu ViewName = "settings_menu"
	ViewIPList       ViewName = "ip_list"
	ViewResultsMenu  ViewName = "results_menu"
	ViewAboutMenu    ViewName = "about_menu"

	// Settings sub-menus
	ViewICMPConfigMenu ViewName = "icmp_config_menu"
	ViewTCPConfigMenu  ViewName = "tcp_config_menu"
	ViewHTTPConfigMenu ViewName = "http_config_menu"
	ViewXRayConfigMenu ViewName = "xray_config_menu"
)

var ViewTitles = map[ViewName]string{
	ViewMainMenu: "Main Menu",

	// Main menu
	ViewScanMenu:     "Run Scan",
	ViewSettingsMenu: "Settings",
	ViewIPList:       "IP Lists",
	ViewResultsMenu:  "Results",
	ViewAboutMenu:    "About",

	// Settings
	ViewICMPConfigMenu: "ICMP Configuration",
	ViewTCPConfigMenu:  "TCP Configuration",
	ViewHTTPConfigMenu: "HTTP Configuration",
	ViewXRayConfigMenu: "Xray Configuration",
}

// OpenViewMsg requests navigation to another view.
type OpenViewMsg struct {
	View ViewName
}

// View binds a Bubble Tea model to a logical view.
type View struct {
	Name  ViewName // unique identifier
	Model tea.Model
}

func TitleFor(v ViewName) string {
	if t, ok := ViewTitles[v]; ok {
		return t
	}
	return string(v)
}
