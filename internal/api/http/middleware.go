package http

import (
	"fmt"
	"log/slog"
	"net/http"
	"slices"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/shrtyk/pvz-service/internal/core/domain/auth"
	pAuth "github.com/shrtyk/pvz-service/internal/core/ports/auth"
	"github.com/shrtyk/pvz-service/internal/core/ports/metrics"
	ts "github.com/shrtyk/pvz-service/internal/infrastructure/tservice"
	"github.com/shrtyk/pvz-service/pkg/logger"
	xerr "github.com/shrtyk/pvz-service/pkg/xerrors"
)

type Middlewares struct {
	tokenService  pAuth.TokenService
	log           *slog.Logger
	metrics       metrics.Collector
	handleAuthErr func(http.ResponseWriter, *http.Request, error)
}

func NewMiddlewares(
	tokenService pAuth.TokenService,
	log *slog.Logger,
	metrics metrics.Collector,
) *Middlewares {
	eh := func(w http.ResponseWriter, r *http.Request, err error) {
		WriteHTTPError(w, r, mapTokenServiceErrsToHTTP(err))
	}

	return &Middlewares{
		tokenService:  tokenService,
		log:           log,
		metrics:       metrics,
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

		m.metrics.ObserveHTTPRequestDuration(r.Method, reqEnd.Seconds())
		m.metrics.IncHTTPRequestsTotal(r.Method, strconv.Itoa(custWriter.statusCode))

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

		claims, err := m.tokenService.GetTokenClaims(bt)
		if err != nil {
			m.handleAuthErr(w, r, err)
			return
		}

		l := logger.FromCtx(r.Context())
		newLog := l.With(slog.String("user_id", claims.UserID()))
		ctxWithLog := logger.ToCtx(r.Context(), newLog)
		ctxWithClaims := ts.ClaimsToCtx(ctxWithLog, claims)

		newReq := r.WithContext(ctxWithClaims)
		next.ServeHTTP(w, newReq)
	})
}

func (m *Middlewares) AuthorizeRoles(allowedRoles ...auth.UserRole) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			const op = "middlewares.AuthorizeRoles"

			claims, err := ts.ClaimsFromCtx(r.Context())
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
