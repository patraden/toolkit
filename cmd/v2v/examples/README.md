# v2v - Vertica to Vertica Table Copy Tool

A production-ready Far Manager-style TUI (Text User Interface) for copying tables between Vertica clusters.

## Overview

**v2v** provides an interactive dual-pane interface for viewing and copying tables between two Vertica database clusters. Built with [Bubble Tea](https://github.com/charmbracelet/bubbletea), it offers a familiar file-manager-like experience optimized for database operations.

## Current Status

**🎨 Prototype Complete** - Full TUI with mocked backend, ready for Vertica integration.

The interface, navigation, logging, and all user interactions are production-ready. Only the database operations are simulated.

## Quick Start

```bash
# Build
cd /Users/d.patrakhin/go/toolkit/go
make build-v2v

# Run
make run-v2v

# Or directly
./bin/v2v

# Help
./bin/v2v --help
```

## Architecture

```
cmd/v2v/
├── main.go                    # Entry point, CLI args (78 lines)
└── tui/
    ├── model.go               # Core state & logic (803 lines)
    ├── table_panel.go         # Table display (360 lines)
    ├── connection_dialog.go   # Connection form (228 lines)
    └── log_panel.go           # Operation logs (295 lines)
```

### Key Components

#### 1. Model (model.go)
- **Bubble Tea Model**: Main application state and update logic
- **State Management**: Connection states, view modes, active panels
- **Message Handlers**: Connection, copy, progress updates
- **Render Logic**: Layout composition, help bar generation

#### 2. Table Panel (table_panel.go)
- **Dual Panes**: Left (source) and Right (target)
- **Multiple States**: Disconnected, Connecting, Connected, Error
- **Navigation**: Cursor movement, scrolling
- **Visual Feedback**: Active highlighting, borders, metadata display

#### 3. Connection Dialog (connection_dialog.go)
- **5-Field Form**: Host, Port, User, Password, Database
- **Smart Navigation**: Tab, Shift+Tab, Enter, arrows
- **Password Masking**: Hidden input with bullets
- **Default Values**: Auto-fill on empty fields

#### 4. Log Panel (log_panel.go)
- **Scrollable History**: 100 entries, 8 lines visible
- **Log Levels**: Info, Success, Warning, Error, Debug
- **Focus Mode**: Dedicated navigation when active
- **Scroll Indicators**: Position feedback

## Features

### Connection Management
- **F2**: Open connection dialog for active pane
- **F3**: Disconnect active pane
- **Auto-show logs**: Logs appear during connection
- **Error handling**: Clear error messages with retry option

### Table Operations
- **Dual-pane view**: Side-by-side source and target
- **Metadata display**: Schema, table name, row count, size
- **Navigation**: Arrow keys or vim-style (k/j)
- **Tab switching**: Move between left and right panes
- **F5 Copy**: Simulated 4-step copy process

### Operation Logging
- **F4**: Toggle log panel visibility
- **F6**: Focus/unfocus for scrolling
- **Color-coded**: Different colors for log levels
- **Timestamped**: Every entry shows HH:MM:SS
- **Scrollable**: Navigate log history

### Copy Workflow (Simulated)
1. **Step 1**: Check if table exists on target (300ms)
2. **Step 2**: Retrieve DDL from source (400ms)
3. **Step 3**: Create table on target (500ms)
4. **Step 4**: Copy data via STDOUT/STDIN (1500ms)

Each step is logged with timestamps and status.

## Keyboard Controls

### Global
- **F2**: Connect active pane to Vertica
- **F3**: Disconnect active pane
- **F4**: Toggle operation log panel
- **F6**: Focus/unfocus log panel
- **Tab**: Switch between left/right panes
- **q / Ctrl+C**: Quit application

### Table Navigation
- **↑/↓** or **k/j**: Move cursor up/down
- **F5**: Copy selected table to other pane

### Log Navigation (when focused)
- **↑/↓**: Scroll one line
- **PgUp/PgDn**: Scroll 5 lines
- **Home**: Jump to oldest log
- **End**: Jump to latest log

### Connection Dialog
- **Tab / ↓**: Next field
- **Shift+Tab / ↑**: Previous field
- **Enter**: Next field or submit (from last field)
- **Esc**: Cancel dialog

## Visual Design

### Color Scheme
- **Green (42)**: Active elements, success messages
- **Blue (39)**: Headers, info messages
- **Yellow (220)**: Connecting state, warnings
- **Red (196)**: Errors, failed operations
- **Gray (240-244)**: Borders, debug info

### Layout
```
┌─────────────────────────┬─────────────────────────┐
│ Left Panel              │ Right Panel             │
│ [tables or state]       │ [tables or state]       │
└─────────────────────────┴─────────────────────────┘
──────────────────────────────────────────────────────
Operation Log (if visible)
[timestamped log entries with icons and colors]
──────────────────────────────────────────────────────
Status: Brief message about current operation
──────────────────────────────────────────────────────
F2 Connect | Tab Switch | F5 Copy | F4 Log | Q Quit
```

### Context-Aware Help
The bottom help bar changes based on:
- Connection state (connected vs disconnected)
- Active panel (tables vs logs)
- Available operations (can copy if both connected)

## State Management

### Connection States
```go
type ConnectionState int
const (
    Disconnected     // Initial state
    Connecting       // F2 pressed, dialog submitted
    Connected        // Tables loaded and displayed
    ConnectionError  // Failed with error message
)
```

### View Modes
```go
type ViewMode int
const (
    NormalView            // Default dual-pane view
    ConnectionDialogView  // F2 dialog overlay
)
```

### Panel Focus
- Table panels: Navigate with Tab
- Log panel: Focus with F6 (green border)
- Connection dialog: Automatic focus on open

## Technical Details

### Height Calculations
```go
// Panel total height includes borders
panelHeight := terminalHeight - statusBar(1) - helpBar(1) - logPanel(optional)

// Content height (what lipgloss Height() expects)
contentHeight := panelHeight - 2  // Subtract top/bottom borders

// Visible rows in table
visibleRows := contentHeight - 3  // Subtract title, columns, separator
```

### Log Scrolling
- `offset = 0`: Show latest logs (bottom)
- `offset > 0`: Scrolled up by N entries
- Auto-scroll to bottom on new logs (unless scrolled)

### Width Distribution
```go
// Total width split between panels
panelWidth := (terminalWidth - 1) / 2  // -1 for spacing
contentWidth := panelWidth - 2          // -2 for left/right borders
```

## Dependencies

```go
require (
    github.com/charmbracelet/bubbletea    // TUI framework
    github.com/charmbracelet/lipgloss     // Styling & layout
    github.com/charmbracelet/bubbles      // textinput, progress
    github.com/vertica/vertica-sql-go     // Vertica driver (future)
)
```

## Future Implementation

### Phase 1: Real Vertica Connection

**Replace `simulateConnection()` in model.go:**
```go
func (m Model) connectToVertica(msg ConnectionStartMsg) tea.Cmd {
    return func() tea.Msg {
        // Use existing pkg/vertica wrapper
        connParams := vertica.NewConnString(
            msg.host,
            parseInt(msg.port),
            msg.user,
            msg.password,
            msg.db,
        )
        
        db, err := vertica.NewDB(connParams)
        if err != nil {
            return ConnectionCompleteMsg{
                pane: msg.pane,
                success: false,
                message: err.Error(),
            }
        }
        
        // Query real tables
        tables, err := queryVerticaTables(db)
        if err != nil {
            return ConnectionCompleteMsg{
                pane: msg.pane,
                success: false,
                message: err.Error(),
            }
        }
        
        return ConnectionCompleteMsg{
            pane: msg.pane,
            success: true,
            tables: tables,
        }
    }
}
```

**Query tables from v_catalog:**
```go
func queryVerticaTables(db *vertica.DB) ([]Table, error) {
    query := `
        SELECT 
            table_schema,
            table_name,
            SUM(row_count) as row_count,
            SUM(used_bytes) as used_bytes
        FROM v_catalog.tables
        WHERE NOT is_system_table
        GROUP BY table_schema, table_name
        ORDER BY table_schema, table_name
    `
    
    rows, err := db.QueryContext(ctx, query)
    // Parse and return []Table
}
```

### Phase 2: Real Table Copy

**Replace `simulateCopy()` in model.go:**
```go
func (m Model) copyTable(table Table, sourceDB, destDB *vertica.DB) tea.Cmd {
    return func() tea.Msg {
        // Step 1: Check target
        exists, err := tableExists(destDB, table.Schema, table.Name)
        
        // Step 2: Get DDL
        ddl, err := getTableDDL(sourceDB, table.Schema, table.Name)
        
        // Step 3: Create table
        err = destDB.ExecContext(ctx, ddl)
        
        // Step 4: Copy data
        // Option A: COPY pipeline
        err = copyViaPipe(sourceDB, destDB, table)
        
        // Option B: EXPORT/IMPORT
        err = copyViaExport(sourceDB, destDB, table)
        
        return CopyCompleteMsg{success: true}
    }
}
```

**COPY pipeline implementation:**
```go
func copyViaPipe(source, dest *vertica.DB, table Table) error {
    // Start COPY TO STDOUT on source
    // Start COPY FROM STDIN on dest
    // Pipe data between them
    // Uses existing cmd/v2v/../../sql/vertica/vsql/ scripts as reference
}
```

### Phase 3: Production Features
- [ ] Real connection management (store DB connections)
- [ ] Connection pooling
- [ ] Progress tracking for large tables
- [ ] Parallel copy support
- [ ] DDL transformation options
- [ ] Column mapping
- [ ] Filter/transform data
- [ ] Connection profiles/config file
- [ ] Export logs to file

## Testing

### Manual Testing Checklist
- [ ] F2 connects both panes successfully
- [ ] F3 disconnects cleanly
- [ ] Tab switches between panes
- [ ] F5 copies table (mocked)
- [ ] F4 toggles log panel
- [ ] F6 focuses/unfocuses logs
- [ ] Arrow keys navigate tables when panels focused
- [ ] Arrow keys scroll logs when log focused
- [ ] PgUp/PgDn fast-scrolls logs
- [ ] Home/End jumps in logs
- [ ] Terminal resize handled gracefully
- [ ] All borders visible in all states
- [ ] Progress bar centered during connection
- [ ] Connection errors displayed clearly

### Unit Tests (Future)
```bash
go test ./cmd/v2v/...
```

## Troubleshooting

### Borders Not Visible
- Check terminal size (minimum ~80x24)
- Ensure `recalculateHeights()` called on resize
- Verify `contentHeight = height - 2` used consistently

### Log Panel Issues
- F6 to focus/unfocus
- Check `logPanel.active` for focus state
- Verify `logPanel.offset` for scroll position

### Connection Dialog Not Appearing
- Check `viewMode == ConnectionDialogView`
- Ensure `renderWithDialog()` used in View()
- Verify Esc cancels properly

## License

Copyright 2025 The Toolkit Authors  
Licensed under Apache License 2.0

## Contributing

This tool is part of the [toolkit](https://github.com/patraden/toolkit) repository.

## Support

For questions or issues, refer to the inline documentation in:
- `model.go`: State management and business logic
- `table_panel.go`: Panel rendering details
- `connection_dialog.go`: Form implementation
- `log_panel.go`: Log scrolling logic

