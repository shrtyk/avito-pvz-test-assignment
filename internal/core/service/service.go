package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/shrtyk/avito-backend-spring-2025/internal/core/domain"
	pRepo "github.com/shrtyk/avito-backend-spring-2025/internal/core/ports/repository"
	pService "github.com/shrtyk/avito-backend-spring-2025/internal/core/ports/service"
	xerr "github.com/shrtyk/avito-backend-spring-2025/pkg/xerrors"
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

func (s *service) NewPVZ(ctx context.Context, pvz *domain.PVZ) (*domain.PVZ, error) {
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
