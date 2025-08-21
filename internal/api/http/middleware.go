package http

import (
	"fmt"
	"log/slog"
	"net/http"
	"slices"
	"time"

	"github.com/google/uuid"
	"github.com/shrtyk/avito-pvz-test-assignment/internal/core/domain/auth"
	pAuth "github.com/shrtyk/avito-pvz-test-assignment/internal/core/ports/auth"
	tservice "github.com/shrtyk/avito-pvz-test-assignment/internal/infrastructure/token_service"
	"github.com/shrtyk/avito-pvz-test-assignment/pkg/logger"
	xerr "github.com/shrtyk/avito-pvz-test-assignment/pkg/xerrors"
)

type Middlewares struct {
	tService      pAuth.TokensService
	log           *slog.Logger
	handleAuthErr func(http.ResponseWriter, *http.Request, error)
}

func NewMiddlewares(
	tService pAuth.TokensService,
	log *slog.Logger,
) *Middlewares {
	eh := func(w http.ResponseWriter, r *http.Request, err error) {
		WriteHTTPError(w, r, mapAuthServiceErrsToHTTP(err))
	}

	return &Middlewares{
		tService:      tService,
		log:           log,
		handleAuthErr: eh,
	}
}

func (m Middlewares) PanicRecoveryMW(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				m.log.Error(
					"Error occurred",
					"error", fmt.Sprintf("%s", err),
				)
				w.Header().Set("Connection", "close")
				WriteHTTPError(w, r, &HTTPError{
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
		ua, ip := UserAgentAndIP(r)
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

		bt, err := BearerToken(r)
		if err != nil {
			m.handleAuthErr(w, r, err)
			return
		}

		claims, err := m.tService.GetTokenClaims(bt)
		if err != nil {
			m.handleAuthErr(w, r, err)
			return
		}

		l := logger.FromCtx(r.Context())
		newLog := l.With(slog.String("user_id", claims.UserID()))
		ctxWithLog := logger.ToCtx(r.Context(), newLog)
		ctxWithClaims := tservice.ClaimsToCtx(ctxWithLog, claims)

		newReq := r.WithContext(ctxWithClaims)
		next.ServeHTTP(w, newReq)
	})
}

func (m *Middlewares) AuthorizeRoles(allowedRoles ...auth.UserRole) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			op := "middlewares.AuthorizeRoles"

			claims, err := tservice.ClaimsFromCtx(r.Context())
			if err != nil {
				m.handleAuthErr(w, r, err)
				return
			}

			if !slices.Contains(allowedRoles, auth.UserRole(claims.Role)) {
				m.handleAuthErr(w, r, xerr.NewErr(op, pAuth.NotAuthorized))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
