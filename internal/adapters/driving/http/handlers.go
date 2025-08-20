package http

import (
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/shrtyk/avito-pvz-test-assignment/internal/adapters/driving/http/dto"
	"github.com/shrtyk/avito-pvz-test-assignment/internal/core/domain/auth"
	pService "github.com/shrtyk/avito-pvz-test-assignment/internal/core/ports/service"

	pAuth "github.com/shrtyk/avito-pvz-test-assignment/internal/core/ports/auth"
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
	if err := ReadJson(w, r, req); err != nil {
		return BadRequestBodyError(err)
	}

	if err := h.validator.Struct(req); err != nil {
		return ValidationError(err)
	}

	jwt, err := h.tService.GenerateAccessToken(auth.AccessTokenData{
		UserID: 0,
		Role:   auth.UserRole(req.Role),
	})
	if err != nil {
		return mapAuthServiceErrsToHTTP(err)
	}

	err = WriteJSON(w, dto.Token{Jwt: jwt}, http.StatusOK, nil)
	if err != nil {
		return InternalError(err)
	}

	return nil
}

func (h *handlers) NewPVZHandler(w http.ResponseWriter, r *http.Request) error {
	pvz := new(dto.PVZ)
	err := ReadJson(w, r, pvz)
	if err != nil {
		return BadRequestBodyError(err)
	}

	if err = h.validator.Struct(pvz); err != nil {
		return ValidationError(err)
	}

	newPvz := toDomainPVZ(pvz)
	newPvz, err = h.appService.NewPVZ(r.Context(), newPvz)
	if err != nil {
		return mapAppServiceErrsToHTTP(err)
	}

	err = WriteJSON(w, toDTOPVZ(newPvz), http.StatusCreated, nil)
	if err != nil {
		return InternalError(err)
	}

	return nil
}

func (h *handlers) NewReceptionHandler(w http.ResponseWriter, r *http.Request) error {
	rec := new(dto.PostReceptionsJSONBody)
	if err := ReadJson(w, r, rec); err != nil {
		return BadRequestBodyError(err)
	}

	if err := h.validator.Struct(rec); err != nil {
		return ValidationError(err)
	}

	newRec := toDomainReception(rec)
	newRec, err := h.appService.OpenNewPVZReception(r.Context(), newRec)
	if err != nil {
		return mapAppServiceErrsToHTTP(err)
	}

	err = WriteJSON(w, toDTOReception(newRec), http.StatusCreated, nil)
	if err != nil {
		return InternalError(err)
	}

	return nil
}

func (h *handlers) AddProductHandler(w http.ResponseWriter, r *http.Request) error {
	prod := new(dto.PostProductsJSONRequestBody)
	if err := ReadJson(w, r, prod); err != nil {
		return BadRequestBodyError(err)
	}

	if err := h.validator.Struct(prod); err != nil {
		return ValidationError(err)
	}

	newProd, err := h.appService.AddProductPVZ(r.Context(), toDomainProduct(prod))
	if err != nil {
		return mapAppServiceErrsToHTTP(err)
	}

	err = WriteJSON(w, toDTOProduct(newProd), http.StatusCreated, nil)
	if err != nil {
		return InternalError(err)
	}

	return nil
}

func (h *handlers) DeleteLastProductHandler(w http.ResponseWriter, r *http.Request) error {
	pvzId, err := PvzIdParam(r)
	if err != nil {
		return BadRequestBodyError(err)
	}

	if err = h.appService.DeleteLastProductPvz(r.Context(), pvzId); err != nil {
		return mapAppServiceErrsToHTTP(err)
	}

	return nil
}

func (h *handlers) CloseReceptionHandler(w http.ResponseWriter, r *http.Request) error {
	pvzId, err := PvzIdParam(r)
	if err != nil {
		return BadRequestBodyError(err)
	}

	if err := h.appService.CloseReceptionInPvz(r.Context(), pvzId); err != nil {
		return mapAppServiceErrsToHTTP(err)
	}

	return nil
}

func (h *handlers) GetPvzHandler(w http.ResponseWriter, r *http.Request) error {
	params, err := PvzParamsFromURL(r)
	if err != nil {
		return BadRequestQueryParamsError(err)
	}

	domainParams := toDomainPvzReadParams(params)

	pvzsData, err := h.appService.GetPvzsData(r.Context(), domainParams)
	if err != nil {
		return mapAppServiceErrsToHTTP(err)
	}

	resp := toDTOPvzData(pvzsData)
	if err = WriteJSON(w, resp, http.StatusOK, nil); err != nil {
		return InternalError(err)
	}

	return nil
}
