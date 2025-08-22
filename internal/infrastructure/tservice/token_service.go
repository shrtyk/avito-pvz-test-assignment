package tservice

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/shrtyk/avito-pvz-test-assignment/internal/core/domain/auth"
	pa "github.com/shrtyk/avito-pvz-test-assignment/internal/core/ports/auth"
	"github.com/shrtyk/avito-pvz-test-assignment/pkg/config"
	xerr "github.com/shrtyk/avito-pvz-test-assignment/pkg/xerrors"
)

type tokenService struct {
	publicKey  *rsa.PublicKey
	privateKey *rsa.PrivateKey
	cfg        *config.AuthTokensCfg
}

func MustCreateTokenService(cfg *config.AuthTokensCfg) *tokenService {
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

	return &tokenService{
		publicKey:  pub,
		privateKey: private,
		cfg:        cfg,
	}
}

func (s *tokenService) GenerateAccessToken(tokenData auth.AccessTokenData) (string, error) {
	claims := auth.AccessTokenClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   tokenData.UserID.String(),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.cfg.JWTLifetime)),
			ID:        uuid.NewString(),
		},
		Role: string(tokenData.Role),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	return token.SignedString(s.privateKey)
}

func (s *tokenService) GenerateRefreshToken(userID, ua, ip string) *auth.RefreshToken {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		// Shouldn't occur at all
		panic("failed to generate refresh token: " + err.Error())
	}
	return &auth.RefreshToken{
		Token:     base64.URLEncoding.EncodeToString(b),
		UserID:    userID,
		UserAgent: ua,
		IP:        ip,
		CreatedAt: time.Now(),
		ExpiresAt:  time.Now().Add(s.cfg.RefreshLifetime),
	}
}

func (s *tokenService) hash(token string) []byte {
	hash := sha256.Sum256([]byte(token))
	return hash[:]
}

func (s *tokenService) Fingerprint(rToken *auth.RefreshToken) string {
	templ := fmt.Sprintf("%s.%s.%s.%s", rToken.Token, rToken.UserAgent, rToken.IP, s.cfg.SecretKey)
	h := s.hash(templ)
	return hex.EncodeToString(h)
}

func (s *tokenService) GetTokenClaims(token string) (*auth.AccessTokenClaims, error) {
	op := "token_service.GetTokenClaims"

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
			return nil, xerr.WrapErr(op, pa.ExpiredJwt, err)
		default:
			return nil, xerr.WrapErr(op, pa.InvalidJwt, err)
		}
	}

	if !t.Valid {
		return nil, xerr.WrapErr(op, pa.InvalidJwt, err)
	}

	return tokenClaims, nil
}
