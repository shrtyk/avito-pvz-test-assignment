package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/shrtyk/avito-backend-spring-2025/internal/core/domain"
	pRepo "github.com/shrtyk/avito-backend-spring-2025/internal/core/ports/repository"
)

type repo struct {
	db *sql.DB
}

func NewRepo(db *sql.DB) *repo {
	return &repo{
		db: db,
	}
}

func (r *repo) CreatePVZ(ctx context.Context, pvz *domain.PVZ) (*domain.PVZ, error) {
	query := `
		INSERT INTO
			pvzs (city)
		VALUES
			($1)
		RETURNING
			id, created_at
	`

	err := r.db.QueryRowContext(ctx, query, pvz.City).Scan(
		&pvz.Id,
		&pvz.RegistrationDate,
	)
	if err != nil {
		return nil, fmt.Errorf("repository: failed to insert pvz: %w", err)
	}
	return pvz, nil
}
func (r *repo) CreateReception(ctx context.Context, rec *domain.Reception) (*domain.Reception, error) {
	query := `
		INSERT INTO
			receptions (pvz_id)
		VALUES
			($1)
		RETURNING
			id, created_at, status
	`

	err := r.db.QueryRowContext(ctx, query, rec.PvzId).Scan(
		&rec.Id,
		&rec.DateTime,
		&rec.Status,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, &pRepo.ErrConstraintViolation{
				Constraint: pgErr.ConstraintName,
				Err:        err,
			}
		}
		return nil, fmt.Errorf("repository: failed to insert reception: %w", err)
	}

	return rec, nil
}
