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
	"time"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Pane represents which pane is active
type Pane int

const (
	LeftPane Pane = iota
	RightPane
)

// ConnectionState represents the connection status of a pane
type ConnectionState int

const (
	Disconnected ConnectionState = iota
	Connecting
	Connected
	ConnectionError
)

// ViewMode represents what the user is currently viewing
type ViewMode int

const (
	NormalView ViewMode = iota
	ConnectionDialogView
)

// Model holds the application state
type Model struct {
	leftPanel        *TablePanel
	rightPanel       *TablePanel
	leftConnState    ConnectionState
	rightConnState   ConnectionState
	leftConnError    string
	rightConnError   string
	leftConnString   string
	rightConnString  string
	activePane       Pane
	viewMode         ViewMode
	connectionDialog *ConnectionDialog
	progressBar      progress.Model
	connectingPane   Pane
	logPanel         *LogPanel
	width            int
	height           int
	statusMsg        string
	copying          bool
}

// CopyCompleteMsg is sent when a copy operation completes
type CopyCompleteMsg struct {
	success bool
	message string
}

// ConnectionStartMsg indicates connection is starting
type ConnectionStartMsg struct {
	pane Pane
	host string
	port string
	user string
	db   string
}

// ConnectionCompleteMsg is sent when connection completes
type ConnectionCompleteMsg struct {
	pane    Pane
	success bool
	message string
	tables  []Table
}

// ProgressMsg is sent to update progress bar
type ProgressMsg float64

// NewModel creates a new TUI model with empty panels
func NewModel() Model {
	prog := progress.New(progress.WithDefaultGradient())
	prog.Width = 40

	logPanel := NewLogPanel()
	logPanel.AddLog(LogInfo, "v2v started - Vertica to Vertica table copy tool")

	return Model{
		leftPanel: NewTablePanel(
			"Left: [Not Connected]",
			[]Table{},
		),
		rightPanel: NewTablePanel(
			"Right: [Not Connected]",
			[]Table{},
		),
		leftConnState:  Disconnected,
		rightConnState: Disconnected,
		activePane:     LeftPane,
		viewMode:       NormalView,
		progressBar:    prog,
		logPanel:       logPanel,
		statusMsg:      "Welcome to v2v - Press F2 to connect to Vertica cluster",
	}
}

// Init initializes the model
func (m Model) Init() tea.Cmd {
	return nil
}

