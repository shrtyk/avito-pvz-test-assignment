package auth

import "errors"

var (
	ErrInvalidJWT = errors.New("invalid token")
	ErrExpiredJWT = errors.New("jwt expired")

	ErrNotAuthenticated = errors.New("not authenticated")
	ErrNotAuthorized    = errors.New("not authorized")

	ErrClaimsFormCtx = errors.New("failed to get JWT claims from context")
)
