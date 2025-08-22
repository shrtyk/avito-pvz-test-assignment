package auth

import (
	"crypto/sha256"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type AccessTokenData struct {
	UserID uuid.UUID
	Role   UserRole
}

type AccessTokenClaims struct {
	Role string `json:"role"`
	jwt.RegisteredClaims
}

func (atc AccessTokenClaims) UserID() string {
	return atc.Subject
}

type RefreshTokenData struct {
	Token     string
	UserID    string
	IP        string
	UserAgent string
	CreatedAt time.Time
	ExpireAt  time.Time
}

func (rtd *RefreshTokenData) PwdHash() []byte {
	h := sha256.Sum256([]byte(rtd.Token))
	return h[:]
}
