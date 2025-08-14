package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"muxyard/internal/config"
	"muxyard/internal/ui"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	p := tea.NewProgram(ui.NewMainModel(cfg), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
}
