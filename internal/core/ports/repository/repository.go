package repository

import (
	"context"

	"github.com/shrtyk/avito-backend-spring-2025/internal/core/domain"
)

type Repository interface {
	SavePVZ(ctx context.Context, pvz *domain.PVZ) (*domain.PVZ, error)
	CreateReception(ctx context.Context, rec *domain.Reception) (*domain.Reception, error)
	CreateProduct(ctx context.Context, prod *domain.Product) (*domain.Product, error)
}
