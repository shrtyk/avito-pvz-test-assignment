package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/shrtyk/avito-backend-spring-2025/internal/core/domain"
	pRepo "github.com/shrtyk/avito-backend-spring-2025/internal/core/ports/repository"
	pService "github.com/shrtyk/avito-backend-spring-2025/internal/core/ports/service"
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
	tctx, tcancel := context.WithTimeout(ctx, s.timeout)
	defer tcancel()

	pvz, err := s.repo.CreatePVZ(tctx, pvz)
	if err != nil {
		return nil, fmt.Errorf("service: failed to create new pvz: %w", err)
	}

	return pvz, nil
}

func (s *service) NewReception(ctx context.Context, rec *domain.Reception) (*domain.Reception, error) {
	tctx, tcancel := context.WithTimeout(ctx, s.timeout)
	defer tcancel()

	newRec, err := s.repo.CreateReception(tctx, rec)
	if err != nil {
		var cErr *pRepo.ErrConstraintViolation
		if errors.As(err, &cErr) {
			return nil, &pService.ErrReceptionInProgress{
				PvzId: rec.PvzId,
				Err:   err,
			}
		}
		return nil, fmt.Errorf("service: failed to create new reception: %w", err)
	}

	return newRec, nil
}
