package http

import (
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/shrtyk/avito-backend-spring-2025/internal/adapters/driving/http/dto"
	"github.com/shrtyk/avito-backend-spring-2025/internal/core/domain/auth"
	pService "github.com/shrtyk/avito-backend-spring-2025/internal/core/ports/service"

	pAuth "github.com/shrtyk/avito-backend-spring-2025/internal/core/ports/auth"
)

type handlers struct {
	appService pService.Service
	tService   pAuth.TokensService
	validator  *validator.Validate
}

func NewHandlers(appService pService.Service, tService pAuth.TokensService) *handlers {
	return &handlers{
		appService: appService,
		tService:   tService,
		validator:  NewValidator(),
	}
}

func (h *handlers) DummyLoginHandler(w http.ResponseWriter, r *http.Request) error {
	req := new(dto.PostDummyLoginJSONRequestBody)
	if err := ReadJSON(w, r, req); err != nil {
		return BadRequestError(err)
	}

	if err := h.validator.Struct(req); err != nil {
		return ValidationError(err)
	}

	jwt, err := h.tService.GenerateAccessToken(auth.AccessTokenData{
		UserID: 0,
		Role:   auth.UserRole(req.Role),
	})
	if err != nil {
		return MapAppServiceErrsToHTTP(err)
	}

	err = WriteJSON(w, dto.Token{Jwt: jwt}, http.StatusOK, nil)
	if err != nil {
		return InternalError(err)
	}

	return nil
}

func (h *handlers) NewPVZHandler(w http.ResponseWriter, r *http.Request) error {
	pvz := new(dto.PVZ)
	err := ReadJSON(w, r, pvz)
	if err != nil {
		return BadRequestError(err)
	}

	if err = h.validator.Struct(pvz); err != nil {
		return ValidationError(err)
	}

	newPvz := toDomainPVZ(pvz)
	newPvz, err = h.appService.NewPVZ(r.Context(), newPvz)
	if err != nil {
		return MapAppServiceErrsToHTTP(err)
	}

	err = WriteJSON(w, toDTOPVZ(newPvz), http.StatusCreated, nil)
	if err != nil {
		return InternalError(err)
	}

	return nil
}

func (h *handlers) NewReceptionHandler(w http.ResponseWriter, r *http.Request) error {
	rec := new(dto.PostReceptionsJSONBody)
	if err := ReadJSON(w, r, rec); err != nil {
		return BadRequestError(err)
	}

	if err := h.validator.Struct(rec); err != nil {
		return ValidationError(err)
	}

	newRec := toDomainReception(rec)
	newRec, err := h.appService.OpenNewPVZReception(r.Context(), newRec)
	if err != nil {
		return MapAppServiceErrsToHTTP(err)
	}

	err = WriteJSON(w, toDTOReception(newRec), http.StatusCreated, nil)
	if err != nil {
		return InternalError(err)
	}

	return nil
}

func (h *handlers) AddProductHandler(w http.ResponseWriter, r *http.Request) error {
	prod := new(dto.PostProductsJSONRequestBody)
	if err := ReadJSON(w, r, prod); err != nil {
		return BadRequestError(err)
	}

	if err := h.validator.Struct(prod); err != nil {
		return ValidationError(err)
	}

	newProd, err := h.appService.AddProductPVZ(r.Context(), toDomainProduct(prod))
	if err != nil {
		return MapAppServiceErrsToHTTP(err)
	}

	err = WriteJSON(w, toDTOProduct(newProd), http.StatusCreated, nil)
	if err != nil {
		return InternalError(err)
	}

	return nil
}

func (h *handlers) DeleteLastProductHandler(w http.ResponseWriter, r *http.Request) error {
	pvzId, err := ReadPvzIDParam(r)
	if err != nil {
		return BadRequestError(err)
	}

	if err = h.appService.DeleteLastProductPvz(r.Context(), pvzId); err != nil {
		return MapAppServiceErrsToHTTP(err)
	}

	return nil
}
