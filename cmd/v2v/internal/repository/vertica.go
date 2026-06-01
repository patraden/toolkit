package repository

import "github.com/patraden/toolkit/pkg/vertica"

type VerticaRepository struct {
	db *vertica.DB
}

func NewVerticaRepository(db *vertica.DB) *VerticaRepository {
	return &VerticaRepository{db: db}
}
