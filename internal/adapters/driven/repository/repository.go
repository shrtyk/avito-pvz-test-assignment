package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/shrtyk/avito-backend-spring-2025/internal/core/domain"
	pRepo "github.com/shrtyk/avito-backend-spring-2025/internal/core/ports/repository"
	xerr "github.com/shrtyk/avito-backend-spring-2025/pkg/xerrors"
)

type repo struct {
	db *sql.DB
}

func NewRepo(db *sql.DB) *repo {
	return &repo{
		db: db,
	}
}

func (r *repo) SavePVZ(ctx context.Context, pvz *domain.PVZ) (*domain.PVZ, error) {
	op := "repository.SavePVZ"

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
		return nil, xerr.NewErr(op, pRepo.KindFailed, err)
	}

	return pvz, nil
}

func (r *repo) CreateReception(ctx context.Context, rec *domain.Reception) (*domain.Reception, error) {
	op := "repository.CreateReception"

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
		if errors.As(err, &pgErr) {
			switch pgErr.ConstraintName {
			case "fk_pvz_id":
				return nil, xerr.NewErr(op, pRepo.KindInvalidReference, err)
			case "one_in_progress_reception_per_pvz_id":
				return nil, xerr.NewErr(op, pRepo.KindConflict, err)
			}
		}
		return nil, xerr.NewErr(op, pRepo.KindFailed, err)
	}

	return rec, nil
}

func (r *repo) CreateProduct(ctx context.Context, prod *domain.Product) (*domain.Product, error) {
	op := "repository.CreateProduct"

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
		if (errors.As(err, &pgErr) && pgErr.Code == "23502") || errors.Is(err, sql.ErrNoRows) {
			return nil, xerr.NewErr(op, pRepo.KindNotFound, err)
		}
		return nil, xerr.NewErr(op, pRepo.KindFailed, err)
	}

	return prod, nil
}
