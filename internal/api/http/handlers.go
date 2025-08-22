package http

import (
	"net/http"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/shrtyk/avito-pvz-test-assignment/internal/api/http/dto"
	"github.com/shrtyk/avito-pvz-test-assignment/internal/core/domain/auth"
	pAuth "github.com/shrtyk/avito-pvz-test-assignment/internal/core/ports/auth"
	pService "github.com/shrtyk/avito-pvz-test-assignment/internal/core/ports/service"
	xerr "github.com/shrtyk/avito-pvz-test-assignment/pkg/xerrors"
)

const (
	refreshTokenKey = "refresh_token"
)

type handlers struct {
	appService   pService.Service
	tokenService pAuth.TokenService
	validator    *validator.Validate
}

func NewHandlers(appService pService.Service, tokenService pAuth.TokenService) *handlers {
	return &handlers{
		appService:   appService,
		tokenService: tokenService,
		validator:    MustNewValidator(),
	}
}

func (h *handlers) HealthZ(w http.ResponseWriter, r *http.Request) error {
	w.WriteHeader(http.StatusOK)
	return nil
}

func (h *handlers) DummyLoginHandler(w http.ResponseWriter, r *http.Request) error {
	req := new(dto.PostDummyLoginJSONRequestBody)
	if err := ReadJson(w, r, req); err != nil {
		return BadRequestBodyError(err)
	}

	if err := h.validator.Struct(req); err != nil {
		return ValidationError(err)
	}

	jwt, err := h.tokenService.GenerateAccessToken(auth.AccessTokenData{
		UserID: uuid.New(),
		Role:   auth.UserRole(req.Role),
	})
	if err != nil {
		return mapTokenServiceErrsToHTTP(err)
	}

	err = WriteJSON(w, dto.Token{Jwt: jwt}, http.StatusOK, nil)
	if err != nil {
		return InternalError(err)
	}

	return nil
}

func (h *handlers) NewPVZHandler(w http.ResponseWriter, r *http.Request) error {
	rBody := new(dto.PostPvzJSONRequestBody)
	err := ReadJson(w, r, rBody)
	if err != nil {
		return BadRequestBodyError(err)
	}

	if err = h.validator.Struct(rBody); err != nil {
		return ValidationError(err)
	}

	newPvz := toDomainPVZ(rBody)
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
	rBody := new(dto.PostReceptionsJSONBody)
	if err := ReadJson(w, r, rBody); err != nil {
		return BadRequestBodyError(err)
	}

	if err := h.validator.Struct(rBody); err != nil {
		return ValidationError(err)
	}

	newRec := toDomainReception(rBody)
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
	rBody := new(dto.PostProductsJSONRequestBody)
	if err := ReadJson(w, r, rBody); err != nil {
		return BadRequestBodyError(err)
	}

	if err := h.validator.Struct(rBody); err != nil {
		return ValidationError(err)
	}

	newProd, err := h.appService.AddProductPVZ(r.Context(), toDomainProduct(rBody))
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

func (h *handlers) RegisterUserHandler(w http.ResponseWriter, r *http.Request) error {
	rBody := new(dto.PostRegisterJSONRequestBody)
	if err := ReadJson(w, r, rBody); err != nil {
		return BadRequestBodyError(err)
	}

	if err := h.validator.Struct(rBody); err != nil {
		return ValidationError(err)
	}

	newUser, err := h.appService.RegisterUser(r.Context(), toDomainUserData(rBody))
	if err != nil {
		return mapAppServiceErrsToHTTP(err)
	}

	err = WriteJSON(w, toDTOUserData(newUser), http.StatusCreated, nil)
	if err != nil {
		return InternalError(err)
	}

	return nil
}

func (h *handlers) LoginUserHandler(w http.ResponseWriter, r *http.Request) error {
	rBody := new(dto.PostLoginJSONRequestBody)
	if err := ReadJson(w, r, rBody); err != nil {
		return BadRequestBodyError(err)
	}

	if err := h.validator.Struct(rBody); err != nil {
		return ValidationError(err)
	}

	ua, ip := UserAgentAndIP(r)
	aToken, rToken, err := h.appService.LoginUser(r.Context(), &auth.LoginUserParams{
		Email:         string(rBody.Email),
		PlainPassword: rBody.Password,
		UserAgent:     ua,
		IP:            ip,
	})
	if err != nil {
		return mapAppServiceErrsToHTTP(err)
	}

	h.setRefreshCookie(w, rToken)

	err = WriteJSON(w, &dto.Token{Jwt: aToken}, http.StatusOK, nil)
	if err != nil {
		return InternalError(err)
	}

	return nil
}

func (h *handlers) RefreshTokensHandler(w http.ResponseWriter, r *http.Request) error {
	rtoken, err := h.getRefreshTokenOutOfCookie(r)
	if err != nil {
		return mapAppServiceErrsToHTTP(err)
	}

	ua, ip := UserAgentAndIP(r)
	newAToken, newRToken, err := h.appService.RefreshTokens(r.Context(), &auth.RefreshToken{
		Token:     rtoken,
		UserAgent: ua,
		IP:        ip,
	})
	if err != nil {
		return mapAppServiceErrsToHTTP(err)
	}

	h.setRefreshCookie(w, newRToken)
	err = WriteJSON(w, &dto.Token{Jwt: newAToken}, http.StatusCreated, nil)
	if err != nil {
		return InternalError(err)
	}

	return nil
}

func (h *handlers) setRefreshCookie(w http.ResponseWriter, rToken *auth.RefreshToken) {
	http.SetCookie(w, &http.Cookie{
		Name:     refreshTokenKey,
		Value:    rToken.Token,
		Path:     "/tokens/refresh",
		MaxAge:   int(time.Until(rToken.ExpiresAt).Seconds()),
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	})
}

func (h *handlers) getRefreshTokenOutOfCookie(r *http.Request) (string, error) {
	op := "handlers.getRefreshTokenOutOfCookie"

	cookie, err := r.Cookie(refreshTokenKey)
	if err != nil {
		return "", xerr.WrapErr(op, pService.WrongCredentials, err)
	}
	return cookie.Value, nil
}
