package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/patraden/toolkit/cmd/v2v/internal/tui"
)

func main() {

	// Initialize the TUI model
	m := tui.NewModel()

	// Create a new program
	p := tea.NewProgram(m, tea.WithAltScreen())

	// Run the program
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}

}
