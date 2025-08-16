package service

import (
	"context"
	"time"

	"github.com/shrtyk/avito-backend-spring-2025/internal/core/domain"
	"github.com/shrtyk/avito-backend-spring-2025/internal/core/ports"
)

type service struct {
	timeout time.Duration
	repo    ports.Repository
}

func NewAppService(timeout time.Duration, repo ports.Repository) *service {
	return &service{
		timeout: timeout,
		repo:    repo,
	}
}

func (s *service) NewPVZ(ctx context.Context, pvz *domain.PVZ) error {
	tctx, tcancel := context.WithTimeout(ctx, s.timeout)
	defer tcancel()

	if err := s.repo.CreatePVZ(tctx, pvz); err != nil {
		return err
	}
	return nil
}
