package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/shrtyk/avito-backend-spring-2025/internal/core/domain"
)

type repo struct {
	db *sql.DB
}

func NewRepo(db *sql.DB) *repo {
	return &repo{
		db: db,
	}
}

func (r *repo) CreatePVZ(ctx context.Context, pvz *domain.PVZ) error {
	query := `
		INSERT INTO pvzs (city)
		VALUES ($1)
		RETURNING id, created_at
	`

	err := r.db.QueryRowContext(ctx, query, pvz.City).Scan(
		&pvz.Id,
		&pvz.RegistrationDate,
	)
	if err != nil {
		return fmt.Errorf("repository: failed to insert pvz: %w", err)
	}
	return nil
}
