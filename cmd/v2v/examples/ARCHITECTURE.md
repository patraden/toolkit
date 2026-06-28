# Ver2Ver Architecture

## Clean Architecture Layers

This tool follows Clean Architecture principles with clear separation of concerns:

```
┌─────────────────────────────────────────────────────┐
│                    main.go                          │
│            (Dependency Injection)                   │
└─────────────────────────────────────────────────────┘
                        │
        ┌───────────────┼───────────────┐
        ▼               ▼               ▼
   ┌────────┐     ┌──────────┐    ┌──────────┐
   │  TUI   │────▶│ UseCase  │───▶│  Domain  │
   └────────┘     └──────────┘    └──────────┘
   (Delivery)     (Application)    (Business)
        │               │
        └───────┬───────┘
                ▼
        ┌──────────────┐
        │  Repository  │
        └──────────────┘
        (Data Access)
                │
                ▼
        ┌──────────────┐
        │ pkg/vertica  │
        └──────────────┘
        (Infrastructure)
```

## Layer Responsibilities

### 1. Domain Layer (`internal/domain/`)
- **Pure business entities** with no external dependencies
- Examples: `Table`, `Schema`, `Connection`
- Business rules and validations
- No awareness of databases, UI, or frameworks

### 2. Repository Layer (`internal/repository/`)
- **Data access interfaces and implementations**
- Abstracts away Vertica-specific details
- Uses `pkg/vertica` for actual DB communication
- Examples: `VerticaRepository` implementing `TableRepository` interface

### 3. UseCase Layer (`internal/usecase/`)
- **Application business logic**
- Orchestrates domain objects and repositories
- Examples: `CopyTableUseCase`, `ListTablesUseCase`
- Independent of delivery mechanism (CLI, TUI, API)

### 4. TUI Layer (`internal/tui/`)
- **Bubble Tea UI components**
- Adapts use cases to terminal interface
- Handles user input, rendering, state management
- Examples: `Model`, `TablePanel`, `ConnectionDialog`

### 5. Main (`main.go`)
- **Dependency injection and wiring**
- Creates repositories, use cases, and UI
- Keeps everything loosely coupled

## Dependency Rule

**Dependencies point inward only:**
- TUI → UseCase → Domain
- Repository → Domain
- TUI/Repository → pkg/vertica (infrastructure)

Domain layer has **zero external dependencies**.

## Why This Structure?

### ✅ Benefits
1. **Testability**: Each layer can be tested independently with mocks
2. **Maintainability**: Clear boundaries, easy to understand
3. **Flexibility**: Can swap TUI for CLI/API without changing business logic
4. **Reusability**: Use cases can be reused in different delivery mechanisms
5. **Go Best Practices**: `internal/` prevents unwanted external dependencies

### 📁 File Organization
```
cmd/ver2ver/
├── main.go                          # Entry point, DI container
└── internal/                        # Private to ver2ver only
    ├── domain/
    │   ├── table.go                 # Business entity
    │   └── connection.go            # Connection info
    ├── repository/
    │   ├── table_repository.go      # Interface
    │   └── vertica_repository.go    # Implementation
    ├── usecase/
    │   ├── copy_table.go            # Copy logic
    │   └── list_tables.go           # List logic
    └── tui/
        ├── model.go                 # Main Bubble Tea model
        ├── table_panel.go           # Table display component
        ├── connection_dialog.go     # Connection form
        └── log_panel.go             # Operation log
```

## Example: Adding a New Feature

To add "Export Table to CSV":

1. **Domain**: Define `CSVExportOptions` struct
2. **Repository**: Add `ExportToCSV(table Table, opts CSVExportOptions) error` to interface
3. **UseCase**: Create `ExportTableUseCase` with business logic
4. **TUI**: Add F7 key binding to trigger export
5. **Main**: Wire up the new use case

Each layer changes independently without cascading modifications.

## Testing Strategy

```go
// Domain: Pure unit tests
func TestTable_SizeMb(t *testing.T) { ... }

// UseCase: Test with mock repository
func TestCopyTableUseCase(t *testing.T) {
    mockRepo := &MockRepository{}
    useCase := NewCopyTableUseCase(mockRepo)
    // Test business logic
}

// Repository: Integration tests with real Vertica
func TestVerticaRepository_Integration(t *testing.T) {
    // Requires real DB connection
}

// TUI: Test state transitions
func TestModel_Update(t *testing.T) {
    // Test Bubble Tea messages
}
```

## Next Steps

1. Move existing code to `internal/` subdirectories
2. Define repository interfaces in `repository/table_repository.go`
3. Implement use cases that orchestrate the business logic
4. Update `main.go` to wire dependencies together
5. Update TUI to call use cases instead of direct DB access
