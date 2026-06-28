// Copyright 2025 The Toolkit Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/patraden/toolkit/cmd/v2v/tui"
)

func main() {
	// For now, we'll start with mocked data
	// Later, this will accept connection strings as arguments

	if len(os.Args) > 1 && (os.Args[1] == "-h" || os.Args[1] == "--help") {
		printHelp()
		return
	}

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

func printHelp() {
	fmt.Println(`v2v - Vertica to Vertica table copy tool

A Far Manager-style TUI for copying tables between Vertica clusters.

Usage:
  v2v [flags]

Flags:
  -h, --help     Show this help message

Controls:
  F2             Connect active pane to Vertica cluster
  F3             Disconnect active pane
  F4             Toggle operation log panel
  F6             Focus/unfocus log panel
  Tab            Switch between left and right pane
  ↑/↓            Navigate tables or scroll logs (when focused)
  PgUp/PgDn      Fast scroll in logs
  Home/End       Jump to top/bottom of logs
  F5             Copy selected table from active pane to other pane
  Esc            Cancel connection dialog
  q/Ctrl+C       Quit

Note: Currently running in simulation mode with mock data.
Connection dialog is functional, but uses simulated backend.

Examples:
  v2v                                      # Run with mock data
  v2v -left "vertica://..." -right "..."   # Connect to real Vertica clusters (TODO)`)
}
