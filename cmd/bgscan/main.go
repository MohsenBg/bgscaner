package main

import (
	"bgscan/internal/logger"
	"bgscan/internal/startup"
	"bgscan/internal/ui/main/app"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {

	startup.RunHealthChecks()

	defer logger.CloseAll()

	p := tea.NewProgram(app.New())

	if _, err := p.Run(); err != nil {
		fmt.Printf("BubbleTea runtime error:%s", err.Error())
		os.Exit(1)
	}
}
