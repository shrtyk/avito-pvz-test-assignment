package http

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/shrtyk/avito-pvz-test-assignment/internal/core/domain/auth"
	pAuthMock "github.com/shrtyk/avito-pvz-test-assignment/internal/core/ports/auth/mocks"
	metricsmocks "github.com/shrtyk/avito-pvz-test-assignment/internal/core/ports/metrics/mocks"
	tservice "github.com/shrtyk/avito-pvz-test-assignment/internal/infrastructure/tservice"
	"github.com/shrtyk/avito-pvz-test-assignment/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestMiddlewares_PanicRecoveryMW(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name                 string
		handler              http.HandlerFunc
		expectedCode         int
		expectedBodyContains string
		expectedLogContains  string
	}{
		{
			name: "panic recovery",
			handler: func(w http.ResponseWriter, r *http.Request) {
				panic("test panic")
			},
			expectedCode:         http.StatusInternalServerError,
			expectedBodyContains: "The server encountered a problem and could not process your request",
			expectedLogContains:  "test panic",
		},
		{
			name: "no panic",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			},
			expectedCode: http.StatusOK,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			l, logs := logger.NewTestLogger()
			metrics := new(metricsmocks.MockCollector)
			m := NewMiddlewares(nil, l, metrics)

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			rr := httptest.NewRecorder()

			m.PanicRecoveryMW(tc.handler).ServeHTTP(rr, req)

			assert.Equal(t, tc.expectedCode, rr.Code)
			if tc.expectedBodyContains != "" {
				assert.Contains(t, rr.Body.String(), tc.expectedBodyContains)
			}
			if tc.expectedLogContains != "" {
				assert.Contains(t, logs.String(), tc.expectedLogContains)
			}
		})
	}
}

func TestMiddlewares_LoggingMW(t *testing.T) {
	t.Parallel()

	l, logs := logger.NewTestLogger()
	metrics := new(metricsmocks.MockCollector)
	m := NewMiddlewares(nil, l, metrics)

	metrics.On("ObserveHTTPRequestDuration", http.MethodGet, mock.AnythingOfType("float64")).Return()
	metrics.On("IncHTTPRequestsTotal", http.MethodGet, "202").Return()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctxL := logger.FromCtx(r.Context())
		assert.NotNil(t, ctxL)
		w.WriteHeader(http.StatusAccepted)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("User-Agent", "test-agent")
	rr := httptest.NewRecorder()

	m.LoggingMW(handler).ServeHTTP(rr, req)

	assert.Equal(t, http.StatusAccepted, rr.Code)
	logStr := logs.String()
	assert.Contains(
		t,
		logStr,
		"New HTTP request", "HTTP request processed", "ip",
		"user_agent", "test-agent", "request_id", "method",
		"GET", "uri", "/test", "status_code", "request_duration")
}

