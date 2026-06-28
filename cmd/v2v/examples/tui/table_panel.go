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

	"github.com/charmbracelet/lipgloss"
)

// Table represents a database table with metadata
type Table struct {
	Schema   string
	Name     string
	RowCount int
	Size     string
}

// TablePanel represents one side of the dual-pane interface
type TablePanel struct {
	title  string
	tables []Table
	cursor int
	offset int
	width  int
	height int
}

// NewTablePanel creates a new table panel
func NewTablePanel(title string, tables []Table) *TablePanel {
	return &TablePanel{
		title:  title,
		tables: tables,
		cursor: 0,
		offset: 0,
		width:  40,
		height: 20,
	}
}

// MoveUp moves the cursor up
func (p *TablePanel) MoveUp() {
	if p.cursor > 0 {
		p.cursor--
		if p.cursor < p.offset {
			p.offset = p.cursor
		}
	}
}

// MoveDown moves the cursor down
func (p *TablePanel) MoveDown() {
	if p.cursor < len(p.tables)-1 {
		p.cursor++
		// Calculate visible height: p.height - 2 (borders) - 3 (header, columns, separator)
		visibleHeight := p.height - 5
		if p.cursor >= p.offset+visibleHeight {
			p.offset++
		}
	}
}

// Render renders the table panel
func (p *TablePanel) Render(active bool) string {
	// Styles
	borderColor := lipgloss.Color("240")
	activeBorderColor := lipgloss.Color("42")
	headerColor := lipgloss.Color("39")
	selectedBg := lipgloss.Color("237")

	if active {
		borderColor = activeBorderColor
	}

	// p.width is the total panel width including borders
	// p.height is the total panel height including borders
	// Content area width = p.width - 2 (for left and right border)
	// Content area height = p.height - 2 (for top and bottom border)
	contentWidth := p.width - 2
	contentHeight := p.height - 2

	panelStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Width(contentWidth).
		Height(contentHeight)

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(headerColor).
		Width(contentWidth).
		Align(lipgloss.Center)

	// Header
	header := titleStyle.Render(p.title)

	// Column headers
	columnHeader := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("242")).
		Width(contentWidth)

	cols := fmt.Sprintf("%-18s %-15s %10s %10s", "Schema", "Table", "Rows", "Size")
	colHeader := columnHeader.Render(cols)
	separator := strings.Repeat("─", contentWidth)

	// Table rows
	var rows []string
	// contentHeight already excludes borders, now subtract header, column header, separator
	visibleHeight := contentHeight - 3

	endIdx := p.offset + visibleHeight
	if endIdx > len(p.tables) {
		endIdx = len(p.tables)
	}

	for i := p.offset; i < endIdx; i++ {
		table := p.tables[i]

		// Format the row
		rowText := fmt.Sprintf("%-18s %-15s %10d %10s",
			truncate(table.Schema, 18),
			truncate(table.Name, 15),
			table.RowCount,
			table.Size,
		)

		// Apply selection style
		if i == p.cursor {
			rowStyle := lipgloss.NewStyle().
				Background(selectedBg).
				Width(contentWidth)
			if active {
				rowStyle = rowStyle.Bold(true).Foreground(lipgloss.Color("15"))
			}
			rowText = rowStyle.Render(rowText)
		}

		rows = append(rows, rowText)
	}

	// Show scroll indicator if needed
	scrollInfo := ""
	if len(p.tables) > visibleHeight {
		scrollInfo = fmt.Sprintf(" (%d/%d)", p.cursor+1, len(p.tables))
	}

	// Fill empty space
	for len(rows) < visibleHeight {
		rows = append(rows, strings.Repeat(" ", contentWidth))
	}

	// Combine all elements
	content := lipgloss.JoinVertical(
		lipgloss.Left,
		header+scrollInfo,
		colHeader,
		separator,
		strings.Join(rows, "\n"),
	)

	return panelStyle.Render(content)
}

// truncate truncates a string to a maximum length
func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	if max <= 3 {
		return s[:max]
	}
	return s[:max-3] + "..."
}

