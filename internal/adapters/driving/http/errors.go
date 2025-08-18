package http

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/shrtyk/avito-backend-spring-2025/internal/core/ports/auth"
	ps "github.com/shrtyk/avito-backend-spring-2025/internal/core/ports/service"
	xerr "github.com/shrtyk/avito-backend-spring-2025/pkg/xerrors"
)

type HTTPError struct {
	Code    int
	Message string
	Err     error
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf("HTTP Code: %d. Message: %s. Error: %s", e.Code, e.Message, e.Err)
}

func (e *HTTPError) Unwrap() error {
	return e.Err
}

func mapAppServiceErrsToHTTP(err error) *HTTPError {
	e := new(HTTPError)

	var bErr *xerr.BaseErr[ps.ServiceErrKind]
	if errors.As(err, &bErr) {
		e.Message = bErr.Kind.String()
		e.Err = bErr

		switch bErr.Kind {
		case ps.Unexpected, ps.FailedToAddPvz:
			e.Code = http.StatusInternalServerError
		case ps.ActiveReceptionExists,
			ps.PvzNotFound,
			ps.NoActiveReception,
			ps.NoProdOrActiveReception,
			ps.FailedToCloseReception:
			e.Code = http.StatusBadRequest
		}
		return e
	}

	return InternalError(err)
}

func MapAuthServiceErrsToHTTP(err error) *HTTPError {
	e := new(HTTPError)

	var bErr *xerr.BaseErr[auth.AuthErrKind]
	if errors.As(err, &bErr) {
		e.Message = bErr.Kind.String()
		e.Err = err

		switch bErr.Kind {
		case auth.JwtCreation, auth.JwtClaimsFromCtx:
			e.Code = http.StatusInternalServerError
		case auth.InvalidJwt, auth.ExpiredJwt, auth.NotAuthenticated:
			e.Code = http.StatusUnauthorized
		case auth.NotAuthorized:
			e.Code = http.StatusForbidden
		}
		return e
	}

	return InternalError(err)
}

func BadRequestError(err error) *HTTPError {
	return &HTTPError{
		Code:    http.StatusBadRequest,
		Message: "Badly formed request body",
		Err:     err,
	}
}

func InternalError(err error) *HTTPError {
	return &HTTPError{
		Code:    http.StatusInternalServerError,
		Message: "Internal error",
		Err:     err,
	}
}

func ValidationError(err error) *HTTPError {
	return &HTTPError{
		// Would be better to use StatusUnprocessableEntity but i followed swagger.yaml
		Code:    http.StatusBadRequest,
		Message: "Invalid data in request body",
		Err:     err,
	}
}
