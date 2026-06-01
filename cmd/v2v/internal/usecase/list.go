package usecase

import (
	"context"

	"github.com/patraden/toolkit/cmd/v2v/internal/domain"
)

type IListTablesUsecase interface {
	Execute(ctx context.Context) ([]domain.Table, error)
}