// Update handles messages and updates the model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Handle connection dialog if it's open
	if m.viewMode == ConnectionDialogView && m.connectionDialog != nil {
		return m.handleConnectionDialog(msg)
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Global keybindings
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "f2":
			// Open connection dialog for active pane
			m.viewMode = ConnectionDialogView
			m.connectionDialog = NewConnectionDialog(m.activePane)
			m.statusMsg = "Enter connection details..."
			return m, nil

		case "f3":
			// Disconnect active pane
			if m.activePane == LeftPane && m.leftConnState != Disconnected {
				m.leftConnState = Disconnected
				m.leftConnString = ""
				m.leftConnError = ""
				m.leftPanel.title = "Left: [Not Connected]"
				m.leftPanel.tables = []Table{}
				m.leftPanel.cursor = 0
				m.leftPanel.offset = 0
				m.statusMsg = "Left pane disconnected"
				m.logPanel.AddLog(LogInfo, "Left pane disconnected")
			} else if m.activePane == RightPane && m.rightConnState != Disconnected {
				m.rightConnState = Disconnected
				m.rightConnString = ""
				m.rightConnError = ""
				m.rightPanel.title = "Right: [Not Connected]"
				m.rightPanel.tables = []Table{}
				m.rightPanel.cursor = 0
				m.rightPanel.offset = 0
				m.statusMsg = "Right pane disconnected"
				m.logPanel.AddLog(LogInfo, "Right pane disconnected")
			}
			return m, nil

		case "f4":
			// Toggle log panel
			m.logPanel.Toggle()
			if m.logPanel.visible {
				m.statusMsg = "Log panel visible (F4 to hide, F6 to focus)"
				m.logPanel.active = false // Reset focus when toggling
			} else {
				m.statusMsg = "Log panel hidden (F4 to show)"
				m.logPanel.active = false
			}
			// Recalculate panel heights
			m = m.recalculateHeights()
			return m, nil

		case "f6":
			// Toggle focus between table panels and log panel
			if m.logPanel.visible {
				m.logPanel.active = !m.logPanel.active
				if m.logPanel.active {
					m.statusMsg = "Log panel focused (F6 to unfocus, ↑↓ to scroll)"
				} else {
					m.statusMsg = "Table panels focused"
				}
			}
			return m, nil

		case "tab":
			// Switch active pane (only if both disconnected or both connected)
			if m.activePane == LeftPane {
				m.activePane = RightPane
			} else {
				m.activePane = LeftPane
			}
			return m, nil

		case "f5":
			// Trigger copy operation (only if both panes connected)
			if m.copying || m.leftConnState != Connected || m.rightConnState != Connected {
				if m.leftConnState != Connected || m.rightConnState != Connected {
					m.statusMsg = "Both panes must be connected to copy tables"
				}
				return m, nil
			}
			return m, m.startCopy()

		case "up", "k":
			// Navigate up in active panel
			if m.logPanel.active && m.logPanel.visible {
				// Scroll up in log panel
				m.logPanel.ScrollUp()
			} else if m.activePane == LeftPane && m.leftConnState == Connected {
				m.leftPanel.MoveUp()
			} else if m.activePane == RightPane && m.rightConnState == Connected {
				m.rightPanel.MoveUp()
			}
			return m, nil

		case "down", "j":
			// Navigate down in active panel
			if m.logPanel.active && m.logPanel.visible {
				// Scroll down in log panel
				m.logPanel.ScrollDown()
			} else if m.activePane == LeftPane && m.leftConnState == Connected {
				m.leftPanel.MoveDown()
			} else if m.activePane == RightPane && m.rightConnState == Connected {
				m.rightPanel.MoveDown()
			}
			return m, nil

		case "pgup":
			// Page up in logs (faster scrolling)
			if m.logPanel.visible {
				for i := 0; i < 5; i++ {
					m.logPanel.ScrollUp()
				}
			}
			return m, nil

		case "pgdown":
			// Page down in logs (faster scrolling)
			if m.logPanel.visible {
				for i := 0; i < 5; i++ {
					m.logPanel.ScrollDown()
				}
			}
			return m, nil

		case "home":
			// Jump to oldest logs
			if m.logPanel.visible && m.logPanel.active {
				m.logPanel.ScrollToTop()
				m.statusMsg = "Jumped to oldest log"
			}
			return m, nil

		case "end":
			// Jump to newest logs
			if m.logPanel.visible && m.logPanel.active {
				m.logPanel.ScrollToBottom()
				m.statusMsg = "Jumped to latest log"
			}
			return m, nil
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.logPanel.width = msg.Width
		m = m.recalculateHeights()
		return m, nil

	case CopyCompleteMsg:
		m.copying = false
		m.statusMsg = msg.message
		if msg.success {
			// Add the copied table to the destination panel
			sourcePanel := m.leftPanel
			destPanel := m.rightPanel
			if m.activePane == RightPane {
				sourcePanel = m.rightPanel
				destPanel = m.leftPanel
			}

			if sourcePanel.cursor < len(sourcePanel.tables) {
				copiedTable := sourcePanel.tables[sourcePanel.cursor]
				// Check if table already exists in destination
				exists := false
				for _, t := range destPanel.tables {
					if t.Schema == copiedTable.Schema && t.Name == copiedTable.Name {
						exists = true
						break
					}
				}
				if !exists {
					destPanel.tables = append(destPanel.tables, copiedTable)
				}
			}
		}
		return m, nil

	case ConnectionStartMsg:
		// Connection is starting - show progress
		m.connectingPane = msg.pane
		paneName := "Left"
		if msg.pane == RightPane {
			paneName = "Right"
		}

		connStr := fmt.Sprintf("%s:%s/%s", msg.host, msg.port, msg.db)
		m.logPanel.Show()
		m.logPanel.AddLog(LogInfo, fmt.Sprintf("[%s] Connecting to %s as %s", paneName, connStr, msg.user))

		if msg.pane == LeftPane {
			m.leftConnState = Connecting
			m.leftConnString = fmt.Sprintf("vertica://%s@%s:%s/%s", msg.user, msg.host, msg.port, msg.db)
			m.leftPanel.title = fmt.Sprintf("Left: Connecting to %s:%s...", msg.host, msg.port)
		} else {
			m.rightConnState = Connecting
			m.rightConnString = fmt.Sprintf("vertica://%s@%s:%s/%s", msg.user, msg.host, msg.port, msg.db)
			m.rightPanel.title = fmt.Sprintf("Right: Connecting to %s:%s...", msg.host, msg.port)
		}
		m.statusMsg = "Connecting to Vertica..."
		m = m.recalculateHeights()
		return m, m.simulateConnection(msg)

	case ConnectionCompleteMsg:
		// Connection completed
		paneName := "Left"
		if msg.pane == RightPane {
			paneName = "Right"
		}

		if msg.success {
			if msg.pane == LeftPane {
				m.leftConnState = Connected
				m.leftPanel.title = "Left: " + m.leftConnString
				m.leftPanel.tables = msg.tables
				m.leftConnError = ""
			} else {
				m.rightConnState = Connected
				m.rightPanel.title = "Right: " + m.rightConnString
				m.rightPanel.tables = msg.tables
				m.rightConnError = ""
			}
			m.statusMsg = fmt.Sprintf("✓ Connected! Loaded %d tables", len(msg.tables))
			m.logPanel.AddLog(LogSuccess, fmt.Sprintf("[%s] Connected successfully - loaded %d tables", paneName, len(msg.tables)))
		} else {
			if msg.pane == LeftPane {
				m.leftConnState = ConnectionError
				m.leftConnError = msg.message
				m.leftPanel.title = "Left: Connection Failed"
			} else {
				m.rightConnState = ConnectionError
				m.rightConnError = msg.message
				m.rightPanel.title = "Right: Connection Failed"
			}
			m.statusMsg = "✗ Connection failed: " + msg.message
			m.logPanel.AddLog(LogError, fmt.Sprintf("[%s] Connection failed: %s", paneName, msg.message))
		}
		return m, nil

	case ProgressMsg:
		// Update progress bar
		cmd := m.progressBar.SetPercent(float64(msg))
		return m, cmd

	case CopyProgressMsg:
		// Add copy progress log
		m.logPanel.AddLog(msg.level, msg.message)
		return m, nil
	}

	return m, nil
}

