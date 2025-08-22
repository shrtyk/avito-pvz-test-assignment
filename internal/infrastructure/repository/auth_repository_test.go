package repository

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/shrtyk/avito-pvz-test-assignment/internal/core/domain/auth"
	"github.com/shrtyk/avito-pvz-test-assignment/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserByEmail(t *testing.T) {
	t.Parallel()
	type mockArgs struct {
		email string
		rows  *sqlmock.Rows
		err   error
	}
	tests := []struct {
		name     string
		mockArgs mockArgs
		wantErr  bool
	}{
		{
			name: "success",
			mockArgs: mockArgs{
				email: "test@example.com",
				rows: sqlmock.NewRows([]string{
					"id", "password_hash", "role", "created_at"}).
					AddRow(uuid.New(), "hash", "user", time.Now()),
			},
			wantErr: false,
		},
		{
			name: "not found",
			mockArgs: mockArgs{
				email: "test@example.com",
				err:   sql.ErrNoRows,
			},
			wantErr: true,
		},
		{
			name: "unexpected error",
			mockArgs: mockArgs{
				email: "test@example.com",
				err:   errors.New("db error"),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
			require.NoError(t, err)
			defer func(db *sql.DB) { _ = db.Close() }(db)

			repo := NewRepo(db)

			expect := mock.ExpectQuery(".*").WithArgs(tt.mockArgs.email)

			if tt.mockArgs.err != nil {
				expect.WillReturnError(tt.mockArgs.err)
			} else {
				expect.WillReturnRows(tt.mockArgs.rows)
			}

			result, err := repo.UserByEmail(context.Background(), tt.mockArgs.email)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestCreateUser(t *testing.T) {
	t.Parallel()
	type mockArgs struct {
		user *auth.User
		rows *sqlmock.Rows
		err  error
	}
	tests := []struct {
		name     string
		mockArgs mockArgs
		wantErr  bool
	}{
		{
			name: "success",
			mockArgs: mockArgs{
				user: &auth.User{
					Email:        "test@example.com",
					Role:         "user",
					PasswordHash: []byte("hash"),
				},
				rows: sqlmock.NewRows([]string{"id", "email", "role"}).
					AddRow(uuid.New(), "test@example.com", "user"),
			},
			wantErr: false,
		},
		{
			name: "conflict",
			mockArgs: mockArgs{
				user: &auth.User{
					Email:        "test@example.com",
					Role:         "user",
					PasswordHash: []byte("hash"),
				},
				err: &pgconn.PgError{Code: "23505"},
			},
			wantErr: true,
		},
		{
			name: "unexpected error",
			mockArgs: mockArgs{
				user: &auth.User{
					Email:        "test@example.com",
					Role:         "user",
					PasswordHash: []byte("hash"),
				},
				err: errors.New("db error"),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
			require.NoError(t, err)
			defer func(db *sql.DB) { _ = db.Close() }(db)

			repo := NewRepo(db)

			expect := mock.ExpectQuery(".*").WithArgs(tt.mockArgs.user.Email, tt.mockArgs.user.Role, tt.mockArgs.user.PasswordHash)

			if tt.mockArgs.err != nil {
				expect.WillReturnError(tt.mockArgs.err)
			} else {
				expect.WillReturnRows(tt.mockArgs.rows)
			}

			result, err := repo.CreateUser(context.Background(), tt.mockArgs.user)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestSaveRefreshToken(t *testing.T) {
	t.Parallel()
	rToken := &auth.RefreshToken{
		TokenHash:   []byte("hash"),
		Fingerprint: "fp",
		UserID:      uuid.New().String(),
		UserAgent:   "ua",
		IP:          "ip",
		CreatedAt:   time.Now(),
		ExpiresAt:   time.Now().Add(time.Hour),
	}

	tests := []struct {
		name    string
		mockErr error
		wantErr bool
	}{
		{
			name:    "success",
			mockErr: nil,
			wantErr: false,
		},
		{
			name:    "db error",
			mockErr: errors.New("db error"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
			require.NoError(t, err)
			defer func(db *sql.DB) { _ = db.Close() }(db)

			repo := NewRepo(db)

			expect := mock.ExpectExec(".*").WithArgs(
				rToken.TokenHash,
				rToken.Fingerprint,
				rToken.UserID,
				rToken.UserAgent,
				rToken.IP,
				rToken.CreatedAt,
				rToken.ExpiresAt,
			)

			if tt.mockErr != nil {
				expect.WillReturnError(tt.mockErr)
			} else {
				expect.WillReturnResult(sqlmock.NewResult(1, 1))
			}

			err = repo.SaveRefreshToken(context.Background(), rToken)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestUserRoleAndRefreshToken(t *testing.T) {
	t.Parallel()
	type mockArgs struct {
		tokenHash []byte
		rows      *sqlmock.Rows
		err       error
	}
	tests := []struct {
		name     string
		mockArgs mockArgs
		wantErr  bool
	}{
		{
			name: "success",
			mockArgs: mockArgs{
				tokenHash: []byte("hash"),
				rows: sqlmock.NewRows([]string{
					"role", "fingerprint", "user_id", "user_agent",
					"ip_address", "created_at", "expires_at", "revoked"},
				).AddRow("user", "fp", uuid.New(), "ua", "ip", time.Now(), time.Now().Add(time.Hour), false),
			},
			wantErr: false,
		},
		{
			name: "not found",
			mockArgs: mockArgs{
				tokenHash: []byte("hash"),
				err:       sql.ErrNoRows,
			},
			wantErr: true,
		},
		{
			name: "unexpected error",
			mockArgs: mockArgs{
				tokenHash: []byte("hash"),
				err:       errors.New("db error"),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
			require.NoError(t, err)
			defer func(db *sql.DB) { _ = db.Close() }(db)

			repo := NewRepo(db)

			expect := mock.ExpectQuery(".*").WithArgs(tt.mockArgs.tokenHash)

			if tt.mockArgs.err != nil {
				expect.WillReturnError(tt.mockArgs.err)
			} else {
				expect.WillReturnRows(tt.mockArgs.rows)
			}

			result, err := repo.UserRoleAndRefreshToken(context.Background(), tt.mockArgs.tokenHash)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestUpdateUserRefreshToken(t *testing.T) {
	t.Parallel()
	rToken := &auth.RefreshToken{
		TokenHash:   []byte("new_hash"),
		Fingerprint: "fp",
		UserID:      uuid.New().String(),
		UserAgent:   "ua",
		IP:          "ip",
		CreatedAt:   time.Now(),
		ExpiresAt:   time.Now().Add(time.Hour),
	}
	usedHash := []byte("used_hash")

	tests := []struct {
		name    string
		setup   func(mock sqlmock.Sqlmock)
		wantErr bool
	}{
		{
			name: "success",
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("UPDATE refresh_tokens").WithArgs(usedHash).WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectExec("INSERT INTO refresh_tokens").WithArgs(
					rToken.TokenHash,
					rToken.Fingerprint,
					rToken.UserID,
					rToken.UserAgent,
					rToken.IP,
					rToken.CreatedAt,
					rToken.ExpiresAt,
				).WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			wantErr: false,
		},
		{
			name: "begin error",
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin().WillReturnError(errors.New("begin error"))
			},
			wantErr: true,
		},
		{
			name: "revoke error",
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("UPDATE refresh_tokens").WithArgs(usedHash).WillReturnError(errors.New("revoke error"))
				mock.ExpectRollback()
			},
			wantErr: true,
		},
		{
			name: "insert error",
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("UPDATE refresh_tokens").WithArgs(usedHash).WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectExec("INSERT INTO refresh_tokens").WillReturnError(errors.New("insert error"))
				mock.ExpectRollback()
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			db, mock, err := sqlmock.New()
			require.NoError(t, err)
			defer func(db *sql.DB) { _ = db.Close() }(db)

			repo := NewRepo(db)
			tt.setup(mock)

			err = repo.UpdateUserRefreshToken(context.Background(), usedHash, rToken)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestFinishTx(t *testing.T) {
	t.Parallel()
	repo := NewRepo(nil)
	l, _ := logger.NewTestLogger()

	tests := []struct {
		name        string
		txErr       error
		setup       func(mock sqlmock.Sqlmock)
		wantErr     bool
		errContains string
	}{
		{
			name:  "commit success",
			txErr: nil,
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectCommit()
			},
			wantErr: false,
		},
		{
			name:  "rollback success",
			txErr: errors.New("some error"),
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectRollback()
			},
			wantErr:     true,
			errContains: "some error",
		},
		{
			name:  "commit error",
			txErr: nil,
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectCommit().WillReturnError(errors.New("commit error"))
			},
			wantErr:     true,
			errContains: "commit error",
		},
		{
			name:  "rollback error",
			txErr: errors.New("some error"),
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectRollback().WillReturnError(errors.New("rollback error"))
			},
			wantErr:     true,
			errContains: "failed to rollback transaction",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			db, mock, err := sqlmock.New()
			require.NoError(t, err)
			defer func(db *sql.DB) { _ = db.Close() }(db)

			mock.ExpectBegin()
			tx, err := db.Begin()
			require.NoError(t, err)

			tt.setup(mock)

			txErrCopy := tt.txErr
			err = repo.FinishTx(tx, &txErrCopy, l)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
