package http

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shrtyk/avito-pvz-test-assignment/internal/adapters/driving/http/dto"
	"github.com/shrtyk/avito-pvz-test-assignment/internal/core/domain"
	"github.com/stretchr/testify/assert"
)

func Test_toDomainPVZ(t *testing.T) {
	t.Parallel()

	t.Run("nil dto", func(t *testing.T) {
		t.Parallel()
		assert.Nil(t, toDomainPVZ(nil))
	})

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		dtoPvz := &dto.PVZ{
			City: "Moscow",
		}
		domainPvz := toDomainPVZ(dtoPvz)
		assert.Equal(t, domain.PVZCity("Moscow"), domainPvz.City)
	})
}

func Test_toDTOPVZ(t *testing.T) {
	t.Parallel()

	t.Run("nil domain", func(t *testing.T) {
		t.Parallel()
		assert.Nil(t, toDTOPVZ(nil))
	})

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		pvzID := uuid.New()
		domainPvz := &domain.Pvz{
			Id:               pvzID,
			RegistrationDate: time.Now(),
			City:             "Moscow",
		}
		dtoPvz := toDTOPVZ(domainPvz)
		assert.Equal(t, &pvzID, dtoPvz.Id)
		assert.Equal(t, dto.PVZCity("Moscow"), dtoPvz.City)
	})
}

func Test_toDomainPvzReadParams(t *testing.T) {
	t.Parallel()

	t.Run("nil params", func(t *testing.T) {
		t.Parallel()
		domainParams := toDomainPvzReadParams(nil)
		assert.Equal(t, defaultPage, domainParams.Page)
		assert.Equal(t, defaultLimit, domainParams.Limit)
	})

	t.Run("with params", func(t *testing.T) {
		t.Parallel()
		page := 2
		limit := 20
		startDate := time.Now()
		endDate := time.Now().Add(time.Hour)
		dtoParams := &dto.GetPvzParams{
			Page:      &page,
			Limit:     &limit,
			StartDate: &startDate,
			EndDate:   &endDate,
		}
		domainParams := toDomainPvzReadParams(dtoParams)
		assert.Equal(t, page, domainParams.Page)
		assert.Equal(t, limit, domainParams.Limit)
		assert.Equal(t, &startDate, domainParams.StartDate)
		assert.Equal(t, &endDate, domainParams.EndDate)
	})
}

func Test_toDomainReception(t *testing.T) {
	t.Parallel()

	t.Run("nil dto", func(t *testing.T) {
		t.Parallel()
		assert.Nil(t, toDomainReception(nil))
	})

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		pvzID := uuid.New()
		dtoRec := &dto.PostReceptionsJSONBody{
			PvzId: pvzID,
		}
		domainRec := toDomainReception(dtoRec)
		assert.Equal(t, pvzID, domainRec.PvzId)
	})
}

func Test_toDTOReception(t *testing.T) {
	t.Parallel()

	t.Run("nil domain", func(t *testing.T) {
		t.Parallel()
		assert.Nil(t, toDTOReception(nil))
	})

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		recID := uuid.New()
		pvzID := uuid.New()
		domainRec := &domain.Reception{
			Id:       recID,
			PvzId:    pvzID,
			Status:   domain.InProgress,
			DateTime: time.Now(),
		}
		dtoRec := toDTOReception(domainRec)
		assert.Equal(t, &recID, dtoRec.Id)
		assert.Equal(t, pvzID, dtoRec.PvzId)
		assert.Equal(t, dto.ReceptionStatus(domain.InProgress), dtoRec.Status)
	})
}

func Test_toDomainProduct(t *testing.T) {
	t.Parallel()

	t.Run("nil dto", func(t *testing.T) {
		t.Parallel()
		assert.Nil(t, toDomainProduct(nil))
	})

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		pvzID := uuid.New()
		dtoProd := &dto.PostProductsJSONRequestBody{
			PvzId: pvzID,
			Type:  dto.PostProductsJSONBodyTypeClothing,
		}
		domainProd := toDomainProduct(dtoProd)
		assert.Equal(t, pvzID, domainProd.PvzId)
		assert.Equal(t, domain.ProductType(dto.PostProductsJSONBodyTypeClothing), domainProd.Type)
	})
}

func Test_toDTOProduct(t *testing.T) {
	t.Parallel()

	t.Run("nil domain", func(t *testing.T) {
		t.Parallel()
		assert.Nil(t, toDTOProduct(nil))
	})

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		prodID := uuid.New()
		recID := uuid.New()
		domainProd := &domain.Product{
			Id:          prodID,
			ReceptionId: recID,
			DateTime:    time.Now(),
			Type:        domain.ProductTypeClothing,
		}
		dtoProd := toDTOProduct(domainProd)
		assert.Equal(t, &prodID, dtoProd.Id)
		assert.Equal(t, recID, dtoProd.ReceptionId)
		assert.Equal(t, dto.ProductType(domain.ProductTypeClothing), dtoProd.Type)
	})
}

func Test_toDTOReceptionProducts(t *testing.T) {
	t.Parallel()

	t.Run("nil domain", func(t *testing.T) {
		t.Parallel()
		assert.Nil(t, toDTOReceptionProducts(nil))
	})

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		domainRecProd := &domain.ReceptionProducts{
			Reception: &domain.Reception{},
			Products:  []*domain.Product{{}},
		}
		dtoRecProd := toDTOReceptionProducts(domainRecProd)
		assert.NotNil(t, dtoRecProd.Reception)
		assert.Len(t, *dtoRecProd.Products, 1)
	})
}

func Test_toDTOPvzReceptionsProducts(t *testing.T) {
	t.Parallel()

	t.Run("nil domain", func(t *testing.T) {
		t.Parallel()
		assert.Nil(t, toDTOPvzReceptions(nil))
	})

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		domainPvzRecProd := &domain.PvzReceptions{
			Pvz:        &domain.Pvz{},
			Receptions: []*domain.ReceptionProducts{{}},
		}
		dtoPvzRecProd := toDTOPvzReceptions(domainPvzRecProd)
		assert.NotNil(t, dtoPvzRecProd.Pvz)
		assert.Len(t, *dtoPvzRecProd.Receptions, 1)
	})
}

func Test_toDTOPvzData(t *testing.T) {
	t.Parallel()

	t.Run("nil domain", func(t *testing.T) {
		t.Parallel()
		assert.NotNil(t, toDTOPvzData(nil))
	})

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		domainPvzData := []*domain.PvzReceptions{{}}
		dtoPvzData := toDTOPvzData(domainPvzData)
		assert.Len(t, dtoPvzData, 1)
	})
}
