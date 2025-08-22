package service

import (
	"context"
	"crypto/subtle"
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/shrtyk/avito-pvz-test-assignment/internal/core/domain"
	"github.com/shrtyk/avito-pvz-test-assignment/internal/core/domain/auth"
	pa "github.com/shrtyk/avito-pvz-test-assignment/internal/core/ports/auth"
	pwd "github.com/shrtyk/avito-pvz-test-assignment/internal/core/ports/pwd_service"
	pr "github.com/shrtyk/avito-pvz-test-assignment/internal/core/ports/repository"
	ps "github.com/shrtyk/avito-pvz-test-assignment/internal/core/ports/service"
	xerr "github.com/shrtyk/avito-pvz-test-assignment/pkg/xerrors"
)

type service struct {
	timeout time.Duration
	repo    pr.Repository
	pwdSrc  pwd.PasswordService
	tknSrc  pa.TokenService
}

func NewAppService(
	timeout time.Duration,
	repo pr.Repository,
	pwdSrc pwd.PasswordService,
	tknSrc pa.TokenService,
) *service {
	return &service{
		timeout: timeout,
		repo:    repo,
		pwdSrc:  pwdSrc,
		tknSrc:  tknSrc,
	}
}

func (s *service) NewPVZ(ctx context.Context, pvz *domain.Pvz) (*domain.Pvz, error) {
	op := "service.NewPVZ"

	tctx, tcancel := context.WithTimeout(ctx, s.timeout)
	defer tcancel()

	pvz, err := s.repo.CreatePVZ(tctx, pvz)
	if err != nil {
		return nil, xerr.WrapErr(op, ps.FailedToAddPvz, err)
	}

	return pvz, nil
}

func (s *service) OpenNewPVZReception(ctx context.Context, rec *domain.Reception) (*domain.Reception, error) {
	op := "service.OpenNewPVZReception"

	tctx, tcancel := context.WithTimeout(ctx, s.timeout)
	defer tcancel()

	newRec, err := s.repo.CreateReception(tctx, rec)
	if err != nil {
		var repoErr *xerr.BaseErr[pr.RepoErrKind]
		if errors.As(err, &repoErr) {
			switch repoErr.Kind {
			case pr.InvalidReference:
				return nil, xerr.WrapErr(op, ps.PvzNotFound, err)
			case pr.Conflict:
				return nil, xerr.WrapErr(op, ps.ActiveReceptionExists, err)
			}
		}
		return nil, xerr.WrapErr(op, ps.Unexpected, err)
	}

	return newRec, nil
}

func (s *service) AddProductPVZ(ctx context.Context, prod *domain.Product) (*domain.Product, error) {
	op := "service.AddProductPVZ"

	tctx, tcancel := context.WithTimeout(ctx, s.timeout)
	defer tcancel()

	newProd, err := s.repo.CreateProduct(tctx, prod)
	if err != nil {
		var repoErr *xerr.BaseErr[pr.RepoErrKind]
		if errors.As(err, &repoErr) && repoErr.Kind == pr.NotFound {
			return nil, xerr.WrapErr(op, ps.NoActiveReception, err)
		}
		return nil, xerr.WrapErr(op, ps.Unexpected, err)
	}

	return newProd, nil
}

func (s *service) DeleteLastProductPvz(ctx context.Context, pvzId *uuid.UUID) error {
	op := "service.DeleteLastProductPvz"

	tctx, tcancel := context.WithTimeout(ctx, s.timeout)
	defer tcancel()

	if err := s.repo.DeleteLastProduct(tctx, pvzId); err != nil {
		var bErr *xerr.BaseErr[pr.RepoErrKind]
		if errors.As(err, &bErr) && bErr.Kind == pr.NotFound {
			return xerr.WrapErr(op, ps.NoProdOrActiveReception, err)
		}
		return xerr.WrapErr(op, ps.Unexpected, err)
	}

	return nil
}

func (s *service) CloseReceptionInPvz(ctx context.Context, pvzId *uuid.UUID) error {
	op := "service.CloseReceptionInPvz"

	tctx, tcancel := context.WithTimeout(ctx, s.timeout)
	defer tcancel()

	if err := s.repo.CloseReceptionInPvz(tctx, pvzId); err != nil {
		var bErr *xerr.BaseErr[pr.RepoErrKind]
		if errors.As(err, &bErr) && bErr.Kind == pr.Conflict {
			return xerr.WrapErr(op, ps.FailedToCloseReception, err)
		}
		return xerr.WrapErr(op, ps.Unexpected, err)
	}

	return nil
}

func (s *service) GetPvzsData(ctx context.Context, params *domain.PvzsReadParams) ([]*domain.PvzReceptions, error) {
	op := "service.GetPvzsData"

	tctx, tcancel := context.WithTimeout(ctx, s.timeout)
	defer tcancel()

	res, err := s.repo.GetPvzsData(tctx, params)
	if err != nil {
		return nil, xerr.WrapErr(op, ps.Unexpected, err)
	}

	return res, nil
}

func (s *service) GetAllPvzs(ctx context.Context) ([]*domain.Pvz, error) {
	op := "service.GetAllPvzs"

	tctx, tcancel := context.WithTimeout(ctx, s.timeout)
	defer tcancel()

	res, err := s.repo.GetAllPvzs(tctx)
	if err != nil {
		return nil, xerr.WrapErr(op, ps.Unexpected, err)
	}

	return res, nil
}

