# Clean Architecture Examples

## Example: Implementing "List Tables" Feature

### Step 1: Domain Layer (No Dependencies)

```go
// internal/domain/table.go
package domain

import "fmt"

type Table struct {
    Database string
    Schema   string
    Name     string
    RowCount int
    Size     int64
}

func (t Table) FullName() string {
    return fmt.Sprintf("%s.%s.%s", t.Database, t.Schema, t.Name)
}

func (t Table) SizeMB() string {
    return fmt.Sprintf("%d MB", t.Size/1024/1024)
}
```

### Step 2: Repository Layer (Interface + Implementation)

```go
// internal/repository/table_repository.go
package repository

import (
    "context"
    "github.com/patraden/toolkit/cmd/ver2ver/internal/domain"
)

// TableRepository defines data access interface
type TableRepository interface {
    ListTables(ctx context.Context, database string) ([]domain.Table, error)
    GetTableDDL(ctx context.Context, table domain.Table) (string, error)
    CreateTable(ctx context.Context, ddl string) error
    CopyData(ctx context.Context, source, target domain.Table) error
}
```

```go
// internal/repository/vertica_repository.go
package repository

import (
    "context"
    "fmt"
    
    "github.com/patraden/toolkit/cmd/ver2ver/internal/domain"
    "github.com/patraden/toolkit/pkg/vertica"
)

type VerticaRepository struct {
    db *vertica.DB
}

func NewVerticaRepository(db *vertica.DB) *VerticaRepository {
    return &VerticaRepository{db: db}
}

func (r *VerticaRepository) ListTables(ctx context.Context, database string) ([]domain.Table, error) {
    query := `
        SELECT 
            table_schema,
            table_name,
            row_count,
            used_bytes
        FROM v_catalog.tables
        WHERE table_schema NOT IN ('v_catalog', 'v_monitor', 'v_internal')
        ORDER BY table_schema, table_name
    `
    
    rows, err := r.db.QueryContext(ctx, query)
    if err != nil {
        return nil, fmt.Errorf("failed to query tables: %w", err)
    }
    defer rows.Close()
    
    var tables []domain.Table
    for rows.Next() {
        var t domain.Table
        t.Database = database
        if err := rows.Scan(&t.Schema, &t.Name, &t.RowCount, &t.Size); err != nil {
            return nil, fmt.Errorf("failed to scan table: %w", err)
        }
        tables = append(tables, t)
    }
    
    return tables, rows.Err()
}

// Implement other methods...
```

### Step 3: UseCase Layer (Business Logic)

```go
// internal/usecase/list_tables.go
package usecase

import (
    "context"
    "fmt"
    
    "github.com/patraden/toolkit/cmd/ver2ver/internal/domain"
    "github.com/patraden/toolkit/cmd/ver2ver/internal/repository"
)

type ListTablesUseCase struct {
    repo repository.TableRepository
}

func NewListTablesUseCase(repo repository.TableRepository) *ListTablesUseCase {
    return &ListTablesUseCase{repo: repo}
}

func (uc *ListTablesUseCase) Execute(ctx context.Context, database string) ([]domain.Table, error) {
    // Business logic: validate, filter, transform
    if database == "" {
        return nil, fmt.Errorf("database name is required")
    }
    
    tables, err := uc.repo.ListTables(ctx, database)
    if err != nil {
        return nil, fmt.Errorf("failed to list tables: %w", err)
    }
    
    // Additional business rules here (e.g., filter by size, permissions, etc.)
    
    return tables, nil
}
```

