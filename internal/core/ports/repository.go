package ports

import (
	"context"

	"github.com/shrtyk/avito-backend-spring-2025/internal/core/domain"
)

type Repository interface {
	CreatePVZ(ctx context.Context, pvz *domain.PVZ) error
}
