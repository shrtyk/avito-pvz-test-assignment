package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/shrtyk/avito-pvz-test-assignment/internal/core/domain"
	pwd "github.com/shrtyk/avito-pvz-test-assignment/internal/core/ports/pwd_service"
	pr "github.com/shrtyk/avito-pvz-test-assignment/internal/core/ports/repository"
	ps "github.com/shrtyk/avito-pvz-test-assignment/internal/core/ports/service"
	xerr "github.com/shrtyk/avito-pvz-test-assignment/pkg/xerrors"
)

type service struct {
	timeout time.Duration
	repo    pr.Repository
	pwdSrc  pwd.PasswordService
}

func NewAppService(timeout time.Duration, repo pr.Repository, pwdSrc pwd.PasswordService) *service {
	return &service{
		timeout: timeout,
		repo:    repo,
		pwdSrc:  pwdSrc,
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
