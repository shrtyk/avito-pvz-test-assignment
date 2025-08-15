package auth

import "errors"

var (
	ErrInvalidJWT = errors.New("invalid token")
	ErrExpiredJWT = errors.New("jwt expired")

	ErrNotAuthRequest = errors.New("not authorized")
)
