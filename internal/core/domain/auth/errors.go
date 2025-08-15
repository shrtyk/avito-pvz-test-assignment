package auth

import "errors"

var (
	ErrJWTValidation = errors.New("invalid token")
	ErrJWTExpired    = errors.New("jwt expired")

	ErrNotAuthRequest     = errors.New("not authorized")
	ErrInvalidAccessToken = errors.New("invalid token")
)
