package http

import (
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/shrtyk/avito-backend-spring-2025/internal/adapters/driving/http/dto"
	"github.com/shrtyk/avito-backend-spring-2025/internal/core/domain/auth"
	"github.com/shrtyk/avito-backend-spring-2025/internal/core/domain/users"
	"github.com/shrtyk/avito-backend-spring-2025/internal/core/ports"
)

type handlers struct {
	tService  ports.TokensService
	validator *validator.Validate
}

func NewHandlers(tService ports.TokensService) *handlers {
	return &handlers{
		tService:  tService,
		validator: validator.New(),
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
		Role:   users.UserRole(req.Role),
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
