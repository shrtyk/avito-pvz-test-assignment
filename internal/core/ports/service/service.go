package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/shrtyk/avito-pvz-test-assignment/internal/core/domain"
	"github.com/shrtyk/avito-pvz-test-assignment/internal/core/domain/auth"
)

//go:generate mockery
type Service interface {
	NewPVZ(ctx context.Context, pvz *domain.Pvz) (*domain.Pvz, error)
	OpenNewPVZReception(ctx context.Context, rec *domain.Reception) (*domain.Reception, error)
	AddProductPVZ(ctx context.Context, prod *domain.Product) (*domain.Product, error)
	DeleteLastProductPvz(ctx context.Context, pvzId *uuid.UUID) error
	CloseReceptionInPvz(ctx context.Context, pvzId *uuid.UUID) error
	GetPvzsData(ctx context.Context, params *domain.PvzsReadParams) ([]*domain.PvzReceptions, error)
	GetAllPvzs(ctx context.Context) ([]*domain.Pvz, error)
	RegisterUser(ctx context.Context, userParams *auth.RegisterUserParams) (*auth.User, error)
}