func TestMiddlewares_AuthenticationMW(t *testing.T) {
	t.Parallel()

	mockClaims := &auth.AccessTokenClaims{
		Role: string(auth.UserRoleEmployee),
		RegisteredClaims: jwt.RegisteredClaims{
			Subject: "1",
		},
	}

	testCases := []struct {
		name            string
		authHeader      string
		setupMock       func(ts *pAuthMock.MockTokenService)
		expectNext      bool
		expectErrHandle bool
	}{
		{
			name:            "no auth header",
			authHeader:      "",
			setupMock:       func(ts *pAuthMock.MockTokenService) {},
			expectNext:      false,
			expectErrHandle: true,
		},
		{
			name:            "invalid auth header",
			authHeader:      "Token",
			setupMock:       func(ts *pAuthMock.MockTokenService) {},
			expectNext:      false,
			expectErrHandle: true,
		},
		{
			name:       "token validation fails",
			authHeader: "Bearer invalid-token",
			setupMock: func(ts *pAuthMock.MockTokenService) {
				ts.On("GetTokenClaims", "invalid-token").Return(nil, errors.New("validation failed"))
			},
			expectNext:      false,
			expectErrHandle: true,
		},
		{
			name:       "success",
			authHeader: "Bearer valid-token",
			setupMock: func(ts *pAuthMock.MockTokenService) {
				ts.On("GetTokenClaims", "valid-token").Return(mockClaims, nil)
			},
			expectNext:      true,
			expectErrHandle: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ts := pAuthMock.NewMockTokenService(t)
			tc.setupMock(ts)

			l, _ := logger.NewTestLogger()
			metrics := new(metricsmocks.MockCollector)

			errHandled := false
			mockErrHandler := func(w http.ResponseWriter, r *http.Request, err error) {
				errHandled = true
			}

			m := &Middlewares{
				tokenService:  ts,
				log:           l,
				metrics:       metrics,
				handleAuthErr: mockErrHandler,
			}

			nextCalled := false
			nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				nextCalled = true
				if tc.name == "success" {
					claims, err := tservice.ClaimsFromCtx(r.Context())
					assert.NoError(t, err)
					assert.Equal(t, mockClaims, claims)
				}
			})

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tc.authHeader != "" {
				req.Header.Set("Authorization", tc.authHeader)
			}
			rr := httptest.NewRecorder()

			m.AuthenticationMW(nextHandler).ServeHTTP(rr, req)

			assert.Equal(t, tc.expectNext, nextCalled)
			assert.Equal(t, tc.expectErrHandle, errHandled)
		})
	}
}

func TestMiddlewares_AuthorizeRoles(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name            string
		allowedRoles    []auth.UserRole
		claimsInCtx     *auth.AccessTokenClaims
		expectNext      bool
		expectErrHandle bool
	}{
		{
			name:            "no claims in context",
			allowedRoles:    []auth.UserRole{auth.UserRoleEmployee},
			claimsInCtx:     nil,
			expectNext:      false,
			expectErrHandle: true,
		},
		{
			name:            "role allowed",
			allowedRoles:    []auth.UserRole{auth.UserRoleEmployee},
			claimsInCtx:     &auth.AccessTokenClaims{Role: string(auth.UserRoleEmployee)},
			expectNext:      true,
			expectErrHandle: false,
		},
		{
			name:            "role not allowed",
			allowedRoles:    []auth.UserRole{auth.UserRoleModerator},
			claimsInCtx:     &auth.AccessTokenClaims{Role: string(auth.UserRoleEmployee)},
			expectNext:      false,
			expectErrHandle: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			l, _ := logger.NewTestLogger()
			metrics := new(metricsmocks.MockCollector)

			errHandled := false
			mockErrHandler := func(w http.ResponseWriter, r *http.Request, err error) {
				errHandled = true
			}

			m := &Middlewares{
				log:           l,
				metrics:       metrics,
				handleAuthErr: mockErrHandler,
			}

			nextCalled := false
			nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				nextCalled = true
			})

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			ctx := req.Context()
			if tc.claimsInCtx != nil {
				ctx = tservice.ClaimsToCtx(ctx, tc.claimsInCtx)
			}
			req = req.WithContext(ctx)

			rr := httptest.NewRecorder()

			m.AuthorizeRoles(tc.allowedRoles...)(nextHandler).ServeHTTP(rr, req)

			assert.Equal(t, tc.expectNext, nextCalled)
			assert.Equal(t, tc.expectErrHandle, errHandled)
		})
	}
}

func TestCustomResponseWriter(t *testing.T) {
	t.Parallel()

	t.Run("WriteHeader once", func(t *testing.T) {
		t.Parallel()
		rr := httptest.NewRecorder()
		crw := &customResponseWriter{ResponseWriter: rr}

		crw.WriteHeader(http.StatusAccepted)
		crw.WriteHeader(http.StatusBadGateway)

		assert.Equal(t, http.StatusAccepted, rr.Code)
	})

	t.Run("Write without WriteHeader", func(t *testing.T) {
		t.Parallel()
		rr := httptest.NewRecorder()
		crw := &customResponseWriter{ResponseWriter: rr}

		_, err := crw.Write([]byte("test"))
		assert.NoError(t, err)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, "test", rr.Body.String())
	})
}
