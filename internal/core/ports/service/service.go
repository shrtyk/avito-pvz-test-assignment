package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/shrtyk/avito-backend-spring-2025/internal/core/domain"
)

type Service interface {
	NewPVZ(ctx context.Context, pvz *domain.PVZ) (*domain.PVZ, error)
	OpenNewPVZReception(ctx context.Context, rec *domain.Reception) (*domain.Reception, error)
	AddProductPVZ(ctx context.Context, prod *domain.Product) (*domain.Product, error)
	DeleteLastProductPvz(ctx context.Context, pvzId *uuid.UUID) error
}