// handleConnectionDialog handles input when connection dialog is open
func (m Model) handleConnectionDialog(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "ctrl+c":
			// Cancel connection dialog
			m.viewMode = NormalView
			m.connectionDialog = nil
			m.statusMsg = "Connection cancelled"
			return m, nil
		}
	}

	// Update the dialog
	var cmd tea.Cmd
	m.connectionDialog, cmd = m.connectionDialog.Update(msg)

	// Check if form was submitted
	if m.connectionDialog.IsSubmitted() {
		host, port, user, _, db := m.connectionDialog.GetConnectionParams()
		pane := m.connectionDialog.paneSide

		// Close dialog and start connection
		m.viewMode = NormalView
		m.connectionDialog = nil

		// Start connection
		return m, func() tea.Msg {
			return ConnectionStartMsg{
				pane: pane,
				host: host,
				port: port,
				user: user,
				db:   db,
			}
		}
	}

	return m, cmd
}

// View renders the UI
func (m Model) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	// Show connection dialog if active
	if m.viewMode == ConnectionDialogView && m.connectionDialog != nil {
		return m.renderWithDialog()
	}

	// Render both panels side by side
	leftActive := m.activePane == LeftPane
	rightActive := m.activePane == RightPane

	leftView := m.renderPanel(m.leftPanel, m.leftConnState, m.leftConnError, leftActive, LeftPane)
	rightView := m.renderPanel(m.rightPanel, m.rightConnState, m.rightConnError, rightActive, RightPane)

	// Combine panels horizontally
	panels := lipgloss.JoinHorizontal(lipgloss.Top, leftView, " ", rightView)

	// Log panel (if visible)
	logView := m.logPanel.Render()

	// Status and help bars
	statusBar := m.renderStatusBar()
	helpBar := m.renderHelpBar()

	// Combine everything vertically
	if m.logPanel.visible {
		return lipgloss.JoinVertical(lipgloss.Left, panels, logView, statusBar, helpBar)
	}
	return lipgloss.JoinVertical(lipgloss.Left, panels, statusBar, helpBar)
}

