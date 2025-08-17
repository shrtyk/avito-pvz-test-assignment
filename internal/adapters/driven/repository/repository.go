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
		if errors.As(err, &pgErr) && (pgErr.Code == "23505" || pgErr.Code == "23503") {
			return nil, &pRepo.ErrConstraintViolation{
				Constraint: pgErr.ConstraintName,
				Err:        err,
			}
		}
		return nil, fmt.Errorf("repository: failed to insert reception: %w", err)
	}

	return rec, nil
}

func (r *repo) CreateProduct(ctx context.Context, prod *domain.Product) (*domain.Product, error) {
	query := `
		INSERT INTO
			products (reception_id, type)
		VALUES(
			(SELECT id FROM receptions WHERE pvz_id = $1 AND status = $2), $3
		)
		RETURNING
			id, added_at, reception_id, type
	`

	err := r.db.QueryRowContext(ctx, query, prod.PvzId, domain.InProgress, prod.Type).Scan(
		&prod.Id,
		&prod.DateTime,
		&prod.ReceptionId,
		&prod.Type,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		switch {
		case errors.As(err, &pgErr) && pgErr.Code == "23502":
			return nil, &pRepo.ErrNullConstraint{Err: err}
		case errors.Is(err, sql.ErrNoRows):
			return nil, &pRepo.ErrNoRowsInserted{Err: err}
		}
		return nil, fmt.Errorf("repository: failed to add product: %w", err)
	}

	return prod, nil
}
