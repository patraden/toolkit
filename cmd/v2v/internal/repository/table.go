package repository

import (
	"context"

	"github.com/patraden/toolkit/cmd/v2v/internal/domain"
)

type TableRepository interface {
	GetTableDDL(ctx context.Context, table domain.Table) (string, error)
	CreateTable(ctx context.Context, ddl string) error
	ExportData(ctx context.Context, source, target domain.Table) error
	CopyData(ctx context.Context, source, target domain.Table) error
}
