package repository

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/shrtyk/avito-pvz-test-assignment/internal/core/domain"
	"github.com/shrtyk/avito-pvz-test-assignment/pkg/logger"
	"github.com/stretchr/testify/assert"
)

func TestCreatePVZ(t *testing.T) {
	type mockArgs struct {
		pvz  *domain.Pvz
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
				pvz: &domain.Pvz{
					City: "Moscow",
				},
				rows: sqlmock.NewRows([]string{"id", "registration_date"}).
					AddRow(uuid.New(), time.Now()),
			},
			wantErr: false,
		},
		{
			name: "error",
			mockArgs: mockArgs{
				pvz: &domain.Pvz{
					City: "Moscow",
				},
				err: errors.New("db error"),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
			assert.NoError(t, err)
			defer func(db *sql.DB) { _ = db.Close() }(db)

			repo := NewRepo(db)

			expect := mock.ExpectQuery(".*").WithArgs(tt.mockArgs.pvz.City)

			if tt.mockArgs.err != nil {
				expect.WillReturnError(tt.mockArgs.err)
			} else {
				expect.WillReturnRows(tt.mockArgs.rows)
			}

			result, err := repo.CreatePVZ(context.Background(), tt.mockArgs.pvz)

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

func TestCreateReception(t *testing.T) {
	type mockArgs struct {
		rec  *domain.Reception
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
				rec: &domain.Reception{
					PvzId: uuid.New(),
				},
				rows: sqlmock.NewRows([]string{"id", "date_time", "status"}).
					AddRow(uuid.New(), time.Now(), domain.InProgress),
			},
			wantErr: false,
		},
		{
			name: "fk_pvz_id error",
			mockArgs: mockArgs{
				rec: &domain.Reception{
					PvzId: uuid.New(),
				},
				err: &pgconn.PgError{ConstraintName: "fk_pvz_id"},
			},
			wantErr: true,
		},
		{
			name: "conflict error",
			mockArgs: mockArgs{
				rec: &domain.Reception{
					PvzId: uuid.New(),
				},
				err: &pgconn.PgError{ConstraintName: "one_in_progress_reception_per_pvz_id"},
			},
			wantErr: true,
		},
		{
			name: "unexpected error",
			mockArgs: mockArgs{
				rec: &domain.Reception{
					PvzId: uuid.New(),
				},
				err: errors.New("db error"),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
			assert.NoError(t, err)
			defer func(db *sql.DB) { _ = db.Close() }(db)

			repo := NewRepo(db)

			expect := mock.ExpectQuery(".*").WithArgs(tt.mockArgs.rec.PvzId)

			if tt.mockArgs.err != nil {
				expect.WillReturnError(tt.mockArgs.err)
			} else {
				expect.WillReturnRows(tt.mockArgs.rows)
			}

			result, err := repo.CreateReception(context.Background(), tt.mockArgs.rec)

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

func TestCreateProduct(t *testing.T) {
	type mockArgs struct {
		prod *domain.Product
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
				prod: &domain.Product{
					PvzId: uuid.New(),
					Type:  domain.ProductTypeClothing,
				},
				rows: sqlmock.NewRows([]string{"id", "date_time", "reception_id", "type"}).
					AddRow(uuid.New(), time.Now(), uuid.New(), domain.ProductTypeClothing),
			},
			wantErr: false,
		},
		{
			name: "not found error pg",
			mockArgs: mockArgs{
				prod: &domain.Product{
					PvzId: uuid.New(),
					Type:  domain.ProductTypeClothing,
				},
				err: &pgconn.PgError{Code: "23502"},
			},
			wantErr: true,
		},
		{
			name: "not found error no rows",
			mockArgs: mockArgs{
				prod: &domain.Product{
					PvzId: uuid.New(),
					Type:  domain.ProductTypeClothing,
				},
				err: sql.ErrNoRows,
			},
			wantErr: true,
		},
		{
			name: "unexpected error",
			mockArgs: mockArgs{
				prod: &domain.Product{
					PvzId: uuid.New(),
					Type:  domain.ProductTypeClothing,
				},
				err: errors.New("db error"),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
			assert.NoError(t, err)
			defer func(db *sql.DB) { _ = db.Close() }(db)

			repo := NewRepo(db)

			expect := mock.ExpectQuery(".*").
				WithArgs(tt.mockArgs.prod.PvzId, domain.InProgress, tt.mockArgs.prod.Type)

			if tt.mockArgs.err != nil {
				expect.WillReturnError(tt.mockArgs.err)
			} else {
				expect.WillReturnRows(tt.mockArgs.rows)
			}

			result, err := repo.CreateProduct(context.Background(), tt.mockArgs.prod)

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

func TestDeleteLastProduct(t *testing.T) {
	type mockArgs struct {
		pvzId  uuid.UUID
		result sql.Result
		err    error
	}
	tests := []struct {
		name     string
		mockArgs mockArgs
		wantErr  bool
	}{
		{
			name: "success",
			mockArgs: mockArgs{
				pvzId:  uuid.New(),
				result: sqlmock.NewResult(1, 1),
			},
			wantErr: false,
		},
		{
			name: "unexpected error on exec",
			mockArgs: mockArgs{
				pvzId: uuid.New(),
				err:   errors.New("db error"),
			},
			wantErr: true,
		},
		{
			name: "unexpected error on rows affected",
			mockArgs: mockArgs{
				pvzId:  uuid.New(),
				result: sqlmock.NewErrorResult(errors.New("rows affected error")),
			},
			wantErr: true,
		},
		{
			name: "not found",
			mockArgs: mockArgs{
				pvzId:  uuid.New(),
				result: sqlmock.NewResult(0, 0),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
			assert.NoError(t, err)
			defer func(db *sql.DB) { _ = db.Close() }(db)

			repo := NewRepo(db)

			expect := mock.ExpectExec(".*").
				WithArgs(tt.mockArgs.pvzId, domain.InProgress)

			if tt.mockArgs.err != nil {
				expect.WillReturnError(tt.mockArgs.err)
			} else {
				expect.WillReturnResult(tt.mockArgs.result)
			}

			err = repo.DeleteLastProduct(context.Background(), &tt.mockArgs.pvzId)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestCloseReceptionInPvz(t *testing.T) {
	type mockArgs struct {
		pvzId  uuid.UUID
		result sql.Result
		err    error
	}
	tests := []struct {
		name     string
		mockArgs mockArgs
		wantErr  bool
	}{
		{
			name: "success",
			mockArgs: mockArgs{
				pvzId:  uuid.New(),
				result: sqlmock.NewResult(1, 1),
			},
			wantErr: false,
		},
		{
			name: "unexpected error on exec",
			mockArgs: mockArgs{
				pvzId: uuid.New(),
				err:   errors.New("db error"),
			},
			wantErr: true,
		},
		{
			name: "unexpected error on rows affected",
			mockArgs: mockArgs{
				pvzId:  uuid.New(),
				result: sqlmock.NewErrorResult(errors.New("rows affected error")),
			},
			wantErr: true,
		},
		{
			name: "conflict",
			mockArgs: mockArgs{
				pvzId:  uuid.New(),
				result: sqlmock.NewResult(0, 0),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
			assert.NoError(t, err)
			defer func(db *sql.DB) { _ = db.Close() }(db)

			repo := NewRepo(db)

			expect := mock.ExpectExec(".*").
				WithArgs(domain.Close, tt.mockArgs.pvzId, domain.InProgress)

			if tt.mockArgs.err != nil {
				expect.WillReturnError(tt.mockArgs.err)
			} else {
				expect.WillReturnResult(tt.mockArgs.result)
			}

			err = repo.CloseReceptionInPvz(context.Background(), &tt.mockArgs.pvzId)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestGetPvzsData(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	assert.NoError(t, err)
	defer func(db *sql.DB) {
		_ = db.Close()
	}(db)

	repo := NewRepo(db)
	l, _ := logger.NewTestLogger()
	ctx := logger.ToCtx(context.Background(), l)

	t.Run("success no rows", func(t *testing.T) {
		params := &domain.PvzsReadParams{Page: 1, Limit: 10}
		_, args := buildGetPvzDataQuery(params)
		driverArgs := make([]driver.Value, len(args))
		for i, v := range args {
			driverArgs[i] = v
		}

		rows := sqlmock.NewRows(
			[]string{
				"pvz_id", "pvz_city", "pvz_created_at", "rec_id",
				"rec_status", "rec_date_time", "rec_pvz_id", "prod_id",
				"prod_date_time", "prod_rec_id", "prod_type"})

		mock.ExpectQuery(".*").WithArgs(driverArgs...).WillReturnRows(rows)

		result, err := repo.GetPvzsData(ctx, params)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Empty(t, result)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("success one row", func(t *testing.T) {
		params := &domain.PvzsReadParams{Page: 1, Limit: 10}
		_, args := buildGetPvzDataQuery(params)
		driverArgs := make([]driver.Value, len(args))
		for i, v := range args {
			driverArgs[i] = v
		}

		pvzId := uuid.New()
		recId := uuid.New()
		prodId := uuid.New()

		rows := sqlmock.NewRows(
			[]string{
				"pvz_id", "pvz_city", "pvz_created_at", "rec_id",
				"rec_status", "rec_date_time", "rec_pvz_id", "prod_id",
				"prod_date_time", "prod_rec_id", "prod_type"}).
			AddRow(
				pvzId, "Moscow", time.Now(), recId,
				domain.InProgress, time.Now(), pvzId, prodId,
				time.Now(), recId, domain.ProductTypeClothing)

		mock.ExpectQuery(".*").WithArgs(driverArgs...).WillReturnRows(rows)

		result, err := repo.GetPvzsData(ctx, params)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result, 1)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("query error", func(t *testing.T) {
		params := &domain.PvzsReadParams{Page: 1, Limit: 10}
		_, args := buildGetPvzDataQuery(params)
		driverArgs := make([]driver.Value, len(args))
		for i, v := range args {
			driverArgs[i] = v
		}

		mock.ExpectQuery(".*").WithArgs(driverArgs...).WillReturnError(errors.New("db error"))

		result, err := repo.GetPvzsData(ctx, params)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("scan error", func(t *testing.T) {
		params := &domain.PvzsReadParams{Page: 1, Limit: 10}
		_, args := buildGetPvzDataQuery(params)
		driverArgs := make([]driver.Value, len(args))
		for i, v := range args {
			driverArgs[i] = v
		}

		rows := sqlmock.NewRows([]string{"pvz_id"}).AddRow("not a uuid")

		mock.ExpectQuery(".*").WithArgs(driverArgs...).WillReturnRows(rows)

		result, err := repo.GetPvzsData(ctx, params)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("rows error", func(t *testing.T) {
		params := &domain.PvzsReadParams{Page: 1, Limit: 10}
		_, args := buildGetPvzDataQuery(params)
		driverArgs := make([]driver.Value, len(args))
		for i, v := range args {
			driverArgs[i] = v
		}

		rows := sqlmock.NewRows(
			[]string{"pvz_id", "pvz_city", "pvz_created_at", "rec_id",
				"rec_status", "rec_date_time", "rec_pvz_id", "prod_id",
				"prod_date_time", "prod_rec_id", "prod_type"}).
			AddRow(
				uuid.New(), "Moscow", time.Now(), uuid.New(),
				domain.InProgress, time.Now(), uuid.New(), uuid.New(),
				time.Now(), uuid.New(), domain.ProductTypeClothing).
			RowError(0, errors.New("row error"))

		mock.ExpectQuery(".*").WithArgs(driverArgs...).WillReturnRows(rows)

		result, err := repo.GetPvzsData(ctx, params)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
