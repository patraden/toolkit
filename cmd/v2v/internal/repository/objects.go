package repository

import (
	"context"

	"github.com/patraden/toolkit/cmd/v2v/internal/domain"
)

type DBObjectsRepository interface {
	ListSchemas(ctx context.Context) ([]string, error)
	ListTablesBySchema(ctx context.Context, schema string) ([]domain.Table, error)
}
