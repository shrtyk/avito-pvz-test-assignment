package http

import (
	"github.com/oapi-codegen/runtime/types"
	"github.com/shrtyk/avito-backend-spring-2025/internal/adapters/driving/http/dto"
	"github.com/shrtyk/avito-backend-spring-2025/internal/core/domain"
)

func toDomainPVZ(dtoPvz *dto.PVZ) *domain.PVZ {
	return &domain.PVZ{
		City: domain.PVZCity(dtoPvz.City),
	}
}

func toDTOPVZ(domainPvz *domain.PVZ) *dto.PVZ {
	dtoID := types.UUID([]byte(domainPvz.Id.String()))
	return &dto.PVZ{
		Id:               &dtoID,
		RegistrationDate: &domainPvz.RegistrationDate,
		City:             dto.PVZCity(domainPvz.City),
	}
}
