package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/shrtyk/avito-pvz-test-assignment/internal/core/domain"
	pRepo "github.com/shrtyk/avito-pvz-test-assignment/internal/core/ports/repository"
	"github.com/shrtyk/avito-pvz-test-assignment/pkg/logger"
	xerr "github.com/shrtyk/avito-pvz-test-assignment/pkg/xerrors"
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

func (r *repo) GetPvzsData(ctx context.Context, params *domain.PvzsReadParams) ([]*domain.PvzReceptions, error) {
	op := "repository.GetPvzsData"
	l := logger.FromCtx(ctx)

	q, args := buildGetPvzDataQuery(params)
	rows, err := r.db.QueryContext(ctx, string(q), args...)
	if err != nil {
		return nil, xerr.WrapErr(op, pRepo.Unexpected, err)
	}
	defer func() {
		cerr := rows.Close()
		if cerr != nil {
			if err == nil {
				err = xerr.WrapErr(op, pRepo.Unexpected, cerr)
			} else {
				l.Warn("failed to close rows", logger.WithErr(cerr))
			}
		}
	}()

	aggregator := newPvzAggregator()

	for rows.Next() {
		var row pvzRow
		err := rows.Scan(
			&row.PvzID, &row.PvzCity, &row.PvzCreatedAt,
			&row.RecID, &row.RecStatus, &row.RecDateTime, &row.RecPvzID,
			&row.ProdID, &row.ProdDateTime, &row.ProdRecID, &row.ProdType,
		)
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
