package http

import (
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/shrtyk/avito-backend-spring-2025/internal/adapters/driving/http/dto"
	"github.com/shrtyk/avito-backend-spring-2025/internal/core/domain/auth"

	"github.com/shrtyk/avito-backend-spring-2025/internal/core/ports"
)

type handlers struct {
	appService ports.Service
	tService   ports.TokensService
	validator  *validator.Validate
}

func NewHandlers(appService ports.Service, tService ports.TokensService) *handlers {
	return &handlers{
		appService: appService,
		tService:   tService,
		validator:  validator.New(),
	}
}

func (h *handlers) DummyLogin(w http.ResponseWriter, r *http.Request) error {
	req := new(dto.PostDummyLoginJSONRequestBody)
	if err := ReadJSON(w, r, req); err != nil {
		return BadRequestError(err)
	}

	if err := h.validator.Struct(req); err != nil {
		// Would be better to return StatusUnprocessableEntity but i followed swagger.yaml
		return BadRequestError(err)
	}

	jwt, err := h.tService.GenerateAccessToken(auth.AccessTokenData{
		UserID: 0,
		Role:   auth.UserRole(req.Role),
	})
	if err != nil {
		return err
	}

	err = WriteJSON(w, dto.Token{Jwt: jwt}, http.StatusOK, nil)
	if err != nil {
		return err
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
		return BadRequestError(err)
	}

	newPvz := toDomainPVZ(pvz)
	err = h.appService.NewPVZ(r.Context(), newPvz)
	if err != nil {
		return InternalError(err)
	}

	err = WriteJSON(w, toDTOPVZ(newPvz), http.StatusCreated, nil)
	if err != nil {
		return InternalError(err)
	}

	return nil
}
