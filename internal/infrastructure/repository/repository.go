package repository

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/shrtyk/avito-pvz-test-assignment/internal/core/domain"
	"github.com/shrtyk/avito-pvz-test-assignment/internal/core/domain/auth"
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
		if closeErr := rows.Close(); closeErr != nil {
			l.Warn("failed to close rows", logger.WithErr(closeErr))
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

func (r *repo) GetAllPvzs(ctx context.Context) ([]*domain.Pvz, error) {
	op := "repository.GetAllPvzs"
	l := logger.FromCtx(ctx)

	rows, err := r.db.QueryContext(ctx, string(getAllPvzsQuery))
	if err != nil {
		return nil, xerr.WrapErr(op, pRepo.Unexpected, err)
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			l.Warn("failed to close rows", logger.WithErr(closeErr))
		}
	}()

	pvzs := make([]*domain.Pvz, 0)
	for rows.Next() {
		pvz := new(domain.Pvz)
		err := rows.Scan(&pvz.Id, &pvz.RegistrationDate, &pvz.City)
		if err != nil {
			return nil, xerr.WrapErr(op, pRepo.Unexpected, err)
		}
		pvzs = append(pvzs, pvz)
	}

	if err := rows.Err(); err != nil {
		return nil, xerr.WrapErr(op, pRepo.Unexpected, err)
	}

	return pvzs, nil
}

func (r *repo) UserByEmail(ctx context.Context, email string) (*auth.User, error) {
	op := "repository.GetUserByEmail"

	u := new(auth.User)
	err := r.db.QueryRow(string(getUserByEmailQuery), email).Scan(
		&u.Id,
		&u.PasswordHash,
		&u.Role,
		&u.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, xerr.WrapErr(op, pRepo.NotFound, err)
		}
		return nil, xerr.WrapErr(op, pRepo.Unexpected, err)
	}

	return u, nil
}

func (r *repo) CreateUser(ctx context.Context, user *auth.User) (*auth.User, error) {
	op := "repository.CreateUser"

	err := r.db.QueryRowContext(
		ctx,
		string(insertUserQuery),
		user.Email,
		user.Role,
		user.PasswordHash).
		Scan(
			&user.Id,
			&user.Email,
			&user.Role,
		)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, xerr.WrapErr(op, pRepo.Conflict, err)
		}
		return nil, xerr.WrapErr(op, pRepo.Unexpected, err)
	}

	return user, nil
}

func (r *repo) SaveRefreshToken(ctx context.Context, rToken *auth.RefreshToken) error {
	op := "repository.SaveRefreshToken"

	_, err := r.db.ExecContext(
		ctx,
		string(insertRefreshTokenQuery),
		rToken.TokenHash,
		rToken.Fingerprint,
		rToken.UserID,
		rToken.UserAgent,
		rToken.IP,
		rToken.CreatedAt,
		rToken.ExpiresAt,
	)
	if err != nil {
		return xerr.WrapErr(op, pRepo.Unexpected, err)
	}

	return nil
}

func (r *repo) UserRoleAndRefreshToken(ctx context.Context, tokenHash []byte) (*auth.UserRoleAndRToken, error) {
	op := "repository.UserRoleAndRefreshToken"

	urt := &auth.UserRoleAndRToken{
		RToken: new(auth.RefreshToken),
	}

	err := r.db.QueryRowContext(
		ctx,
		string(getRefreshTokenByHashQuery),
		tokenHash).
		Scan(
			&urt.Role,
			&urt.RToken.Fingerprint,
			&urt.RToken.UserID,
			&urt.RToken.UserAgent,
			&urt.RToken.IP,
			&urt.RToken.CreatedAt,
			&urt.RToken.ExpiresAt,
			&urt.RToken.Revoked,
		)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, xerr.WrapErr(op, pRepo.NotFound, err)
		}
		return nil, xerr.WrapErr(op, pRepo.Unexpected, err)
	}

	urt.RToken.TokenHash = tokenHash
	return urt, nil
}

func (r *repo) UpdateUserRefreshToken(ctx context.Context, usedHash []byte, rToken *auth.RefreshToken) error {
	op := "repository.UpdateUserRefreshToken"
	l := logger.FromCtx(ctx)

	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return xerr.WrapErr(op, pRepo.Unexpected, err)
	}
	defer func() {
		err = r.FinishTx(tx, &err, l)
	}()

	_, err = tx.ExecContext(ctx, string(revokeOldRefreshTokenQuery), usedHash)
	if err != nil {
		return xerr.WrapErr(op, pRepo.Unexpected, err)
	}

	_, err = tx.ExecContext(
		ctx,
		string(insertRefreshTokenQuery),
		rToken.TokenHash,
		rToken.Fingerprint,
		rToken.UserID,
		rToken.UserAgent,
		rToken.IP,
		rToken.CreatedAt,
		rToken.ExpiresAt,
	)
	if err != nil {
		return xerr.WrapErr(op, pRepo.Unexpected, err)
	}

	return nil
}

func (r *repo) FinishTx(tx *sql.Tx, err *error, l *slog.Logger) error {
	op := "repository.FinishTx"

	if *err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			l.Error(
				"transaction rollback failed",
				slog.String("original_error", (*err).Error()),
				slog.String("rollback_error", rollbackErr.Error()),
			)
			return xerr.WrapErr(op, pRepo.TxRollbackFailed, *err)
		}
		return *err
	}

	if commitErr := tx.Commit(); commitErr != nil {
		l.Error("transaction commit failed", logger.WithErr(commitErr))
		return xerr.WrapErr(op, pRepo.Unexpected, commitErr)
	}

	return nil
}
