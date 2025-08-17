package middleware

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"
	appHttp "github.com/shrtyk/avito-backend-spring-2025/internal/adapters/driving/http"
	"github.com/shrtyk/avito-backend-spring-2025/internal/adapters/driving/http/dto"
	"github.com/shrtyk/avito-backend-spring-2025/internal/core/domain/auth"
	pAuth "github.com/shrtyk/avito-backend-spring-2025/internal/core/ports/auth"
	"github.com/shrtyk/avito-backend-spring-2025/pkg/logger"
)

type Middlewares struct {
	tService pAuth.TokensService
	log      *slog.Logger
}

func NewMiddlewares(
	tService pAuth.TokensService,
	log *slog.Logger,
) *Middlewares {
	return &Middlewares{
		tService: tService,
		log:      log,
	}
}

func (m Middlewares) PanicRecoveryMW(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				m.log.Error("Error occured", "error", fmt.Sprintf("%s", err))
				w.Header().Set("Connection", "close")
				appHttp.WriteHTTPError(w, r, &appHttp.HTTPError{
					Code:    http.StatusInternalServerError,
					Message: "The server encountered a problem and could not process your request",
					Err:     fmt.Errorf("%s", err),
				})
			}
		}()

		next.ServeHTTP(w, r)
	})
}

type customResponseWriter struct {
	http.ResponseWriter
	statusCode  int
	wroteHeader bool
}

func (w *customResponseWriter) WriteHeader(statusCode int) {
	if w.wroteHeader {
		return
	}

	w.statusCode = statusCode
	w.wroteHeader = true
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *customResponseWriter) Write(b []byte) (int, error) {
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}
	return w.ResponseWriter.Write(b)
}

func (m Middlewares) LoggingMW(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ua, ip := appHttp.GetUserAgentAndIP(r)
		reqID := uuid.NewString()

		l := m.log.With(
			slog.String("ip", ip),
			slog.String("user_agent", ua),
			slog.String("request_id", reqID),
			slog.String("method", r.Method),
			slog.String("uri", r.URL.RequestURI()),
		)

		l.Debug("New HTTP request")
		newCtx := logger.ToCtx(r.Context(), l)
		newReq := r.WithContext(newCtx)
		custWriter := &customResponseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		reqStart := time.Now()
		next.ServeHTTP(custWriter, newReq)
		reqEnd := time.Since(reqStart)

		ttp := fmt.Sprintf("%.5fs", reqEnd.Seconds())
		l.Debug(
			"HTTP request processed",
			slog.Int("status_code", custWriter.statusCode),
			slog.String("request_duration", ttp),
		)
	})
}

func (m Middlewares) AuthenticationMW(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Vary", "Authorization")

		bt, err := appHttp.ExtractBearerToken(r)
		if err != nil {
			switch {
			case errors.Is(err, auth.ErrNotAuthenticated):
				appHttp.WriteHTTPError(w, r, &appHttp.HTTPError{
					Code:    http.StatusUnauthorized,
					Message: "Wrong credentials provided",
					Err:     err,
				})
			case errors.Is(err, auth.ErrInvalidJWT):
				appHttp.WriteHTTPError(w, r, &appHttp.HTTPError{
					Code:    http.StatusUnauthorized,
					Message: "Invalid authentication token",
					Err:     err,
				})
			default:
				appHttp.WriteHTTPError(w, r, appHttp.InternalError(err))
			}
			return
		}

		claims, err := m.tService.GetTokenClaims(bt)
		if err != nil {
			switch {
			case errors.Is(err, auth.ErrInvalidJWT):
				appHttp.WriteHTTPError(w, r, &appHttp.HTTPError{
					Code:    http.StatusUnauthorized,
					Message: "Invalid JWT",
					Err:     err,
				})
			case errors.Is(err, auth.ErrExpiredJWT):
				appHttp.WriteHTTPError(w, r, &appHttp.HTTPError{
					Code:    http.StatusUnauthorized,
					Message: "JWT expired",
					Err:     err,
				})
			default:
				appHttp.WriteHTTPError(w, r, appHttp.InternalError(err))
			}
			return
		}

		l := logger.FromCtx(r.Context())
		newLog := l.With(slog.String("user_id", claims.UserID()))
		ctxWithLog := logger.ToCtx(r.Context(), newLog)
		ctxWithClaims := auth.ClaimsToCtx(ctxWithLog, claims)

		newReq := r.WithContext(ctxWithClaims)
		next.ServeHTTP(w, newReq)
	})
}

func (m Middlewares) ModeratorAuthMW(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, err := auth.ClaimsFromCtx(r.Context())
		if err != nil {
			appHttp.WriteHTTPError(w, r, appHttp.InternalError(err))
			return
		}

		if claims.Role != string(dto.Moderator) {
			appHttp.WriteHTTPError(w, r, &appHttp.HTTPError{
				Code:    http.StatusForbidden,
				Message: "Not authorized to access this endpoint",
				Err:     auth.ErrNotAuthorized,
			})
			return
		}

		next.ServeHTTP(w, r)
	})
}
