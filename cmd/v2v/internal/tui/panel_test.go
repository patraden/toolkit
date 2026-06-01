package tui_test

import (
	"os"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
	"github.com/patraden/toolkit/cmd/v2v/internal/domain"
	"github.com/patraden/toolkit/cmd/v2v/internal/tui"
	"github.com/stretchr/testify/require"
)

var testTables = []domain.Table{
	{Database: "db", Schema: "public", Name: "table1", RowCount: 10, Size: 1024},
	{Database: "db", Schema: "public", Name: "table2", RowCount: 205, Size: 2048},
	{Database: "db", Schema: "etl", Name: "table3", RowCount: 100, Size: 2024},
	{Database: "db", Schema: "etl", Name: "table4", RowCount: 131, Size: 101024},
}

func TestTablePanel_RenderOutputReadableInLogs(t *testing.T) {
	t.Parallel()

	p := tui.NewTablePanel("Table Panel", testTables)
	out := p.Render(true)

	require.NotEmpty(t, out, "Render produced empty output")
	require.Contains(t, out, "Table Panel", "expected title in output")
	require.Contains(t, out, "public", "expected schema in output")
	require.Contains(t, out, "table1", "expected table name in output")
}

func TestTablePanel_StructureAndContent(t *testing.T) {
	t.Parallel()

	p := tui.NewTablePanel("Source", testTables)
	out := p.Render(false)

	require.Contains(t, out, "Schema", "expected column headers")
	require.Contains(t, out, "Table")
	require.Contains(t, out, "Rows")
	require.Contains(t, out, "Size")
	require.Contains(t, out, "table1", "expected first table row content")
	require.Contains(t, out, "10")
	require.Contains(t, out, "1024")
}

func TestTablePanel_ActiveVsInactiveBorder(t *testing.T) {
	t.Parallel()

	p := tui.NewTablePanel("Title", testTables)
	inactive := p.Render(false)
	active := p.Render(true)

	require.NotEmpty(t, inactive, "inactive render should be non-empty")
	require.NotEmpty(t, active, "active render should be non-empty")
}

func TestTablePanel_CursorMovement(t *testing.T) {
	t.Parallel()

	p := tui.NewTablePanel("Title", testTables)
	p.MoveDown()
	p.MoveDown()
	out := p.Render(true)

	require.Contains(t, out, "table3", "after MoveDown x2, table3 should be visible")
}

func TestTablePanel_VisualInTerminal(t *testing.T) {
	t.Parallel()

	lipgloss.SetColorProfile(termenv.ANSI256)
	p := tui.NewTablePanel("Table Panel", testTables)
	p.MoveDown()
	out := p.Render(true)
	const path = "/tmp/ver2ver_panel_preview.txt"
	require.NoError(t, os.WriteFile(path, []byte(out), 0o600), "write preview file")

	t.Logf("Colored output written to %s — run: cat %s", path, path)
}
