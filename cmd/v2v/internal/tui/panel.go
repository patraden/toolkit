package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/patraden/toolkit/cmd/v2v/internal/domain"
	"github.com/patraden/toolkit/cmd/v2v/internal/utils"
)

const (
	tablePanelWidth   = 60
	tablePanelHeight  = 20
	borderColor       = lipgloss.Color("240")
	headerColor       = lipgloss.Color("39")
	activeBorderColor = lipgloss.Color("42")
	selectedBg        = lipgloss.Color("237")
)

// TablePanel represents one side of the dual-pane interface
type TablePanel struct {
	title  string
	tables []domain.Table
	cursor int
	offset int
	width  int // p.width is the total panel width including borders
	height int // p.height is the total panel height including borders
}

// NewTablePanel creates a new table panel
func NewTablePanel(title string, tables []domain.Table) *TablePanel {
	return &TablePanel{
		title:  title,
		tables: tables,
		cursor: 0,
		offset: 0,
		width:  tablePanelWidth,
		height: tablePanelHeight,
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
	borderColor := borderColor
	if active {
		borderColor = activeBorderColor
	}

	// p.width is the total panel width including borders
	// p.height is the total panel height including borders
	// Content area width = p.width - 2 (for left and right border)
	// Content area height = p.height - 2 (for top and bottom border)
	contentWidth := p.width - 2
	contentHeight := p.height - 2

	// Styles
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
		Foreground(headerColor).
		Width(contentWidth)

	cols := fmt.Sprintf("%-18s %-15s %10s %10s", "Schema", "Table", "Rows", "Size")
	colHeader := columnHeader.Render(cols)
	separator := strings.Repeat("─", contentWidth)

	// Table rows
	var rows []string
	// contentHeight already excludes borders, now subtract header, column header, separator
	visibleHeight := contentHeight - 3

	endIdx := min(p.offset+visibleHeight, len(p.tables))

	for i := p.offset; i < endIdx; i++ {
		table := p.tables[i]

		// Format the row
		rowText := fmt.Sprintf("%-18s %-15s %10d %10d",
			utils.Truncate(table.Schema, 18),
			utils.Truncate(table.Name, 15),
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
		strings.Join(rows, utils.EOL),
	)

	return panelStyle.Render(content)
}

func (p *TablePanel) SetWidth(w int) {
	p.width = w
}

func (p *TablePanel) SetHeight(h int) {
	p.height = h
}