// renderPanel renders a panel with connection state
func (m Model) renderPanel(panel *TablePanel, connState ConnectionState, connError string, active bool, pane Pane) string {
	switch connState {
	case Disconnected:
		return panel.RenderDisconnected(active)
	case Connecting:
		if m.connectingPane == pane {
			return panel.RenderConnecting(active, m.progressBar.View())
		}
		return panel.Render(active)
	case ConnectionError:
		return panel.RenderError(active, connError)
	case Connected:
		return panel.Render(active)
	default:
		return panel.Render(active)
	}
}

// renderWithDialog renders the UI with connection dialog overlay
func (m Model) renderWithDialog() string {
	// Render background (dimmed)
	leftView := m.renderPanel(m.leftPanel, m.leftConnState, m.leftConnError, false, LeftPane)
	rightView := m.renderPanel(m.rightPanel, m.rightConnState, m.rightConnError, false, RightPane)
	panels := lipgloss.JoinHorizontal(lipgloss.Top, leftView, " ", rightView)

	// Dim the background
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	dimmedPanels := dimStyle.Render(panels)

	// Render dialog
	dialog := m.connectionDialog.View()

	// Center the dialog
	dialogHeight := strings.Count(dialog, "\n") + 1
	topPadding := (m.height - dialogHeight) / 2
	if topPadding < 0 {
		topPadding = 0
	}

	paddedDialog := strings.Repeat("\n", topPadding) + dialog

	// Combine with status bar
	statusBar := m.renderStatusBar()

	// Layer dialog over background
	return lipgloss.Place(
		m.width, m.height,
		lipgloss.Center, lipgloss.Top,
		lipgloss.JoinVertical(lipgloss.Left, dimmedPanels, statusBar),
		lipgloss.WithWhitespaceChars(""),
		lipgloss.WithWhitespaceForeground(lipgloss.Color("0")),
	) + "\n" + lipgloss.Place(
		m.width, m.height-1,
		lipgloss.Center, lipgloss.Top,
		paddedDialog,
	)
}

// renderStatusBar creates the status bar at the bottom
func (m Model) renderStatusBar() string {
	statusStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Background(lipgloss.Color("235")).
		Width(m.width).
		Padding(0, 1)

	status := m.statusMsg
	if m.copying {
		status = "⏳ Copying table..."
	}

	return statusStyle.Render(status)
}

