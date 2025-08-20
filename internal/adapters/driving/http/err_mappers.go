package http

import (
	"errors"
	"net/http"

	"github.com/shrtyk/avito-pvz-test-assignment/internal/core/ports/auth"
	ps "github.com/shrtyk/avito-pvz-test-assignment/internal/core/ports/service"
	xerr "github.com/shrtyk/avito-pvz-test-assignment/pkg/xerrors"
)

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

func mapAuthServiceErrsToHTTP(err error) *HTTPError {
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
