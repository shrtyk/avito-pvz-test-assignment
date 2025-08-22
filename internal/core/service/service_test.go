package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shrtyk/avito-pvz-test-assignment/internal/core/domain"
	pRepo "github.com/shrtyk/avito-pvz-test-assignment/internal/core/ports/repository"
	repomocks "github.com/shrtyk/avito-pvz-test-assignment/internal/core/ports/repository/mocks"
	"github.com/shrtyk/avito-pvz-test-assignment/internal/core/service"
	xerr "github.com/shrtyk/avito-pvz-test-assignment/pkg/xerrors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNewPVZ(t *testing.T) {
	type mockArgs struct {
		pvz *domain.Pvz
		err error
	}
	tests := []struct {
		name     string
		args     *domain.Pvz
		mockArgs mockArgs
		wantErr  bool
	}{
		{
			name: "success",
			args: &domain.Pvz{City: "Moscow"},
			mockArgs: mockArgs{
				pvz: &domain.Pvz{Id: uuid.New(), City: "Moscow"},
				err: nil,
			},
			wantErr: false,
		},
		{
			name: "error",
			args: &domain.Pvz{City: "Moscow"},
			mockArgs: mockArgs{
				pvz: nil,
				err: errors.New("repo error"),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := new(repomocks.MockRepository)
			s := service.NewAppService(time.Second, repo, nil)

			repo.On("CreatePVZ", mock.Anything, tt.args).Return(tt.mockArgs.pvz, tt.mockArgs.err)

			result, err := s.NewPVZ(context.Background(), tt.args)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}
			repo.AssertExpectations(t)
		})
	}
}

func TestOpenNewPVZReception(t *testing.T) {
	type mockArgs struct {
		rec *domain.Reception
		err error
	}
	tests := []struct {
		name     string
		args     *domain.Reception
		mockArgs mockArgs
		wantErr  bool
	}{
		{
			name: "success",
			args: &domain.Reception{PvzId: uuid.New()},
			mockArgs: mockArgs{
				rec: &domain.Reception{Id: uuid.New()},
				err: nil,
			},
			wantErr: false,
		},
		{
			name: "pvz not found",
			args: &domain.Reception{PvzId: uuid.New()},
			mockArgs: mockArgs{
				rec: nil,
				err: &xerr.BaseErr[pRepo.RepoErrKind]{Kind: pRepo.InvalidReference},
			},
			wantErr: true,
		},
		{
			name: "active reception exists",
			args: &domain.Reception{PvzId: uuid.New()},
			mockArgs: mockArgs{
				rec: nil,
				err: &xerr.BaseErr[pRepo.RepoErrKind]{Kind: pRepo.Conflict},
			},
			wantErr: true,
		},
		{
			name: "unexpected error",
			args: &domain.Reception{PvzId: uuid.New()},
			mockArgs: mockArgs{
				rec: nil,
				err: errors.New("unexpected error"),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := new(repomocks.MockRepository)
			s := service.NewAppService(time.Second, repo, nil)

			repo.On("CreateReception", mock.Anything, tt.args).Return(tt.mockArgs.rec, tt.mockArgs.err)

			result, err := s.OpenNewPVZReception(context.Background(), tt.args)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}
			repo.AssertExpectations(t)
		})
	}
}

func TestAddProductPVZ(t *testing.T) {
	type mockArgs struct {
		prod *domain.Product
		err  error
	}
	tests := []struct {
		name     string
		args     *domain.Product
		mockArgs mockArgs
		wantErr  bool
	}{
		{
			name: "success",
			args: &domain.Product{PvzId: uuid.New()},
			mockArgs: mockArgs{
				prod: &domain.Product{Id: uuid.New()},
				err:  nil,
			},
			wantErr: false,
		},
		{
			name: "no active reception",
			args: &domain.Product{PvzId: uuid.New()},
			mockArgs: mockArgs{
				prod: nil,
				err:  &xerr.BaseErr[pRepo.RepoErrKind]{Kind: pRepo.NotFound},
			},
			wantErr: true,
		},
		{
			name: "unexpected error",
			args: &domain.Product{PvzId: uuid.New()},
			mockArgs: mockArgs{
				prod: nil,
				err:  errors.New("unexpected error"),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := new(repomocks.MockRepository)
			s := service.NewAppService(time.Second, repo, nil)

			repo.On("CreateProduct", mock.Anything, tt.args).Return(tt.mockArgs.prod, tt.mockArgs.err)

			result, err := s.AddProductPVZ(context.Background(), tt.args)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}
			repo.AssertExpectations(t)
		})
	}
}

func TestDeleteLastProductPvz(t *testing.T) {
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
			name:    "not found",
			mockErr: &xerr.BaseErr[pRepo.RepoErrKind]{Kind: pRepo.NotFound},
			wantErr: true,
		},
		{
			name:    "unexpected error",
			mockErr: errors.New("unexpected error"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := new(repomocks.MockRepository)
			s := service.NewAppService(time.Second, repo, nil)
			pvzId := uuid.New()

			repo.On("DeleteLastProduct", mock.Anything, &pvzId).Return(tt.mockErr)

			err := s.DeleteLastProductPvz(context.Background(), &pvzId)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			repo.AssertExpectations(t)
		})
	}
}

func TestCloseReceptionInPvz(t *testing.T) {
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
			name:    "conflict",
			mockErr: &xerr.BaseErr[pRepo.RepoErrKind]{Kind: pRepo.Conflict},
			wantErr: true,
		},
		{
			name:    "unexpected error",
			mockErr: errors.New("unexpected error"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := new(repomocks.MockRepository)
			s := service.NewAppService(time.Second, repo, nil)
			pvzId := uuid.New()

			repo.On("CloseReceptionInPvz", mock.Anything, &pvzId).Return(tt.mockErr)

			err := s.CloseReceptionInPvz(context.Background(), &pvzId)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			repo.AssertExpectations(t)
		})
	}
}

func TestGetPvzsData(t *testing.T) {
	type mockArgs struct {
		res []*domain.PvzReceptions
		err error
	}
	tests := []struct {
		name     string
		args     *domain.PvzsReadParams
		mockArgs mockArgs
		wantErr  bool
	}{
		{
			name: "success",
			args: &domain.PvzsReadParams{},
			mockArgs: mockArgs{
				res: []*domain.PvzReceptions{},
				err: nil,
			},
			wantErr: false,
		},
		{
			name: "error",
			args: &domain.PvzsReadParams{},
			mockArgs: mockArgs{
				res: nil,
				err: errors.New("repo error"),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := new(repomocks.MockRepository)
			s := service.NewAppService(time.Second, repo, nil)

			repo.On("GetPvzsData", mock.Anything, tt.args).Return(tt.mockArgs.res, tt.mockArgs.err)

			result, err := s.GetPvzsData(context.Background(), tt.args)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}
			repo.AssertExpectations(t)
		})
	}
}
