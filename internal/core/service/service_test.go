package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shrtyk/pvz-service/internal/core/domain"
	"github.com/shrtyk/pvz-service/internal/core/domain/auth"
	pAuthMock "github.com/shrtyk/pvz-service/internal/core/ports/auth/mocks"
	metricsmocks "github.com/shrtyk/pvz-service/internal/core/ports/metrics/mocks"
	pwdmocks "github.com/shrtyk/pvz-service/internal/core/ports/pwd_service/mocks"
	pRepo "github.com/shrtyk/pvz-service/internal/core/ports/repository"
	repomocks "github.com/shrtyk/pvz-service/internal/core/ports/repository/mocks"
	"github.com/shrtyk/pvz-service/internal/core/service"
	xerr "github.com/shrtyk/pvz-service/pkg/xerrors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNewPVZ(t *testing.T) {
	t.Parallel()

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
			t.Parallel()

			repo := new(repomocks.MockRepository)
			metrics := new(metricsmocks.MockCollector)
			s := service.NewAppService(time.Second, repo, nil, nil, metrics)

			repo.On("CreatePVZ", mock.Anything, tt.args).Return(tt.mockArgs.pvz, tt.mockArgs.err)
			if !tt.wantErr {
				metrics.On("IncPVZsCreated").Return()
			}

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
	t.Parallel()

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
			t.Parallel()

			repo := new(repomocks.MockRepository)
			metrics := new(metricsmocks.MockCollector)
			s := service.NewAppService(time.Second, repo, nil, nil, metrics)

			repo.On("CreateReception", mock.Anything, tt.args).Return(tt.mockArgs.rec, tt.mockArgs.err)
			if !tt.wantErr {
				metrics.On("IncReceptionsCreated").Return()
			}

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
	t.Parallel()

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
			t.Parallel()

			repo := new(repomocks.MockRepository)
			metrics := new(metricsmocks.MockCollector)
			s := service.NewAppService(time.Second, repo, nil, nil, metrics)

			repo.On("CreateProduct", mock.Anything, tt.args).Return(tt.mockArgs.prod, tt.mockArgs.err)
			if !tt.wantErr {
				metrics.On("IncProductsAdded").Return()
			}

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
	t.Parallel()

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
			t.Parallel()

			repo := new(repomocks.MockRepository)
			metrics := new(metricsmocks.MockCollector)
			s := service.NewAppService(time.Second, repo, nil, nil, metrics)
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
	t.Parallel()

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
			t.Parallel()

			repo := new(repomocks.MockRepository)
			metrics := new(metricsmocks.MockCollector)
			s := service.NewAppService(time.Second, repo, nil, nil, metrics)
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
	t.Parallel()

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
			t.Parallel()

			repo := new(repomocks.MockRepository)
			metrics := new(metricsmocks.MockCollector)
			s := service.NewAppService(time.Second, repo, nil, nil, metrics)

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

