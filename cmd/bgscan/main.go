package main

import (
	"bgscan/internal/core/config"
	"bgscan/internal/logger"
	"bgscan/internal/ui/main/app"
	"bgscan/internal/ui/theme"
	"fmt"
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	// Initialize all loggers
	if err := logger.InitCore(); err != nil {
		log.Fatal("Failed to init core logger:", err)
	}

	if err := logger.InitUI(); err != nil {
		log.Fatal("Failed to init UI logger:", err)
	}

	if err := logger.InitDebug(); err != nil {
		log.Fatal("Failed to init debug logger:", err)
	}

	theme.Init()

	err := config.Init()
	if err != nil {
		log.Fatal(err)
	}

	// Ensure proper cleanup
	defer logger.CloseAll()

	p := tea.NewProgram(app.New())
	if _, err := p.Run(); err != nil {
		fmt.Printf("WTF, there's been an error: %v", err)
		os.Exit(1)
	}
}
