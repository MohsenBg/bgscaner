package main

import (
	"bgscan/logger"
	"bgscan/ui/main/app"
	"bgscan/ui/theme"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	theme.Init()

	// Init debug logger
	if err := logger.Init("app"); err != nil {
		fmt.Printf("Failed to init debug: %v\n", err)
		os.Exit(1)
	}
	defer logger.Close()

	p := tea.NewProgram(app.New())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
