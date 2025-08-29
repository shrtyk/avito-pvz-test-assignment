package repository

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/shrtyk/pvz-service/internal/core/domain/auth"
	pRepo "github.com/shrtyk/pvz-service/internal/core/ports/repository"
	"github.com/shrtyk/pvz-service/pkg/logger"
	xerr "github.com/shrtyk/pvz-service/pkg/xerrors"
)

func (r *repo) UserByEmail(ctx context.Context, email string) (*auth.User, error) {
	const op = "repository.GetUserByEmail"

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
	const op = "repository.CreateUser"

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
	const op = "repository.SaveRefreshToken"

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
	const op = "repository.UserRoleAndRefreshToken"

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
	const op = "repository.UpdateUserRefreshToken"
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
	const op = "repository.FinishTx"

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