func (s *service) RegisterUser(ctx context.Context, rParams *auth.RegisterUserParams) (*auth.User, error) {
	op := "service.RegisterUser"

	tctx, tcancel := context.WithTimeout(ctx, s.timeout)
	defer tcancel()

	pwdHash, err := s.pwdSrc.Hash(rParams.PlainPassword)
	if err != nil {
		return nil, xerr.WrapErr(op, ps.Unexpected, err)
	}

	user := &auth.User{
		Email:        rParams.Email,
		PasswordHash: pwdHash,
		Role:         rParams.Role,
	}

	newUser, err := s.repo.CreateUser(tctx, user)
	if err != nil {
		var bErr *xerr.BaseErr[pr.RepoErrKind]
		if errors.As(err, &bErr) && bErr.Kind == pr.Conflict {
			return nil, xerr.WrapErr(op, ps.EmailAlreadyExists, err)
		}
		return nil, xerr.WrapErr(op, ps.Unexpected, err)
	}

	return newUser, nil
}

func (s *service) LoginUser(
	ctx context.Context,
	lParams *auth.LoginUserParams,
) (aToken string, rToken *auth.RefreshToken, err error) {
	op := "service.LoginUser"

	tctx, tcancel := context.WithTimeout(ctx, s.timeout)
	defer tcancel()

	u, err := s.repo.UserByEmail(tctx, lParams.Email)
	if err != nil {
		var bErr *xerr.BaseErr[pr.RepoErrKind]
		if errors.As(err, &bErr) && bErr.Kind == pr.NotFound {
			return "", nil, xerr.WrapErr(op, ps.WrongCredentials, err)
		}
		return "", nil, xerr.WrapErr(op, ps.Unexpected, err)
	}

	ok, err := s.pwdSrc.Compare(u.PasswordHash, lParams.PlainPassword)
	if err != nil {
		return "", nil, xerr.WrapErr(op, ps.Unexpected, err)
	}

	if !ok {
		return "", nil, xerr.NewErr(op, ps.WrongCredentials)
	}

	aToken, err = s.tknSrc.GenerateAccessToken(auth.AccessTokenData{
		UserID: u.Id,
		Role:   u.Role,
	})
	if err != nil {
		return "", nil, xerr.WrapErr(op, ps.Unexpected, err)
	}
	rTokenData := s.tknSrc.GenerateRefreshToken(
		u.Id.String(),
		lParams.UserAgent,
		lParams.IP,
	)

	rTokenData.TokenHash = s.tknSrc.Hash(rTokenData.Token)
	rTokenData.Fingerprint = s.tknSrc.Fingerprint(rTokenData)
	err = s.repo.SaveRefreshToken(tctx, rTokenData)
	if err != nil {
		return "", nil, xerr.WrapErr(op, ps.Unexpected, err)
	}

	return aToken, rTokenData, nil
}

func (s *service) RefreshTokens(
	ctx context.Context,
	providedToken *auth.RefreshToken,
) (newAToken string, newRToken *auth.RefreshToken, err error) {
	op := "service.RefreshTokens"

	tctx, tcancel := context.WithTimeout(ctx, s.timeout)
	defer tcancel()

	providedToken.TokenHash = s.tknSrc.Hash(providedToken.Token)

	ud, err := s.repo.UserRoleAndRefreshToken(tctx, providedToken.TokenHash)
	if err != nil {
		var bErr *xerr.BaseErr[pr.RepoErrKind]
		if errors.As(err, &bErr) && bErr.Kind == pr.NotFound {
			return "", nil, xerr.WrapErr(op, ps.WrongCredentials, err)
		}
		return "", nil, xerr.WrapErr(op, ps.Unexpected, err)
	}

	providedFp := s.tknSrc.Fingerprint(providedToken)
	actualFp := s.tknSrc.Fingerprint(ud.RToken)
	if subtle.ConstantTimeCompare([]byte(providedFp), []byte(actualFp)) != 1 {
		return "", nil, xerr.NewErr(op, ps.WrongCredentials)
	}

	if ud.RToken.Revoked || time.Until(ud.RToken.ExpiresAt) < 0 {
		return "", nil, xerr.NewErr(op, ps.WrongCredentials)
	}

	userID, err := uuid.Parse(ud.RToken.UserID)
	if err != nil {
		return "", nil, xerr.WrapErr(op, ps.Unexpected, err)
	}

	newAToken, err = s.tknSrc.GenerateAccessToken(auth.AccessTokenData{
		UserID: userID,
		Role:   ud.Role,
	})
	if err != nil {
		return "", nil, xerr.WrapErr(op, ps.Unexpected, err)
	}

	newRToken = s.tknSrc.GenerateRefreshToken(ud.RToken.UserID, ud.RToken.UserAgent, ud.RToken.IP)
	newRToken.TokenHash = s.tknSrc.Hash(newRToken.Token)
	newRToken.Fingerprint = s.tknSrc.Fingerprint(newRToken)

	err = s.repo.UpdateUserRefreshToken(tctx, providedToken.TokenHash, newRToken)
	if err != nil {
		return "", nil, xerr.WrapErr(op, ps.Unexpected, err)
	}

	return newAToken, newRToken, nil
}