```go
// internal/usecase/copy_table.go
package usecase

import (
    "context"
    "fmt"
    
    "github.com/patraden/toolkit/cmd/ver2ver/internal/domain"
    "github.com/patraden/toolkit/cmd/ver2ver/internal/repository"
)

type CopyTableUseCase struct {
    sourceRepo repository.TableRepository
    targetRepo repository.TableRepository
}

func NewCopyTableUseCase(sourceRepo, targetRepo repository.TableRepository) *CopyTableUseCase {
    return &CopyTableUseCase{
        sourceRepo: sourceRepo,
        targetRepo: targetRepo,
    }
}

type CopyProgress struct {
    Step    string
    Message string
}

func (uc *CopyTableUseCase) Execute(
    ctx context.Context,
    table domain.Table,
    progressChan chan<- CopyProgress,
) error {
    // Step 1: Get DDL from source
    progressChan <- CopyProgress{Step: "ddl", Message: "Getting table DDL from source..."}
    ddl, err := uc.sourceRepo.GetTableDDL(ctx, table)
    if err != nil {
        return fmt.Errorf("failed to get DDL: %w", err)
    }
    
    // Step 2: Create table in target
    progressChan <- CopyProgress{Step: "create", Message: "Creating table in target..."}
    if err := uc.targetRepo.CreateTable(ctx, ddl); err != nil {
        return fmt.Errorf("failed to create table: %w", err)
    }
    
    // Step 3: Copy data
    progressChan <- CopyProgress{Step: "copy", Message: "Copying data..."}
    if err := uc.sourceRepo.CopyData(ctx, table, table); err != nil {
        return fmt.Errorf("failed to copy data: %w", err)
    }
    
    progressChan <- CopyProgress{Step: "done", Message: "Copy completed successfully"}
    return nil
}
```

### Step 4: TUI Layer (User Interface)

```go
// internal/tui/model.go (simplified excerpt)
package tui

import (
    tea "github.com/charmbracelet/bubbletea"
    "github.com/patraden/toolkit/cmd/ver2ver/internal/domain"
    "github.com/patraden/toolkit/cmd/ver2ver/internal/usecase"
)

type Model struct {
    // Dependencies (injected from main)
    listTablesUseCase *usecase.ListTablesUseCase
    copyTableUseCase  *usecase.CopyTableUseCase
    
    // State
    leftTables  []domain.Table
    rightTables []domain.Table
    selectedIdx int
    logs        []string
}

// Message types
type tablesLoadedMsg struct {
    side   string
    tables []domain.Table
}

type copyProgressMsg struct {
    step    string
    message string
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.String() {
        case "F2":
            // Connect and load tables
            return m, m.loadTablesCmd("left", "mydb")
            
        case "F5":
            // Copy selected table
            if m.selectedIdx >= 0 && m.selectedIdx < len(m.leftTables) {
                table := m.leftTables[m.selectedIdx]
                return m, m.copyTableCmd(table)
            }
        }
        
    case tablesLoadedMsg:
        if msg.side == "left" {
            m.leftTables = msg.tables
        } else {
            m.rightTables = msg.tables
        }
        
    case copyProgressMsg:
        m.logs = append(m.logs, msg.message)
    }
    
    return m, nil
}

// Commands that call use cases
func (m Model) loadTablesCmd(side, database string) tea.Cmd {
    return func() tea.Msg {
        tables, err := m.listTablesUseCase.Execute(context.Background(), database)
        if err != nil {
            return errorMsg{err}
        }
        return tablesLoadedMsg{side: side, tables: tables}
    }
}

func (m Model) copyTableCmd(table domain.Table) tea.Cmd {
    return func() tea.Msg {
        progressChan := make(chan usecase.CopyProgress)
        
        go func() {
            err := m.copyTableUseCase.Execute(context.Background(), table, progressChan)
            if err != nil {
                // Handle error
            }
            close(progressChan)
        }()
        
        // Convert progress to messages
        for progress := range progressChan {
            // Send to Bubble Tea (simplified)
            return copyProgressMsg{step: progress.Step, message: progress.Message}
        }
        
        return nil
    }
}
```

### Step 5: Main (Dependency Injection)

```go
// main.go
package main

import (
    "fmt"
    "os"
    
    tea "github.com/charmbracelet/bubbletea"
    "github.com/patraden/toolkit/cmd/ver2ver/internal/repository"
    "github.com/patraden/toolkit/cmd/ver2ver/internal/tui"
    "github.com/patraden/toolkit/cmd/ver2ver/internal/usecase"
    "github.com/patraden/toolkit/pkg/vertica"
)

func main() {
    // Create infrastructure (databases)
    sourceDB, err := vertica.Open(vertica.ConnParams{
        Host:     "localhost",
        Port:     5433,
        Database: "sourcedb",
        User:     "dbadmin",
    })
    if err != nil {
        fmt.Printf("Failed to connect to source: %v\n", err)
        os.Exit(1)
    }
    defer sourceDB.Close()
    
    targetDB, err := vertica.Open(vertica.ConnParams{
        Host:     "localhost",
        Port:     5434,
        Database: "targetdb",
        User:     "dbadmin",
    })
    if err != nil {
        fmt.Printf("Failed to connect to target: %v\n", err)
        os.Exit(1)
    }
    defer targetDB.Close()
    
    // Create repositories
    sourceRepo := repository.NewVerticaRepository(sourceDB)
    targetRepo := repository.NewVerticaRepository(targetDB)
    
    // Create use cases
    listTablesUseCase := usecase.NewListTablesUseCase(sourceRepo)
    copyTableUseCase := usecase.NewCopyTableUseCase(sourceRepo, targetRepo)
    
    // Create TUI with dependencies
    m := tui.NewModel(listTablesUseCase, copyTableUseCase)
    
    // Run
    p := tea.NewProgram(m, tea.WithAltScreen())
    if _, err := p.Run(); err != nil {
        fmt.Printf("Error running program: %v\n", err)
        os.Exit(1)
    }
}
```

