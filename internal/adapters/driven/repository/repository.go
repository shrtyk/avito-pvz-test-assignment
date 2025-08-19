package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
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

func (r *repo) CreatePVZ(ctx context.Context, pvz *domain.Pvz) (*domain.Pvz, error) {
	op := "repository.CreatePVZ"

	err := r.db.QueryRowContext(ctx, string(createPvzQuery), pvz.City).Scan(
		&pvz.Id,
		&pvz.RegistrationDate,
	)
	if err != nil {
		return nil, xerr.WrapErr(op, pRepo.FailedCreatePvz, err)
	}

	return pvz, nil
}

func (r *repo) CreateReception(ctx context.Context, rec *domain.Reception) (*domain.Reception, error) {
	op := "repository.CreateReception"

	err := r.db.QueryRowContext(
		ctx,
		string(createReceptionQuery),
		rec.PvzId).
		Scan(&rec.Id, &rec.DateTime, &rec.Status)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.ConstraintName {
			case "fk_pvz_id":
				return nil, xerr.WrapErr(op, pRepo.InvalidReference, err)
			case "one_in_progress_reception_per_pvz_id":
				return nil, xerr.WrapErr(op, pRepo.Conflict, err)
			}
		}
		return nil, xerr.WrapErr(op, pRepo.Unexpected, err)
	}

	return rec, nil
}

func (r *repo) CreateProduct(ctx context.Context, prod *domain.Product) (*domain.Product, error) {
	op := "repository.CreateProduct"

	err := r.db.QueryRowContext(
		ctx,
		string(createProductQuery),
		prod.PvzId,
		domain.InProgress,
		prod.Type).
		Scan(&prod.Id, &prod.DateTime, &prod.ReceptionId, &prod.Type)
	if err != nil {
		var pgErr *pgconn.PgError
		if (errors.As(err, &pgErr) && pgErr.Code == "23502") || errors.Is(err, sql.ErrNoRows) {
			return nil, xerr.WrapErr(op, pRepo.NotFound, err)
		}
		return nil, xerr.WrapErr(op, pRepo.Unexpected, err)
	}

	return prod, nil
}

func (r *repo) DeleteLastProduct(ctx context.Context, pvzId *uuid.UUID) error {
	op := "repository.DeleteLastProduct"

	res, err := r.db.ExecContext(ctx, string(deleteLastProductQuery), pvzId, domain.InProgress)
	if err != nil {
		return xerr.WrapErr(op, pRepo.Unexpected, err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return xerr.WrapErr(op, pRepo.Unexpected, err)
	}

	if rowsAffected == 0 {
		return xerr.WrapErr(op, pRepo.NotFound, sql.ErrNoRows)
	}

	return nil
}

func (r *repo) CloseReceptionInPvz(ctx context.Context, pvzId *uuid.UUID) error {
	op := "repository.CloseReceptionInPvz"

	res, err := r.db.ExecContext(
		ctx,
		string(closeReceptionPvzQuery),
		domain.Close,
		pvzId,
		domain.InProgress,
	)
	if err != nil {
		return xerr.WrapErr(op, pRepo.Unexpected, err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return xerr.WrapErr(op, pRepo.Unexpected, err)
	}

	if rowsAffected == 0 {
		return xerr.WrapErr(op, pRepo.Conflict, sql.ErrNoRows)
	}

	return nil
}

func (r *repo) GetPvzsData(ctx context.Context, params *domain.PvzsReadParams) ([]*domain.PvzReceptionsProducts, error) {
	op := "repository.GetPvzsData"

	q, args := buildGetPvzDataQuery(params)
	rows, err := r.db.QueryContext(ctx, string(q), args...)
	if err != nil {
		return nil, xerr.WrapErr(op, pRepo.Unexpected, err)
	}
	defer rows.Close()

	aggregator := newPvzAggregator()

	for rows.Next() {
		row, err := scanPvzRow(rows)
		if err != nil {
			return nil, xerr.WrapErr(op, pRepo.Unexpected, err)
		}

		if err := aggregator.processRow(row); err != nil {
			return nil, xerr.WrapErr(op, pRepo.Unexpected, err)
		}
	}

	if err := rows.Err(); err != nil {
		return nil, xerr.WrapErr(op, pRepo.Unexpected, err)
	}

	return aggregator.Results(), nil
}
