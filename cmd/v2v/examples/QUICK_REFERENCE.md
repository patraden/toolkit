# Quick Reference: Where Does My Code Go?

## Decision Tree

```
Is this code...

├─ A business entity (Table, Schema, Connection)?
│  → internal/domain/
│  Example: type Table struct { ... }
│
├─ Data access (querying Vertica)?
│  ├─ Generic Vertica client? → pkg/vertica/ (already exists)
│  └─ Ver2ver-specific queries? → internal/repository/
│      Example: ListTables(), GetTableDDL()
│
├─ Business logic (copying, validation)?
│  → internal/usecase/
│  Example: CopyTableUseCase, ListTablesUseCase
│
├─ User interface (Bubble Tea components)?
│  → internal/tui/
│  Example: Model, TablePanel, ConnectionDialog
│
└─ Wiring dependencies together?
   → main.go
   Example: Create repos, use cases, inject into TUI
```

## File Naming Conventions

```
internal/domain/
  ├─ table.go          # Table entity
  └─ connection.go     # Connection info

internal/repository/
  ├─ table_repository.go        # Interface definition
  ├─ vertica_repository.go      # Vertica implementation
  └─ vertica_repository_test.go # Integration tests

internal/usecase/
  ├─ list_tables.go       # ListTablesUseCase
  ├─ list_tables_test.go  # Unit tests (with mocks)
  ├─ copy_table.go        # CopyTableUseCase
  └─ copy_table_test.go   # Unit tests

internal/tui/
  ├─ model.go              # Main Bubble Tea model
  ├─ table_panel.go        # Table display component
  ├─ connection_dialog.go  # Connection form
  └─ log_panel.go          # Operation log
```

## Import Rules

### ✅ Allowed Imports

```go
// Domain can import: NOTHING (pure business logic)
package domain

// Repository can import: domain + pkg/vertica
package repository
import (
    "github.com/patraden/toolkit/cmd/ver2ver/internal/domain"
    "github.com/patraden/toolkit/pkg/vertica"
)

// UseCase can import: domain + repository interfaces
package usecase
import (
    "github.com/patraden/toolkit/cmd/ver2ver/internal/domain"
    "github.com/patraden/toolkit/cmd/ver2ver/internal/repository"
)

// TUI can import: domain + usecase (NOT repository!)
package tui
import (
    "github.com/patraden/toolkit/cmd/ver2ver/internal/domain"
    "github.com/patraden/toolkit/cmd/ver2ver/internal/usecase"
)

// Main can import: EVERYTHING (it's the DI container)
package main
import (
    "github.com/patraden/toolkit/cmd/ver2ver/internal/domain"
    "github.com/patraden/toolkit/cmd/ver2ver/internal/repository"
    "github.com/patraden/toolkit/cmd/ver2ver/internal/usecase"
    "github.com/patraden/toolkit/cmd/ver2ver/internal/tui"
    "github.com/patraden/toolkit/pkg/vertica"
)
```

### ❌ Forbidden Imports

```go
// ❌ Domain importing anything external
package domain
import "github.com/patraden/toolkit/pkg/vertica" // NO!

// ❌ UseCase importing TUI
package usecase
import "github.com/patraden/toolkit/cmd/ver2ver/internal/tui" // NO!

// ❌ TUI importing repository implementations
package tui
import "github.com/patraden/toolkit/cmd/ver2ver/internal/repository" // NO!
// (TUI should depend on usecase, not repository)
```

## Common Questions

### Q: Should I put shared utilities in `internal/`?

**A:** Depends on scope:
- **Used by multiple tools**: `/go/internal/utils/`
- **Used only by ver2ver**: `/go/cmd/ver2ver/internal/utils/`

### Q: What about configuration structs?

**A:** 
- Config specific to ver2ver → `internal/config/config.go`
- Config parser used by many tools → `/go/internal/config/`

### Q: Can I have subpackages in `internal/`?

**A:** Yes! Example:
```
internal/
├─ domain/
│  ├─ table/          # Table-related entities
│  └─ connection/     # Connection-related entities
├─ repository/
│  ├─ vertica/        # Vertica-specific repos
│  └─ mock/           # Mock repos for testing
└─ usecase/
   ├─ copy/           # Copy-related use cases
   └─ metadata/       # Metadata use cases
```

### Q: Where do database migrations go?

**A:**
```
cmd/ver2ver/
├─ migrations/        # SQL migrations
│  ├─ 001_init.sql
│  └─ 002_add_index.sql
└─ internal/
   └─ migration/      # Migration runner code
```

### Q: What if I need to share code between v2v and ver2ver?

**A:** Extract to `/go/internal/` or `/go/pkg/`:
```
go/
├─ internal/          # Shared private code
│  └─ tablecopy/      # Common copy logic
├─ pkg/               # Shared public code
│  └─ vertica/        # Already here!
└─ cmd/
   ├─ v2v/
   └─ ver2ver/
```

## Build Commands

```bash
# Build
cd /Users/d.patrakhin/go/toolkit/go/cmd/ver2ver
go build -o ver2ver

# Run
./ver2ver

# Test
go test ./...                      # All tests
go test ./internal/usecase/...     # UseCase tests only
go test -tags=integration ./...    # Integration tests

# Test with coverage
go test -cover ./internal/...

# Check imports
go mod tidy
go mod verify
```

## Next Steps to Implement

1. **Define repository interfaces** in `internal/repository/table_repository.go`
2. **Implement repository** in `internal/repository/vertica_repository.go`
3. **Create use cases** in `internal/usecase/`
4. **Update TUI** to use use cases instead of direct DB calls
5. **Wire dependencies** in `main.go`

## Example Workflow: Adding "List Tables" Feature

```bash
# 1. Create domain entity (if needed)
# Already exists: internal/domain/table.go

# 2. Create repository interface
cat > internal/repository/table_repository.go << 'EOF'
package repository

import (
    "context"
    "github.com/patraden/toolkit/cmd/ver2ver/internal/domain"
)

type TableRepository interface {
    ListTables(ctx context.Context, database string) ([]domain.Table, error)
}
EOF

# 3. Implement repository
# Edit: internal/repository/vertica_repository.go

# 4. Create use case
# Edit: internal/usecase/list_tables.go

# 5. Update TUI to call use case
# Edit: internal/tui/model.go

# 6. Wire in main.go
# Edit: main.go

# 7. Test
go test ./internal/usecase/ -v
```

## Resources

- Clean Architecture: https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html
- Go Project Layout: https://github.com/golang-standards/project-layout
- Dependency Rule: Dependencies point INWARD only (UI → UseCase → Domain)