// renderHelpBar creates the persistent help bar with context-aware commands
func (m Model) renderHelpBar() string {
	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("250")).
		Background(lipgloss.Color("236")).
		Width(m.width).
		Padding(0, 1)

	keyStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("39")).
		Bold(true)

	descStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("250"))

	// Build help based on current state
	var helps []string

	// If log panel is active, show log-specific controls
	if m.logPanel.active && m.logPanel.visible {
		helps = append(helps, keyStyle.Render("F6")+" "+descStyle.Render("Unfocus"))
		helps = append(helps, keyStyle.Render("↑↓")+" "+descStyle.Render("Scroll"))
		helps = append(helps, keyStyle.Render("PgUp/PgDn")+" "+descStyle.Render("Fast Scroll"))
		helps = append(helps, keyStyle.Render("Home/End")+" "+descStyle.Render("Top/Bottom"))
		helps = append(helps, keyStyle.Render("F4")+" "+descStyle.Render("Hide"))
		helps = append(helps, keyStyle.Render("Q")+" "+descStyle.Render("Quit"))
	} else {
		// Normal table panel controls
		// Get current pane state
		currentPaneState := m.leftConnState
		if m.activePane == RightPane {
			currentPaneState = m.rightConnState
		}

		// Connection commands
		if currentPaneState == Disconnected || currentPaneState == ConnectionError {
			helps = append(helps, keyStyle.Render("F2")+" "+descStyle.Render("Connect"))
		} else if currentPaneState == Connected {
			helps = append(helps, keyStyle.Render("F3")+" "+descStyle.Render("Disconnect"))
		}

		// Navigation
		helps = append(helps, keyStyle.Render("Tab")+" "+descStyle.Render("Switch"))

		// Copy (only if both connected)
		if m.leftConnState == Connected && m.rightConnState == Connected {
			helps = append(helps, keyStyle.Render("F5")+" "+descStyle.Render("Copy"))
		}

		// Movement (only if current pane connected)
		if currentPaneState == Connected {
			helps = append(helps, keyStyle.Render("↑↓")+" "+descStyle.Render("Navigate"))
		}

		// Log toggle
		if m.logPanel.visible {
			helps = append(helps, keyStyle.Render("F4")+" "+descStyle.Render("Hide Log"))
			helps = append(helps, keyStyle.Render("F6")+" "+descStyle.Render("Focus Log"))
		} else {
			helps = append(helps, keyStyle.Render("F4")+" "+descStyle.Render("Show Log"))
		}

		// Always show quit
		helps = append(helps, keyStyle.Render("Q")+" "+descStyle.Render("Quit"))
	}

	helpText := strings.Join(helps, " │ ")
	return helpStyle.Render(helpText)
}

// recalculateHeights recalculates panel heights based on log visibility
func (m Model) recalculateHeights() Model {
	// Divide width for two panels
	panelWidth := (m.width - 1) / 2 // -1 for the space between panels
	m.leftPanel.width = panelWidth
	m.rightPanel.width = panelWidth

	// Calculate available height
	// Status bar (1) + Help bar (1) = 2 lines
	usedHeight := 2
	if m.logPanel.visible {
		usedHeight += m.logPanel.height + 1 // +1 for border
	}

	panelHeight := m.height - usedHeight
	if panelHeight < 5 {
		panelHeight = 5
	}

	m.leftPanel.height = panelHeight
	m.rightPanel.height = panelHeight

	return m
}

// startCopy initiates a copy operation
func (m Model) startCopy() tea.Cmd {
	sourcePanel := m.leftPanel
	destPanel := m.rightPanel
	direction := "→"

	if m.activePane == RightPane {
		sourcePanel = m.rightPanel
		destPanel = m.leftPanel
		direction = "←"
	}

	if sourcePanel.cursor >= len(sourcePanel.tables) {
		return func() tea.Msg {
			return CopyCompleteMsg{
				success: false,
				message: "No table selected",
			}
		}
	}

	m.copying = true
	m.logPanel.Show()
	selectedTable := sourcePanel.tables[sourcePanel.cursor]

	sourceName := "Left"
	destName := "Right"
	if m.activePane == RightPane {
		sourceName = "Right"
		destName = "Left"
	}

	m.logPanel.AddLog(LogInfo, fmt.Sprintf("Starting table copy: %s.%s from %s to %s", selectedTable.Schema, selectedTable.Name, sourceName, destName))

	// Simulate copy operation with detailed logging
	return m.simulateCopy(selectedTable, sourceName, destName, direction, strings.Split(destPanel.title, ":")[0])
}

