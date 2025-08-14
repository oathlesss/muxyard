package main

import (
	"flag"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"muxyard/internal/config"
	"muxyard/internal/tmux"
	"muxyard/internal/ui"
)

var version = "dev" // Set by build flags

func main() {
	var showVersion = flag.Bool("version", false, "Show version information")
	var showHelp = flag.Bool("help", false, "Show help information")
	flag.Parse()

	if *showVersion {
		fmt.Printf("muxyard version %s\n", version)
		os.Exit(0)
	}

	if *showHelp {
		fmt.Println("Muxyard - Tmux Session Manager")
		fmt.Println("")
		fmt.Println("Usage:")
		fmt.Println("  muxyard              Start the interactive TUI")
		fmt.Println("  muxyard --version    Show version information")
		fmt.Println("  muxyard --help       Show this help message")
		fmt.Println("")
		fmt.Println("For more information, visit: https://github.com/rubenhesselink/muxyard")
		os.Exit(0)
	}

	// Check if tmux is available
	if !tmux.IsTmuxAvailable() {
		fmt.Println("Error: tmux is not installed or not found in PATH")
		fmt.Println("Please install tmux to use muxyard")
		os.Exit(1)
	}

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
