package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/patraden/toolkit/cmd/v2v/internal/domain"
)

const minPanelHeight = 5

// Pane represents which pane is active
type Pane int

const (
	LeftPane Pane = iota
	RightPane
)

type Model struct {
	leftPanel  *TablePanel
	rightPanel *TablePanel
	helpBar    *HelpBar
	activePane Pane
	width      int
	height     int
}

// NewModel creates a new TUI model with empty panels
func NewModel() Model {
	m := Model{
		leftPanel:  NewTablePanel("Left: [Not Connected]", []domain.Table{}),
		rightPanel: NewTablePanel("Right: [Not Connected]", []domain.Table{}),
	}

	helpBar := NewHelpBar(m.width, 1)
	m.helpBar = helpBar
	return m
}

// Init initializes the model
func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.resizeElements()
		m.helpBar.SetWidth(m.width)

		return m, nil
	}

	// Return the updated model to the Bubble Tea runtime for processing.
	// Note that we're not returning a command.
	return m, nil
}

func (m Model) View() string {
	if m.width <= 0 || m.height <= 0 {
		return "Loading..."
	}

	leftActive := m.activePane == LeftPane
	rightActive := m.activePane == RightPane
	leftView := m.leftPanel.Render(leftActive)
	rightView := m.rightPanel.Render(rightActive)
	// Combine panels horizontally
	panels := lipgloss.JoinHorizontal(lipgloss.Top, leftView, " ", rightView)

	helpBar := m.helpBar.Render()
	content := lipgloss.JoinVertical(lipgloss.Top, panels, helpBar)

	// Place content in a full width×height box so every terminal cell is
	// overwritten on redraw (avoids resize artifacts from previous size).
	return lipgloss.Place(m.width, m.height, lipgloss.Left, lipgloss.Top, content,
		lipgloss.WithWhitespaceChars(" "),
		lipgloss.WithWhitespaceForeground(lipgloss.Color("0")),
	)
}

func (m Model) resizeElements() Model {
	// Divide width for two panels
	// -1 for the space between panels
	panelWidth := (m.width - 1) / 2

	m.leftPanel.SetWidth(panelWidth)
	m.rightPanel.SetWidth(panelWidth)

	// Calculate available height
	// Status bar (1) + Help bar (1) = 2 lines
	usedHeight := 2
	panelHeight := m.height - usedHeight
	panelHeight = max(panelHeight, minPanelHeight)

	m.leftPanel.SetHeight(panelHeight)
	m.rightPanel.SetHeight(panelHeight)
	m.helpBar.SetWidth(panelWidth * 2)
	m.helpBar.SetHeight(1)

	return m
}