// RenderDisconnected renders the panel in disconnected state
func (p *TablePanel) RenderDisconnected(active bool) string {
	borderColor := lipgloss.Color("240")
	if active {
		borderColor = lipgloss.Color("42")
	}

	contentWidth := p.width - 2
	contentHeight := p.height - 2 // Subtract borders

	panelStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Width(contentWidth).
		Height(contentHeight)

	messageStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Italic(true).
		Align(lipgloss.Center).
		Width(contentWidth)

	message := messageStyle.Render(
		"Not Connected\n\n" +
			"Press F2 to connect\n" +
			"to Vertica cluster",
	)

	// Center the message vertically
	messageHeight := 4 // 4 lines of text
	topPadding := (contentHeight - messageHeight) / 2
	if topPadding < 0 {
		topPadding = 0
	}

	centeredMessage := strings.Repeat("\n", topPadding) + message

	return panelStyle.Render(centeredMessage)
}

// RenderConnecting renders the panel in connecting state
func (p *TablePanel) RenderConnecting(active bool, progressBar string) string {
	borderColor := lipgloss.Color("240")
	connectingColor := lipgloss.Color("220")
	if active {
		borderColor = lipgloss.Color("42")
	}

	contentWidth := p.width - 2
	contentHeight := p.height - 2 // Subtract borders

	panelStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Width(contentWidth).
		Height(contentHeight)

	messageStyle := lipgloss.NewStyle().
		Foreground(connectingColor).
		Bold(true).
		Align(lipgloss.Center).
		Width(contentWidth)

	subStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Italic(true).
		Align(lipgloss.Center).
		Width(contentWidth)

	progressStyle := lipgloss.NewStyle().
		Align(lipgloss.Center).
		Width(contentWidth)

	message := messageStyle.Render("⟳ Connecting...") + "\n\n" +
		subStyle.Render("Please wait") + "\n\n" +
		progressStyle.Render(progressBar)

	// Center the message vertically
	messageHeight := 5 // 5 lines total
	topPadding := (contentHeight - messageHeight) / 2
	if topPadding < 0 {
		topPadding = 0
	}

	centeredMessage := strings.Repeat("\n", topPadding) + message

	return panelStyle.Render(centeredMessage)
}

// RenderError renders the panel in error state
func (p *TablePanel) RenderError(active bool, errorMsg string) string {
	borderColor := lipgloss.Color("240")
	errorColor := lipgloss.Color("196")
	if active {
		borderColor = lipgloss.Color("42")
	}

	contentWidth := p.width - 2
	contentHeight := p.height - 2 // Subtract borders

	panelStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Width(contentWidth).
		Height(contentHeight)

	titleStyle := lipgloss.NewStyle().
		Foreground(errorColor).
		Bold(true).
		Align(lipgloss.Center).
		Width(contentWidth)

	errorStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Align(lipgloss.Center).
		Width(contentWidth - 4)

	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Italic(true).
		Align(lipgloss.Center).
		Width(contentWidth)

	// Wrap error message if too long
	wrappedError := wrapText(errorMsg, contentWidth-4)

	message := titleStyle.Render("✗ Connection Failed") + "\n\n" +
		errorStyle.Render(wrappedError) + "\n\n" +
		helpStyle.Render("Press F2 to try again")

	// Center the message vertically
	// Count lines in message
	messageHeight := 5 + strings.Count(wrappedError, "\n")
	topPadding := (contentHeight - messageHeight) / 2
	if topPadding < 0 {
		topPadding = 0
	}

	centeredMessage := strings.Repeat("\n", topPadding) + message

	return panelStyle.Render(centeredMessage)
}

// wrapText wraps text to fit within maxWidth
func wrapText(text string, maxWidth int) string {
	if len(text) <= maxWidth {
		return text
	}

	var result strings.Builder
	words := strings.Fields(text)
	lineLen := 0

	for i, word := range words {
		wordLen := len(word)
		if lineLen+wordLen+1 > maxWidth {
			result.WriteString("\n")
			result.WriteString(word)
			lineLen = wordLen
		} else {
			if i > 0 {
				result.WriteString(" ")
				lineLen++
			}
			result.WriteString(word)
			lineLen += wordLen
		}
	}

	return result.String()
}