// simulateConnection simulates a connection to Vertica with progress updates
func (m Model) simulateConnection(msg ConnectionStartMsg) tea.Cmd {
	return func() tea.Msg {
		// Simulate connection with progress
		steps := []string{
			"Resolving hostname...",
			"Connecting to server...",
			"Authenticating...",
			"Loading metadata...",
			"Fetching tables...",
		}

		// Simulate random failure (10% chance for demo purposes)
		if time.Now().Unix()%10 == 0 {
			time.Sleep(1 * time.Second)
			return ConnectionCompleteMsg{
				pane:    msg.pane,
				success: false,
				message: "Connection refused: could not connect to " + msg.host + ":" + msg.port,
			}
		}

		// Simulate successful connection with progress
		for range steps {
			time.Sleep(300 * time.Millisecond)
			// Progress update (not displayed in mock, but ready for real implementation)
		}

		// Generate mock table data
		mockTables := []Table{
			{Schema: "public", Name: "users", RowCount: 15420, Size: "2.3 MB"},
			{Schema: "public", Name: "orders", RowCount: 89532, Size: "15.7 MB"},
			{Schema: "public", Name: "products", RowCount: 3421, Size: "1.1 MB"},
			{Schema: "analytics", Name: "events", RowCount: 1234567, Size: "234.5 MB"},
			{Schema: "analytics", Name: "sessions", RowCount: 456789, Size: "89.2 MB"},
			{Schema: "staging", Name: "raw_data", RowCount: 789012, Size: "156.8 MB"},
			{Schema: "public", Name: "customers", RowCount: 5432, Size: "987 KB"},
			{Schema: "public", Name: "invoices", RowCount: 23456, Size: "4.2 MB"},
		}

		return ConnectionCompleteMsg{
			pane:    msg.pane,
			success: true,
			message: "Connected successfully",
			tables:  mockTables,
		}
	}
}

// CopyProgressMsg is sent to update copy progress with logs
type CopyProgressMsg struct {
	level   LogLevel
	message string
}

// simulateCopy simulates a table copy operation with detailed step-by-step logging
func (m Model) simulateCopy(table Table, sourceName, destName, direction, destPaneName string) tea.Cmd {
	return tea.Batch(
		func() tea.Msg {
			time.Sleep(200 * time.Millisecond)
			return CopyProgressMsg{
				level:   LogInfo,
				message: fmt.Sprintf("[%s → %s] Step 1/4: Checking if table %s.%s exists on target", sourceName, destName, table.Schema, table.Name),
			}
		},
		func() tea.Msg {
			time.Sleep(500 * time.Millisecond)
			return CopyProgressMsg{
				level:   LogInfo,
				message: fmt.Sprintf("[%s → %s] Step 2/4: Retrieving DDL for %s.%s from source", sourceName, destName, table.Schema, table.Name),
			}
		},
		func() tea.Msg {
			time.Sleep(800 * time.Millisecond)
			ddl := fmt.Sprintf("CREATE TABLE %s.%s (...)", table.Schema, table.Name)
			return CopyProgressMsg{
				level:   LogDebug,
				message: fmt.Sprintf("[%s → %s] DDL: %s", sourceName, destName, ddl),
			}
		},
		func() tea.Msg {
			time.Sleep(1100 * time.Millisecond)
			return CopyProgressMsg{
				level:   LogInfo,
				message: fmt.Sprintf("[%s → %s] Step 3/4: Creating table %s.%s on target", sourceName, destName, table.Schema, table.Name),
			}
		},
		func() tea.Msg {
			time.Sleep(1400 * time.Millisecond)
			return CopyProgressMsg{
				level:   LogSuccess,
				message: fmt.Sprintf("[%s → %s] Table created successfully", sourceName, destName),
			}
		},
		func() tea.Msg {
			time.Sleep(1700 * time.Millisecond)
			return CopyProgressMsg{
				level:   LogInfo,
				message: fmt.Sprintf("[%s → %s] Step 4/4: Copying %d rows using COPY TO STDOUT / FROM STDIN", sourceName, destName, table.RowCount),
			}
		},
		func() tea.Msg {
			time.Sleep(2500 * time.Millisecond)
			return CopyProgressMsg{
				level:   LogSuccess,
				message: fmt.Sprintf("[%s → %s] Data transfer complete: %d rows, %s", sourceName, destName, table.RowCount, table.Size),
			}
		},
		func() tea.Msg {
			time.Sleep(2700 * time.Millisecond)
			message := fmt.Sprintf(
				"✓ Copied %s.%s %s %s",
				table.Schema,
				table.Name,
				direction,
				destPaneName,
			)
			return CopyCompleteMsg{
				success: true,
				message: message,
			}
		},
	)
}
