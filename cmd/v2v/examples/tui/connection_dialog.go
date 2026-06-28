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

package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ConnectionDialog represents a form for entering Vertica connection details
type ConnectionDialog struct {
	inputs     []textinput.Model
	focusIndex int
	submitted  bool
	paneSide   Pane // Which pane is this connection for
	width      int
	height     int
}

const (
	hostField = iota
	portField
	userField
	passwordField
	databaseField
)

// NewConnectionDialog creates a new connection dialog
func NewConnectionDialog(paneSide Pane) *ConnectionDialog {
	inputs := make([]textinput.Model, 5)

	// Host
	inputs[hostField] = textinput.New()
	inputs[hostField].Placeholder = "localhost"
	inputs[hostField].Focus()
	inputs[hostField].CharLimit = 256
	inputs[hostField].Width = 40
	inputs[hostField].Prompt = "Host: "

	// Port
	inputs[portField] = textinput.New()
	inputs[portField].Placeholder = "5433"
	inputs[portField].CharLimit = 5
	inputs[portField].Width = 40
	inputs[portField].Prompt = "Port: "

	// User
	inputs[userField] = textinput.New()
	inputs[userField].Placeholder = "dbadmin"
	inputs[userField].CharLimit = 128
	inputs[userField].Width = 40
	inputs[userField].Prompt = "User: "

	// Password
	inputs[passwordField] = textinput.New()
	inputs[passwordField].Placeholder = "password"
	inputs[passwordField].CharLimit = 256
	inputs[passwordField].Width = 40
	inputs[passwordField].EchoMode = textinput.EchoPassword
	inputs[passwordField].EchoCharacter = '•'
	inputs[passwordField].Prompt = "Pass: "

	// Database
	inputs[databaseField] = textinput.New()
	inputs[databaseField].Placeholder = "VMart"
	inputs[databaseField].CharLimit = 128
	inputs[databaseField].Width = 40
	inputs[databaseField].Prompt = "DB:   "

	return &ConnectionDialog{
		inputs:     inputs,
		focusIndex: 0,
		paneSide:   paneSide,
		width:      60,
		height:     20,
	}
}

// Update handles messages for the connection dialog
func (d *ConnectionDialog) Update(msg tea.Msg) (*ConnectionDialog, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			// Move to next field or submit
			if d.focusIndex == len(d.inputs)-1 {
				d.submitted = true
				return d, nil
			}
			d.focusIndex++
			d.updateFocus()
			return d, nil

		case "shift+tab", "up":
			// Move to previous field
			if d.focusIndex > 0 {
				d.focusIndex--
				d.updateFocus()
			}
			return d, nil

		case "tab", "down":
			// Move to next field
			if d.focusIndex < len(d.inputs)-1 {
				d.focusIndex++
				d.updateFocus()
			}
			return d, nil
		}
	}

	// Update the focused input
	cmd := d.updateInput(msg)
	return d, cmd
}

// updateInput updates the currently focused input
func (d *ConnectionDialog) updateInput(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	d.inputs[d.focusIndex], cmd = d.inputs[d.focusIndex].Update(msg)
	return cmd
}

// updateFocus updates which input field has focus
func (d *ConnectionDialog) updateFocus() {
	for i := range d.inputs {
		if i == d.focusIndex {
			d.inputs[i].Focus()
		} else {
			d.inputs[i].Blur()
		}
	}
}

// View renders the connection dialog
func (d *ConnectionDialog) View() string {
	dialogStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		Padding(1, 2).
		Width(d.width)

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("62")).
		Align(lipgloss.Center).
		Width(d.width - 4)

	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Italic(true).
		MarginTop(1)

	paneName := "Left"
	if d.paneSide == RightPane {
		paneName = "Right"
	}

	title := titleStyle.Render(fmt.Sprintf("Connect %s Pane to Vertica", paneName))

	var b strings.Builder
	b.WriteString(title)
	b.WriteString("\n\n")

	for i, input := range d.inputs {
		b.WriteString(input.View())
		if i < len(d.inputs)-1 {
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")
	b.WriteString(helpStyle.Render("Tab/↑↓: Navigate • Enter: Next/Connect • Esc: Cancel"))

	return dialogStyle.Render(b.String())
}

// GetConnectionParams returns the entered connection parameters
func (d *ConnectionDialog) GetConnectionParams() (host, port, user, password, database string) {
	host = d.inputs[hostField].Value()
	if host == "" {
		host = d.inputs[hostField].Placeholder
	}

	port = d.inputs[portField].Value()
	if port == "" {
		port = d.inputs[portField].Placeholder
	}

	user = d.inputs[userField].Value()
	if user == "" {
		user = d.inputs[userField].Placeholder
	}

	password = d.inputs[passwordField].Value()
	if password == "" {
		password = d.inputs[passwordField].Placeholder
	}

	database = d.inputs[databaseField].Value()
	if database == "" {
		database = d.inputs[databaseField].Placeholder
	}

	return
}

// IsSubmitted returns whether the form was submitted
func (d *ConnectionDialog) IsSubmitted() bool {
	return d.submitted
}