func TestGetAllPvzs(t *testing.T) {
	t.Parallel()

	type mockArgs struct {
		res []*domain.Pvz
		err error
	}
	tests := []struct {
		name     string
		mockArgs mockArgs
		wantErr  bool
	}{
		{
			name: "success",
			mockArgs: mockArgs{
				res: []*domain.Pvz{},
				err: nil,
			},
			wantErr: false,
		},
		{
			name: "error",
			mockArgs: mockArgs{
				res: nil,
				err: errors.New("repo error"),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			repo := new(repomocks.MockRepository)
			metrics := new(metricsmocks.MockCollector)
			s := service.NewAppService(time.Second, repo, nil, nil, metrics)

			repo.On("GetAllPvzs", mock.Anything).Return(tt.mockArgs.res, tt.mockArgs.err)

			result, err := s.GetAllPvzs(context.Background())

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

func TestRegisterUser(t *testing.T) {
	t.Parallel()

	type mocks struct {
		repo   *repomocks.MockRepository
		pwdSvc *pwdmocks.MockPasswordService
	}

	tests := []struct {
		name    string
		params  *auth.RegisterUserParams
		setup   func(m mocks)
		wantErr bool
	}{
		{
			name:   "success",
			params: &auth.RegisterUserParams{Email: "e@e.com", PlainPassword: "password", Role: "user"},
			setup: func(m mocks) {
				m.pwdSvc.On("Hash", "password").Return([]byte("hashed"), nil).Once()
				m.repo.On("CreateUser", mock.Anything, mock.AnythingOfType("*auth.User")).Return(&auth.User{}, nil).Once()
			},
			wantErr: false,
		},
		{
			name:   "password hash error",
			params: &auth.RegisterUserParams{Email: "e@e.com", PlainPassword: "password", Role: "user"},
			setup: func(m mocks) {
				m.pwdSvc.On("Hash", "password").Return(nil, errors.New("hash error")).Once()
			},
			wantErr: true,
		},
		{
			name:   "create user conflict",
			params: &auth.RegisterUserParams{Email: "e@e.com", PlainPassword: "password", Role: "user"},
			setup: func(m mocks) {
				m.pwdSvc.On("Hash", "password").Return([]byte("hashed"), nil).Once()
				m.repo.On("CreateUser", mock.Anything, mock.AnythingOfType("*auth.User")).
					Return(nil, &xerr.BaseErr[pRepo.RepoErrKind]{Kind: pRepo.Conflict}).Once()
			},
			wantErr: true,
		},
		{
			name:   "create user unexpected error",
			params: &auth.RegisterUserParams{Email: "e@e.com", PlainPassword: "password", Role: "user"},
			setup: func(m mocks) {
				m.pwdSvc.On("Hash", "password").Return([]byte("hashed"), nil).Once()
				m.repo.On("CreateUser", mock.Anything, mock.AnythingOfType("*auth.User")).
					Return(nil, errors.New("unexpected error")).Once()
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			repo := new(repomocks.MockRepository)
			pwdSvc := new(pwdmocks.MockPasswordService)
			metrics := new(metricsmocks.MockCollector)
			s := service.NewAppService(time.Second, repo, pwdSvc, nil, metrics)

			tt.setup(mocks{repo, pwdSvc})

			_, err := s.RegisterUser(context.Background(), tt.params)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			repo.AssertExpectations(t)
			pwdSvc.AssertExpectations(t)
		})
	}
}

func TestLoginUser(t *testing.T) {
	t.Parallel()

	type mocks struct {
		repo   *repomocks.MockRepository
		pwdSvc *pwdmocks.MockPasswordService
		tknSvc *pAuthMock.MockTokenService
	}

	loginParams := &auth.LoginUserParams{Email: "e@e.com", PlainPassword: "password"}
	user := &auth.User{Id: uuid.New(), PasswordHash: []byte("hashed")}

	tests := []struct {
		name    string
		params  *auth.LoginUserParams
		setup   func(m mocks)
		wantErr bool
	}{
		{
			name:   "success",
			params: loginParams,
			setup: func(m mocks) {
				m.repo.On("UserByEmail", mock.Anything, loginParams.Email).Return(user, nil).Once()
				m.pwdSvc.On("Compare", user.PasswordHash, loginParams.PlainPassword).Return(true, nil).Once()
				m.tknSvc.On("GenerateAccessToken", mock.Anything).Return("access_token", nil).Once()
				m.tknSvc.On("GenerateRefreshToken", mock.Anything, mock.Anything, mock.Anything).
					Return(&auth.RefreshToken{}).Once()
				m.tknSvc.On("Hash", mock.Anything).Return([]byte("hashed_token")).Once()
				m.tknSvc.On("Fingerprint", mock.Anything).Return("fingerprint").Once()
				m.repo.On("SaveRefreshToken", mock.Anything, mock.Anything).Return(nil).Once()
			},
			wantErr: false,
		},
		{
			name:   "user not found",
			params: loginParams,
			setup: func(m mocks) {
				m.repo.On("UserByEmail", mock.Anything, loginParams.Email).
					Return(nil, &xerr.BaseErr[pRepo.RepoErrKind]{Kind: pRepo.NotFound}).Once()
			},
			wantErr: true,
		},
		{
			name:   "wrong password",
			params: loginParams,
			setup: func(m mocks) {
				m.repo.On("UserByEmail", mock.Anything, loginParams.Email).Return(user, nil).Once()
				m.pwdSvc.On("Compare", user.PasswordHash, loginParams.PlainPassword).Return(false, nil).Once()
			},
			wantErr: true,
		},
		{
			name:   "generate access token error",
			params: loginParams,
			setup: func(m mocks) {
				m.repo.On("UserByEmail", mock.Anything, loginParams.Email).Return(user, nil).Once()
				m.pwdSvc.On("Compare", user.PasswordHash, loginParams.PlainPassword).Return(true, nil).Once()
				m.tknSvc.On("GenerateAccessToken", mock.Anything).Return("", errors.New("token error")).Once()
			},
			wantErr: true,
		},
		{
			name:   "save refresh token error",
			params: loginParams,
			setup: func(m mocks) {
				m.repo.On("UserByEmail", mock.Anything, loginParams.Email).Return(user, nil).Once()
				m.pwdSvc.On("Compare", user.PasswordHash, loginParams.PlainPassword).Return(true, nil).Once()
				m.tknSvc.On("GenerateAccessToken", mock.Anything).Return("access_token", nil).Once()
				m.tknSvc.On("GenerateRefreshToken", mock.Anything, mock.Anything, mock.Anything).
					Return(&auth.RefreshToken{}).Once()
				m.tknSvc.On("Hash", mock.Anything).Return([]byte("hashed_token")).Once()
				m.tknSvc.On("Fingerprint", mock.Anything).Return("fingerprint").Once()
				m.repo.On("SaveRefreshToken", mock.Anything, mock.Anything).Return(errors.New("db error")).Once()
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			repo := new(repomocks.MockRepository)
			pwdSvc := new(pwdmocks.MockPasswordService)
			tknSvc := new(pAuthMock.MockTokenService)
			metrics := new(metricsmocks.MockCollector)
			s := service.NewAppService(time.Second, repo, pwdSvc, tknSvc, metrics)

			tt.setup(mocks{repo, pwdSvc, tknSvc})

			_, _, err := s.LoginUser(context.Background(), tt.params)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			repo.AssertExpectations(t)
			pwdSvc.AssertExpectations(t)
			tknSvc.AssertExpectations(t)
		})
	}
}

func TestRefreshTokens(t *testing.T) {
	t.Parallel()

	type mocks struct {
		repo   *repomocks.MockRepository
		tknSvc *pAuthMock.MockTokenService
	}

	userRoleAndToken := &auth.UserRoleAndRToken{
		Role:   auth.UserRoleEmployee,
		RToken: &auth.RefreshToken{UserID: uuid.New().String(), ExpiresAt: time.Now().Add(time.Hour)},
	}

	tests := []struct {
		name    string
		token   *auth.RefreshToken
		setup   func(m mocks)
		wantErr bool
	}{
		{
			name:  "success",
			token: &auth.RefreshToken{Token: "refresh_token"},
			setup: func(m mocks) {
				m.tknSvc.On("Hash", "refresh_token").Return([]byte("hashed")).Once()
				m.repo.On("UserRoleAndRefreshToken", mock.Anything, []byte("hashed")).Return(userRoleAndToken, nil).Once()
				m.tknSvc.On("Fingerprint", mock.Anything).Return("fingerprint").Times(3)
				m.tknSvc.On("GenerateAccessToken", mock.Anything).Return("new_access_token", nil).Once()
				m.tknSvc.On("GenerateRefreshToken", mock.Anything, mock.Anything, mock.Anything).
					Return(&auth.RefreshToken{}).Once()
				m.tknSvc.On("Hash", mock.Anything).Return([]byte("new_hashed_token")).Once()
				m.repo.On("UpdateUserRefreshToken", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
			},
			wantErr: false,
		},
		{
			name:  "token not found",
			token: &auth.RefreshToken{Token: "refresh_token"},
			setup: func(m mocks) {
				m.tknSvc.On("Hash", "refresh_token").Return([]byte("hashed")).Once()
				m.repo.On("UserRoleAndRefreshToken", mock.Anything, []byte("hashed")).
					Return(nil, &xerr.BaseErr[pRepo.RepoErrKind]{Kind: pRepo.NotFound}).Once()
			},
			wantErr: true,
		},
		{
			name:  "fingerprint mismatch",
			token: &auth.RefreshToken{Token: "refresh_token"},
			setup: func(m mocks) {
				m.tknSvc.On("Hash", "refresh_token").Return([]byte("hashed")).Once()
				m.repo.On("UserRoleAndRefreshToken", mock.Anything, []byte("hashed")).Return(userRoleAndToken, nil).Once()
				m.tknSvc.On("Fingerprint", mock.MatchedBy(func(token *auth.RefreshToken) bool {
					return token.Token == "refresh_token"
				})).Return("one_fingerprint").Once()
				m.tknSvc.On("Fingerprint", userRoleAndToken.RToken).Return("another_fingerprint").Once()
			},
			wantErr: true,
		},
		{
			name:  "token revoked",
			token: &auth.RefreshToken{Token: "refresh_token"},
			setup: func(m mocks) {
				revokedToken := &auth.UserRoleAndRToken{
					Role: auth.UserRoleEmployee,
					RToken: &auth.RefreshToken{
						UserID:    uuid.New().String(),
						ExpiresAt: time.Now().Add(time.Hour),
						Revoked:   true,
					},
				}
				m.tknSvc.On("Hash", "refresh_token").Return([]byte("hashed")).Once()
				m.repo.On("UserRoleAndRefreshToken", mock.Anything, []byte("hashed")).Return(revokedToken, nil).Once()
				m.tknSvc.On("Fingerprint", mock.Anything).Return("fingerprint").Twice()
			},
			wantErr: true,
		},
		{
			name:  "update refresh token error",
			token: &auth.RefreshToken{Token: "refresh_token"},
			setup: func(m mocks) {
				m.tknSvc.On("Hash", "refresh_token").Return([]byte("hashed")).Once()
				m.repo.On("UserRoleAndRefreshToken", mock.Anything, []byte("hashed")).Return(userRoleAndToken, nil).Once()
				m.tknSvc.On("Fingerprint", mock.Anything).Return("fingerprint").Times(3)
				m.tknSvc.On("GenerateAccessToken", mock.Anything).Return("new_access_token", nil).Once()
				m.tknSvc.On("GenerateRefreshToken", mock.Anything, mock.Anything, mock.Anything).Return(&auth.RefreshToken{}).Once()
				m.tknSvc.On("Hash", mock.Anything).Return([]byte("new_hashed_token")).Once()
				m.repo.On("UpdateUserRefreshToken", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("db error")).Once()
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			repo := new(repomocks.MockRepository)
			tknSvc := new(pAuthMock.MockTokenService)
			metrics := new(metricsmocks.MockCollector)
			s := service.NewAppService(time.Second, repo, nil, tknSvc, metrics)

			tt.setup(mocks{repo, tknSvc})

			tokenCopy := *tt.token
			_, _, err := s.RefreshTokens(context.Background(), &tokenCopy)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			repo.AssertExpectations(t)
			tknSvc.AssertExpectations(t)
		})
	}
}