## Testing Examples

### Unit Test: Domain

```go
// internal/domain/table_test.go
package domain_test

import (
    "testing"
    
    "github.com/patraden/toolkit/cmd/ver2ver/internal/domain"
)

func TestTable_FullName(t *testing.T) {
    table := domain.Table{
        Database: "mydb",
        Schema:   "public",
        Name:     "users",
    }
    
    want := "mydb.public.users"
    got := table.FullName()
    
    if got != want {
        t.Errorf("FullName() = %q, want %q", got, want)
    }
}
```

### Unit Test: UseCase with Mock

```go
// internal/usecase/list_tables_test.go
package usecase_test

import (
    "context"
    "testing"
    
    "github.com/patraden/toolkit/cmd/ver2ver/internal/domain"
    "github.com/patraden/toolkit/cmd/ver2ver/internal/usecase"
)

// Mock repository
type mockTableRepository struct {
    tables []domain.Table
    err    error
}

func (m *mockTableRepository) ListTables(ctx context.Context, database string) ([]domain.Table, error) {
    return m.tables, m.err
}

func (m *mockTableRepository) GetTableDDL(ctx context.Context, table domain.Table) (string, error) {
    return "", nil
}

// ... other methods

func TestListTablesUseCase_Execute(t *testing.T) {
    mockRepo := &mockTableRepository{
        tables: []domain.Table{
            {Database: "mydb", Schema: "public", Name: "users"},
            {Database: "mydb", Schema: "public", Name: "orders"},
        },
    }
    
    uc := usecase.NewListTablesUseCase(mockRepo)
    
    tables, err := uc.Execute(context.Background(), "mydb")
    if err != nil {
        t.Fatalf("Execute() error = %v", err)
    }
    
    if len(tables) != 2 {
        t.Errorf("Execute() returned %d tables, want 2", len(tables))
    }
}
```

### Integration Test: Repository

```go
// internal/repository/vertica_repository_test.go
// +build integration

package repository_test

import (
    "context"
    "testing"
    
    "github.com/patraden/toolkit/cmd/ver2ver/internal/repository"
    "github.com/patraden/toolkit/pkg/vertica"
)

func TestVerticaRepository_ListTables_Integration(t *testing.T) {
    // Skip if no test database available
    if testing.Short() {
        t.Skip("Skipping integration test")
    }
    
    db, err := vertica.Open(vertica.ConnParams{
        Host:     "localhost",
        Port:     5433,
        Database: "testdb",
        User:     "dbadmin",
    })
    if err != nil {
        t.Fatalf("Failed to connect: %v", err)
    }
    defer db.Close()
    
    repo := repository.NewVerticaRepository(db)
    
    tables, err := repo.ListTables(context.Background(), "testdb")
    if err != nil {
        t.Fatalf("ListTables() error = %v", err)
    }
    
    // Verify results
    if len(tables) == 0 {
        t.Error("Expected at least one table")
    }
}
```

## Summary: Why This Structure?

1. **`cmd/ver2ver/internal/`** → All code specific to ver2ver tool
2. **Domain** → Pure business logic, no dependencies, highly testable
3. **Repository** → Abstracts data access, mockable for testing
4. **UseCase** → Orchestrates business logic, independent of UI
5. **TUI** → Just presentation, thin layer over use cases
6. **Main** → Wires everything together (Dependency Injection)

This structure allows you to:
- Test each layer independently
- Replace TUI with CLI or API without changing business logic
- Understand one layer at a time without cognitive overload
- Follow Go idioms and best practices
