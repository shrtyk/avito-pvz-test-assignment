package auth

import (
	"github.com/golang-jwt/jwt/v5"
	"github.com/shrtyk/avito-backend-spring-2025/internal/core/domain/users"
)

type AccessTokenData struct {
	UserID int64
	Role   users.UserRole
}

type AccessTokenClaims struct {
	Role string `json:"role"`
	jwt.RegisteredClaims
}
