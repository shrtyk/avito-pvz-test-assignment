package ports

import (
	"context"

	"github.com/shrtyk/avito-backend-spring-2025/internal/core/domain"
)

type Service interface {
	NewPVZ(ctx context.Context, pvz *domain.PVZ) error
}
