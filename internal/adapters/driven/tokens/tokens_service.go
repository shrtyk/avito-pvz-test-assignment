package tokens

import (
	"crypto/rsa"
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/shrtyk/avito-backend-spring-2025/internal/core/domain/auth"
	"github.com/shrtyk/avito-backend-spring-2025/pkg/config"
)

type tokensService struct {
	publicKey           *rsa.PublicKey
	privateKey          *rsa.PrivateKey
	accessTokenLifetime time.Duration
}

func MustCreateTokenService(cfg *config.AuthTokensCfg) *tokensService {
	pubKeyData, err := os.ReadFile(cfg.PublicRSAPath)
	if err != nil {
		msg := fmt.Sprintf("failed to read: %s: %s", cfg.PublicRSAPath, err)
		panic(msg)
	}

	pub, err := jwt.ParseRSAPublicKeyFromPEM(pubKeyData)
	if err != nil {
		msg := fmt.Sprintf("failed to parse public key: %s", err)
		panic(msg)
	}

	privateKeyData, err := os.ReadFile(cfg.PrivateRSAPath)
	if err != nil {
		msg := fmt.Sprintf("failed to read: %s: %s", cfg.PrivateRSAPath, err)
		panic(msg)
	}

	private, err := jwt.ParseRSAPrivateKeyFromPEM(privateKeyData)
	if err != nil {
		msg := fmt.Sprintf("failed to parse private key: %s", err)
		panic(msg)
	}

	return &tokensService{
		publicKey:           pub,
		privateKey:          private,
		accessTokenLifetime: cfg.JWTLifetime,
	}
}

func (s *tokensService) GenerateAccessToken(tokenData auth.AccessTokenData) (string, error) {
	claims := auth.AccessTokenClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   strconv.FormatInt(tokenData.UserID, 10),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.accessTokenLifetime)),
			ID:        uuid.NewString(),
		},
		Role: string(tokenData.Role),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	return token.SignedString(s.privateKey)
}

func (s *tokensService) GetTokenClaims(token string) (*auth.AccessTokenClaims, error) {
	tokenClaims := new(auth.AccessTokenClaims)
	t, err := jwt.ParseWithClaims(token, tokenClaims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return s.publicKey, nil
	})
	if err != nil {
		switch {
		case errors.Is(err, jwt.ErrTokenExpired):
			return nil, auth.ErrExpiredJWT
		default:
			return nil, auth.ErrInvalidJWT
		}
	}

	if !t.Valid {
		return nil, auth.ErrInvalidJWT
	}

	return tokenClaims, nil
}
