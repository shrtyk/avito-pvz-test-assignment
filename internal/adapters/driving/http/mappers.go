package http

import (
	"github.com/shrtyk/avito-backend-spring-2025/internal/adapters/driving/http/dto"
	"github.com/shrtyk/avito-backend-spring-2025/internal/core/domain"
)

const (
	defaultPage  = 1
	defaultLimit = 10
)

func toDomainPVZ(dtoPvz *dto.PVZ) *domain.Pvz {
	if dtoPvz == nil {
		return nil
	}

	return &domain.Pvz{
		City: domain.PVZCity(dtoPvz.City),
	}
}

func toDTOPVZ(domainPvz *domain.Pvz) *dto.PVZ {
	if domainPvz == nil {
		return nil
	}

	return &dto.PVZ{
		Id:               &domainPvz.Id,
		RegistrationDate: &domainPvz.RegistrationDate,
		City:             dto.PVZCity(domainPvz.City),
	}
}

func toDomainReception(dtoRec *dto.PostReceptionsJSONBody) *domain.Reception {
	if dtoRec == nil {
		return nil
	}

	return &domain.Reception{
		PvzId: dtoRec.PvzId,
	}
}

func toDTOReception(domainRec *domain.Reception) *dto.Reception {
	if domainRec == nil {
		return nil
	}

	return &dto.Reception{
		Id:       &domainRec.Id,
		PvzId:    domainRec.PvzId,
		Status:   dto.ReceptionStatus(domainRec.Status),
		DateTime: domainRec.DateTime,
	}
}

func toDomainProduct(dtoProd *dto.PostProductsJSONRequestBody) *domain.Product {
	if dtoProd == nil {
		return nil
	}

	return &domain.Product{
		PvzId: dtoProd.PvzId,
		Type:  domain.ProductType(dtoProd.Type),
	}
}

func toDTOProduct(domainProd *domain.Product) *dto.Product {
	if domainProd == nil {
		return nil
	}

	return &dto.Product{
		Id:          &domainProd.Id,
		ReceptionId: domainProd.ReceptionId,
		DateTime:    &domainProd.DateTime,
		Type:        dto.ProductType(domainProd.Type),
	}
}

func toDomainPvzReadParams(dtoParams *dto.GetPvzParams) *domain.PvzsReadParams {
	domainParams := &domain.PvzsReadParams{
		Page:  defaultPage,
		Limit: defaultLimit,
	}

	if dtoParams == nil {
		return domainParams
	}

	if dtoParams.Limit != nil && *dtoParams.Limit >= 1 && *dtoParams.Limit <= 30 {
		domainParams.Limit = *dtoParams.Limit
	}

	if dtoParams.Page == nil && *dtoParams.Page >= 1 {
		domainParams.Page = *dtoParams.Page
	}

	domainParams.StartDate = dtoParams.StartDate
	domainParams.EndDate = dtoParams.EndDate

	return domainParams
}

func toDTOReceptionProducts(dm *domain.ReceptionProducts) *dto.ReceptionProducts {
	if dm == nil {
		return nil
	}

	dt := &dto.ReceptionProducts{
		Reception: toDTOReception(dm.Reception),
	}

	if dm.Products != nil {
		p := make([]dto.Product, len(dm.Products))
		for i, dmp := range dm.Products {
			if dtoProduct := toDTOProduct(dmp); dtoProduct != nil {
				p[i] = *dtoProduct
			}
		}
		dt.Products = &p
	}
	return dt
}

func toDTOPvzReceptionsProducts(dm *domain.PvzReceptionsProducts) *dto.PvzReceptions {
	if dm == nil {
		return nil
	}

	dt := &dto.PvzReceptions{
		Pvz: toDTOPVZ(dm.Pvz),
	}

	if dm.Receptions != nil {
		recs := make([]dto.ReceptionProducts, len(dm.Receptions))
		for i, r := range dm.Receptions {
			if dtoRec := toDTOReceptionProducts(r); dtoRec != nil {
				recs[i] = *dtoRec
			}
		}
		dt.Receptions = &recs
	}

	return dt
}

func toDTOPvzData(dd []*domain.PvzReceptionsProducts) []*dto.PvzReceptions {
	res := make([]*dto.PvzReceptions, len(dd))
	if dd == nil {
		return res
	}

	for i, d := range dd {
		res[i] = toDTOPvzReceptionsProducts(d)
	}
	return res
}
