package http

import (
	"github.com/shrtyk/avito-backend-spring-2025/internal/adapters/driving/http/dto"
	"github.com/shrtyk/avito-backend-spring-2025/internal/core/domain"
)

func toDomainPVZ(dtoPvz *dto.PVZ) *domain.PVZ {
	return &domain.PVZ{
		City: domain.PVZCity(dtoPvz.City),
	}
}

func toDTOPVZ(domainPvz *domain.PVZ) *dto.PVZ {
	return &dto.PVZ{
		Id:               &domainPvz.Id,
		RegistrationDate: &domainPvz.RegistrationDate,
		City:             dto.PVZCity(domainPvz.City),
	}
}

func toDomainReception(dtoRec *dto.PostReceptionsJSONBody) *domain.Reception {
	return &domain.Reception{
		PvzId: dtoRec.PvzId,
	}
}

func toDTOReception(domainRec *domain.Reception) *dto.Reception {
	return &dto.Reception{
		Id:       &domainRec.Id,
		PvzId:    domainRec.PvzId,
		Status:   dto.ReceptionStatus(domainRec.Status),
		DateTime: domainRec.DateTime,
	}
}

func toDomainProduct(dtoProd *dto.PostProductsJSONRequestBody) *domain.Product {
	return &domain.Product{
		PvzId: dtoProd.PvzId,
		Type:  domain.ProductType(dtoProd.Type),
	}
}

func toDTOProduct(domainProd *domain.Product) *dto.Product {
	return &dto.Product{
		Id:          &domainProd.Id,
		ReceptionId: domainProd.ReceptionId,
		DateTime:    &domainProd.DateTime,
		Type:        dto.ProductType(domainProd.Type),
	}
}
