package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/shrtyk/avito-pvz-test-assignment/internal/core/domain"
	"github.com/shrtyk/avito-pvz-test-assignment/internal/core/domain/auth"
)

//go:generate mockery
type Repository interface {
	PvzsRepo
	AuthRepo
}

type PvzsRepo interface {
	CreatePVZ(ctx context.Context, pvz *domain.Pvz) (*domain.Pvz, error)
	CreateReception(ctx context.Context, rec *domain.Reception) (*domain.Reception, error)
	CreateProduct(ctx context.Context, prod *domain.Product) (*domain.Product, error)
	DeleteLastProduct(ctx context.Context, pvzId *uuid.UUID) error
	CloseReceptionInPvz(ctx context.Context, pvzId *uuid.UUID) error
	GetPvzsData(ctx context.Context, params *domain.PvzsReadParams) ([]*domain.PvzReceptions, error)
	GetAllPvzs(ctx context.Context) ([]*domain.Pvz, error)
}

type AuthRepo interface {
	GetUserByEmail(ctx context.Context, email string) (*auth.User, error)
	CreateUser(ctx context.Context, user *auth.User) (*auth.User, error)
	SaveRefreshToken(ctx context.Context, rToken *auth.RefreshTokenData, fp string) error
}
