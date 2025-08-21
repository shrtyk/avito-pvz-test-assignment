package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/shrtyk/avito-pvz-test-assignment/internal/core/domain"
	pRepo "github.com/shrtyk/avito-pvz-test-assignment/internal/core/ports/repository"
	pService "github.com/shrtyk/avito-pvz-test-assignment/internal/core/ports/service"
	xerr "github.com/shrtyk/avito-pvz-test-assignment/pkg/xerrors"
)

type service struct {
	timeout time.Duration
	repo    pRepo.Repository
}

func NewAppService(timeout time.Duration, repo pRepo.Repository) *service {
	return &service{
		timeout: timeout,
		repo:    repo,
	}
}

func (s *service) NewPVZ(ctx context.Context, pvz *domain.Pvz) (*domain.Pvz, error) {
	op := "service.NewPVZ"

	tctx, tcancel := context.WithTimeout(ctx, s.timeout)
	defer tcancel()

	pvz, err := s.repo.CreatePVZ(tctx, pvz)
	if err != nil {
		return nil, xerr.WrapErr(op, pService.FailedToAddPvz, err)
	}

	return pvz, nil
}

func (s *service) OpenNewPVZReception(ctx context.Context, rec *domain.Reception) (*domain.Reception, error) {
	op := "service.OpenNewPVZReception"

	tctx, tcancel := context.WithTimeout(ctx, s.timeout)
	defer tcancel()

	newRec, err := s.repo.CreateReception(tctx, rec)
	if err != nil {
		var repoErr *xerr.BaseErr[pRepo.RepoErrKind]
		if errors.As(err, &repoErr) {
			switch repoErr.Kind {
			case pRepo.InvalidReference:
				return nil, xerr.WrapErr(op, pService.PvzNotFound, err)
			case pRepo.Conflict:
				return nil, xerr.WrapErr(op, pService.ActiveReceptionExists, err)
			}
		}
		return nil, xerr.WrapErr(op, pService.Unexpected, err)
	}

	return newRec, nil
}

func (s *service) AddProductPVZ(ctx context.Context, prod *domain.Product) (*domain.Product, error) {
	op := "service.AddProductPVZ"

	tctx, tcancel := context.WithTimeout(ctx, s.timeout)
	defer tcancel()

	newProd, err := s.repo.CreateProduct(tctx, prod)
	if err != nil {
		var repoErr *xerr.BaseErr[pRepo.RepoErrKind]
		if errors.As(err, &repoErr) && repoErr.Kind == pRepo.NotFound {
			return nil, xerr.WrapErr(op, pService.NoActiveReception, err)
		}
		return nil, xerr.WrapErr(op, pService.Unexpected, err)
	}

	return newProd, nil
}

func (s *service) DeleteLastProductPvz(ctx context.Context, pvzId *uuid.UUID) error {
	op := "service.DeleteLastProductPvz"

	tctx, tcancel := context.WithTimeout(ctx, s.timeout)
	defer tcancel()

	if err := s.repo.DeleteLastProduct(tctx, pvzId); err != nil {
		var bErr *xerr.BaseErr[pRepo.RepoErrKind]
		if errors.As(err, &bErr) && bErr.Kind == pRepo.NotFound {
			return xerr.WrapErr(op, pService.NoProdOrActiveReception, err)
		}
		return xerr.WrapErr(op, pService.Unexpected, err)
	}

	return nil
}

func (s *service) CloseReceptionInPvz(ctx context.Context, pvzId *uuid.UUID) error {
	op := "service.CloseReceptionInPvz"

	tctx, tcancel := context.WithTimeout(ctx, s.timeout)
	defer tcancel()

	if err := s.repo.CloseReceptionInPvz(tctx, pvzId); err != nil {
		var bErr *xerr.BaseErr[pRepo.RepoErrKind]
		if errors.As(err, &bErr) && bErr.Kind == pRepo.Conflict {
			return xerr.WrapErr(op, pService.FailedToCloseReception, err)
		}
		return xerr.WrapErr(op, pService.Unexpected, err)
	}

	return nil
}

func (s *service) GetPvzsData(ctx context.Context, params *domain.PvzsReadParams) ([]*domain.PvzReceptions, error) {
	op := "service.GetPvzsData"

	tctx, tcancel := context.WithTimeout(ctx, s.timeout)
	defer tcancel()

	res, err := s.repo.GetPvzsData(tctx, params)
	if err != nil {
		return nil, xerr.WrapErr(op, pService.Unexpected, err)
	}

	return res, nil
}

func (s *service) GetAllPvzs(ctx context.Context) ([]*domain.Pvz, error) {
	op := "service.GetAllPvzs"

	tctx, tcancel := context.WithTimeout(ctx, s.timeout)
	defer tcancel()

	res, err := s.GetAllPvzs(tctx)
	if err != nil {
		return nil, xerr.WrapErr(op, pService.Unexpected, err)
	}

	return res, nil
}
