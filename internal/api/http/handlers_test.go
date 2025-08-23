package http

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/shrtyk/pvz-service/internal/api/http/dto"
	"github.com/shrtyk/pvz-service/internal/core/domain"
	"github.com/shrtyk/pvz-service/internal/core/domain/auth"
	pAuthMock "github.com/shrtyk/pvz-service/internal/core/ports/auth/mocks"
	pServiceMock "github.com/shrtyk/pvz-service/internal/core/ports/service/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type handlerWithMocks struct {
	appService *pServiceMock.MockService
	tService   *pAuthMock.MockTokenService
}

func setup(t *testing.T) (*handlers, *handlerWithMocks) {
	t.Helper()
	fields := &handlerWithMocks{
		appService: pServiceMock.NewMockService(t),
		tService:   pAuthMock.NewMockTokenService(t),
	}
	h := NewHandlers(fields.appService, fields.tService)
	return h, fields
}

// failingWriter is a mock http.ResponseWriter that fails on Write().
type failingWriter struct {
	headers    http.Header
	statusCode int
}

func (fw *failingWriter) Header() http.Header {
	if fw.headers == nil {
		fw.headers = make(http.Header)
	}
	return fw.headers
}

func (fw *failingWriter) Write(b []byte) (int, error) {
	return 0, errors.New("write failed")
}

func (fw *failingWriter) WriteHeader(statusCode int) {
	fw.statusCode = statusCode
}

func TestHandlers_DummyLoginHandler(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		body       any
		setup      func(f *handlerWithMocks)
		wantStatus int
		writer     http.ResponseWriter
	}{
		{
			name: "success",
			body: dto.PostDummyLoginJSONRequestBody{Role: "employee"},
			setup: func(f *handlerWithMocks) {
				f.tService.On("GenerateAccessToken", mock.Anything).
					Return("token", nil).Once()
			},
			wantStatus: http.StatusOK,
			writer:     httptest.NewRecorder(),
		},
		{
			name:       "bad request",
			body:       "not json",
			setup:      func(f *handlerWithMocks) {},
			wantStatus: http.StatusBadRequest,
			writer:     httptest.NewRecorder(),
		},
		{
			name:       "validation error",
			body:       dto.PostDummyLoginJSONRequestBody{Role: "invalid"},
			setup:      func(f *handlerWithMocks) {},
			wantStatus: http.StatusBadRequest,
			writer:     httptest.NewRecorder(),
		},
		{
			name: "token service error",
			body: dto.PostDummyLoginJSONRequestBody{Role: "employee"},
			setup: func(f *handlerWithMocks) {
				f.tService.On("GenerateAccessToken", mock.Anything).
					Return("", assert.AnError).Once()
			},
			wantStatus: http.StatusInternalServerError,
			writer:     httptest.NewRecorder(),
		},
		{
			name: "writejson error",
			body: dto.PostDummyLoginJSONRequestBody{Role: "employee"},
			setup: func(f *handlerWithMocks) {
				f.tService.On("GenerateAccessToken", mock.Anything).
					Return("token", nil).Once()
			},
			wantStatus: http.StatusInternalServerError,
			writer:     &failingWriter{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			h, f := setup(t)
			tt.setup(f)

			bodyBytes, _ := json.Marshal(tt.body)
			req := httptest.NewRequest(http.MethodPost, "/dummy-login", bytes.NewReader(bodyBytes))

			err := h.DummyLoginHandler(tt.writer, req)

			if err != nil {
				var httpErr *HTTPError
				if errors.As(err, &httpErr) {
					assert.Equal(t, tt.wantStatus, httpErr.Code)
				}
			} else if rr, ok := tt.writer.(*httptest.ResponseRecorder); ok {
				assert.Equal(t, tt.wantStatus, rr.Code)
			}
		})
	}
}

func TestHandlers_NewPVZHandler(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		body       any
		setup      func(f *handlerWithMocks)
		wantStatus int
	}{
		{
			name: "success",
			body: dto.PVZ{City: "Москва"},
			setup: func(f *handlerWithMocks) {
				f.appService.On("NewPVZ", mock.Anything, mock.Anything).
					Return(&domain.Pvz{}, nil).Once()
			},
			wantStatus: http.StatusCreated,
		},
		{
			name:       "validation error",
			body:       dto.PVZ{City: ""},
			setup:      func(f *handlerWithMocks) {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "service error",
			body: dto.PVZ{City: "Москва"},
			setup: func(f *handlerWithMocks) {
				f.appService.On("NewPVZ", mock.Anything, mock.Anything).
					Return(nil, assert.AnError).Once()
			},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:       "wrong body error",
			body:       "",
			setup:      func(f *handlerWithMocks) {},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			h, f := setup(t)
			tt.setup(f)

			bodyBytes, _ := json.Marshal(tt.body)
			req := httptest.NewRequest(http.MethodPost, "/pvz", bytes.NewReader(bodyBytes))
			rr := httptest.NewRecorder()

			err := h.NewPVZHandler(rr, req)

			if err != nil {
				var httpErr *HTTPError
				if errors.As(err, &httpErr) {
					assert.Equal(t, tt.wantStatus, httpErr.Code)
				}
			} else {
				assert.Equal(t, tt.wantStatus, rr.Code)
			}
		})
	}
}

func TestHandlers_NewReceptionHandler(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		body       any
		setup      func(f *handlerWithMocks)
		wantStatus int
	}{
		{
			name: "success",
			body: dto.PostReceptionsJSONBody{PvzId: uuid.New()},
			setup: func(f *handlerWithMocks) {
				f.appService.On("OpenNewPVZReception", mock.Anything, mock.Anything).
					Return(&domain.Reception{}, nil).Once()
			},
			wantStatus: http.StatusCreated,
		},
		{
			name:       "validation error",
			body:       dto.PostReceptionsJSONBody{},
			setup:      func(f *handlerWithMocks) {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "service error",
			body: dto.PostReceptionsJSONBody{PvzId: uuid.New()},
			setup: func(f *handlerWithMocks) {
				f.appService.On("OpenNewPVZReception", mock.Anything, mock.Anything).
					Return(nil, assert.AnError).Once()
			},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:       "wrong body error",
			body:       "",
			setup:      func(f *handlerWithMocks) {},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			h, f := setup(t)
			tt.setup(f)

			bodyBytes, _ := json.Marshal(tt.body)
			req := httptest.NewRequest(http.MethodPost, "/receptions", bytes.NewReader(bodyBytes))
			rr := httptest.NewRecorder()

			err := h.NewReceptionHandler(rr, req)

			if err != nil {
				var httpErr *HTTPError
				if errors.As(err, &httpErr) {
					assert.Equal(t, tt.wantStatus, httpErr.Code)
				}
			} else {
				assert.Equal(t, tt.wantStatus, rr.Code)
			}
		})
	}
}

func TestHandlers_AddProductHandler(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		body       any
		setup      func(f *handlerWithMocks)
		wantStatus int
	}{
		{
			name: "success",
			body: dto.PostProductsJSONRequestBody{PvzId: uuid.New(), Type: "одежда"},
			setup: func(f *handlerWithMocks) {
				f.appService.On("AddProductPVZ", mock.Anything, mock.Anything).
					Return(&domain.Product{}, nil).Once()
			},
			wantStatus: http.StatusCreated,
		},
		{
			name:       "validation error",
			body:       dto.PostProductsJSONRequestBody{PvzId: uuid.New(), Type: ""},
			setup:      func(f *handlerWithMocks) {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "service error",
			body: dto.PostProductsJSONRequestBody{PvzId: uuid.New(), Type: "одежда"},
			setup: func(f *handlerWithMocks) {
				f.appService.On("AddProductPVZ", mock.Anything, mock.Anything).
					Return(nil, assert.AnError).Once()
			},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:       "wrong body error",
			body:       "",
			setup:      func(f *handlerWithMocks) {},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			h, f := setup(t)
			tt.setup(f)

			bodyBytes, _ := json.Marshal(tt.body)
			req := httptest.NewRequest(http.MethodPost, "/products", bytes.NewReader(bodyBytes))
			rr := httptest.NewRecorder()

			err := h.AddProductHandler(rr, req)

			if err != nil {
				var httpErr *HTTPError
				if errors.As(err, &httpErr) {
					assert.Equal(t, tt.wantStatus, httpErr.Code)
				}
			} else {
				assert.Equal(t, tt.wantStatus, rr.Code)
			}
		})
	}
}

func TestHandlers_DeleteLastProductHandler(t *testing.T) {
	t.Parallel()

	pvzID := uuid.New()

	tests := []struct {
		name       string
		pvzID      string
		setup      func(f *handlerWithMocks)
		wantStatus int
	}{
		{
			name:  "success",
			pvzID: pvzID.String(),
			setup: func(f *handlerWithMocks) {
				f.appService.On("DeleteLastProductPvz", mock.Anything, &pvzID).
					Return(nil).Once()
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "invalid pvzId",
			pvzID:      "invalid-uuid",
			setup:      func(f *handlerWithMocks) {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:  "service error",
			pvzID: pvzID.String(),
			setup: func(f *handlerWithMocks) {
				f.appService.On("DeleteLastProductPvz", mock.Anything, &pvzID).
					Return(assert.AnError).Once()
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			h, f := setup(t)
			tt.setup(f)

			req := httptest.NewRequest(http.MethodDelete, "/product/"+tt.pvzID, nil)
			chiCtx := chi.NewRouteContext()
			chiCtx.URLParams.Add("pvzId", tt.pvzID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiCtx))
			rr := httptest.NewRecorder()

			err := h.DeleteLastProductHandler(rr, req)

			if err != nil {
				var httpErr *HTTPError
				if errors.As(err, &httpErr) {
					assert.Equal(t, tt.wantStatus, httpErr.Code)
				}
			} else {
				assert.Equal(t, tt.wantStatus, rr.Code)
			}
		})
	}
}

func TestHandlers_CloseReceptionHandler(t *testing.T) {
	t.Parallel()

	pvzID := uuid.New()

	tests := []struct {
		name       string
		pvzID      string
		setup      func(f *handlerWithMocks)
		wantStatus int
	}{
		{
			name:  "success",
			pvzID: pvzID.String(),
			setup: func(f *handlerWithMocks) {
				f.appService.On("CloseReceptionInPvz", mock.Anything, &pvzID).
					Return(nil).Once()
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "invalid pvzId",
			pvzID:      "invalid-uuid",
			setup:      func(f *handlerWithMocks) {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:  "service error",
			pvzID: pvzID.String(),
			setup: func(f *handlerWithMocks) {
				f.appService.On("CloseReceptionInPvz", mock.Anything, &pvzID).
					Return(assert.AnError).Once()
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			h, f := setup(t)
			tt.setup(f)

			req := httptest.NewRequest(http.MethodPost, "/receptions/"+tt.pvzID+"/close", nil)
			chiCtx := chi.NewRouteContext()
			chiCtx.URLParams.Add("pvzId", tt.pvzID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiCtx))
			rr := httptest.NewRecorder()

			err := h.CloseReceptionHandler(rr, req)

			if err != nil {
				var httpErr *HTTPError
				if errors.As(err, &httpErr) {
					assert.Equal(t, tt.wantStatus, httpErr.Code)
				}
			} else {
				assert.Equal(t, tt.wantStatus, rr.Code)
			}
		})
	}
}

func TestHandlers_GetPvzHandler(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		url        string
		setup      func(f *handlerWithMocks)
		wantStatus int
	}{
		{
			name: "success",
			url:  "/pvz",
			setup: func(f *handlerWithMocks) {
				f.appService.On("GetPvzsData", mock.Anything, mock.Anything).
					Return([]*domain.PvzReceptions{}, nil).Once()
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "bad query params",
			url:        "/pvz?page=abc",
			setup:      func(f *handlerWithMocks) {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "service error",
			url:  "/pvz",
			setup: func(f *handlerWithMocks) {
				f.appService.On("GetPvzsData", mock.Anything, mock.Anything).
					Return(nil, assert.AnError).Once()
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			h, f := setup(t)
			tt.setup(f)

			req := httptest.NewRequest(http.MethodGet, tt.url, nil)
			rr := httptest.NewRecorder()

			err := h.GetPvzHandler(rr, req)

			if err != nil {
				var httpErr *HTTPError
				if errors.As(err, &httpErr) {
					assert.Equal(t, tt.wantStatus, httpErr.Code)
				}
			} else {
				assert.Equal(t, tt.wantStatus, rr.Code)
			}
		})
	}
}

func TestHandlers_HealthZ(t *testing.T) {
	t.Parallel()

	h, _ := setup(t)
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rr := httptest.NewRecorder()

	err := h.HealthZ(rr, req)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestHandlers_RegisterUserHandler(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		body       any
		setup      func(f *handlerWithMocks)
		wantStatus int
		writer     http.ResponseWriter
	}{
		{
			name: "success",
			body: dto.PostRegisterJSONRequestBody{Email: "test@test.com", Password: "password", Role: "employee"},
			setup: func(f *handlerWithMocks) {
				f.appService.On("RegisterUser", mock.Anything, mock.Anything).
					Return(&auth.User{
						Id:    uuid.New(),
						Role:  "moderator",
						Email: "a@a.com",
					}, nil).Once()
			},
			wantStatus: http.StatusCreated,
			writer:     httptest.NewRecorder(),
		},
		{
			name:       "bad request",
			body:       "not json",
			setup:      func(f *handlerWithMocks) {},
			wantStatus: http.StatusBadRequest,
			writer:     httptest.NewRecorder(),
		},
		{
			name:       "validation error",
			body:       dto.PostRegisterJSONRequestBody{Email: "not-an-email", Password: "password", Role: "employee"},
			setup:      func(f *handlerWithMocks) {},
			wantStatus: http.StatusBadRequest,
			writer:     httptest.NewRecorder(),
		},
		{
			name: "service error",
			body: dto.PostRegisterJSONRequestBody{Email: "test@test.com", Password: "password", Role: "employee"},
			setup: func(f *handlerWithMocks) {
				f.appService.On("RegisterUser", mock.Anything, mock.Anything).
					Return(nil, assert.AnError).Once()
			},
			wantStatus: http.StatusInternalServerError,
			writer:     httptest.NewRecorder(),
		},
		{
			name: "write json error",
			body: dto.PostRegisterJSONRequestBody{Email: "test@test.com", Password: "password", Role: "employee"},
			setup: func(f *handlerWithMocks) {
				f.appService.On("RegisterUser", mock.Anything, mock.Anything).
					Return(&auth.User{}, nil).Once()
			},
			wantStatus: http.StatusInternalServerError,
			writer:     &failingWriter{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			h, f := setup(t)
			tt.setup(f)

			bodyBytes, _ := json.Marshal(tt.body)
			req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader(bodyBytes))

			err := h.RegisterUserHandler(tt.writer, req)

			if err != nil {
				var httpErr *HTTPError
				if errors.As(err, &httpErr) {
					assert.Equal(t, tt.wantStatus, httpErr.Code)
				}
			} else if rr, ok := tt.writer.(*httptest.ResponseRecorder); ok {
				assert.Equal(t, tt.wantStatus, rr.Code)
			}
		})
	}
}

func TestHandlers_LoginUserHandler(t *testing.T) {
	t.Parallel()

	rToken := &auth.RefreshToken{
		Token:     "refresh-token",
		ExpiresAt: time.Now().Add(time.Hour),
	}

	tests := []struct {
		name       string
		body       any
		setup      func(f *handlerWithMocks)
		wantStatus int
		writer     http.ResponseWriter
		check      func(t *testing.T, w http.ResponseWriter)
	}{
		{
			name: "success",
			body: dto.PostLoginJSONRequestBody{Email: "test@test.com", Password: "password"},
			setup: func(f *handlerWithMocks) {
				f.appService.On("LoginUser", mock.Anything, mock.Anything).
					Return("access-token", rToken, nil).Once()
			},
			wantStatus: http.StatusOK,
			writer:     httptest.NewRecorder(),
			check: func(t *testing.T, w http.ResponseWriter) {
				rr := w.(*httptest.ResponseRecorder)
				cookies := rr.Result().Cookies()
				assert.Len(t, cookies, 1)
				assert.Equal(t, refreshTokenKey, cookies[0].Name)
				assert.Equal(t, "refresh-token", cookies[0].Value)
			},
		},
		{
			name:       "bad request",
			body:       "not json",
			setup:      func(f *handlerWithMocks) {},
			wantStatus: http.StatusBadRequest,
			writer:     httptest.NewRecorder(),
			check:      func(t *testing.T, w http.ResponseWriter) {},
		},
		{
			name:       "validation error",
			body:       dto.PostLoginJSONRequestBody{Email: "not-an-email", Password: "password"},
			setup:      func(f *handlerWithMocks) {},
			wantStatus: http.StatusBadRequest,
			writer:     httptest.NewRecorder(),
			check:      func(t *testing.T, w http.ResponseWriter) {},
		},
		{
			name: "service error",
			body: dto.PostLoginJSONRequestBody{Email: "test@test.com", Password: "password"},
			setup: func(f *handlerWithMocks) {
				f.appService.On("LoginUser", mock.Anything, mock.Anything).
					Return("", nil, assert.AnError).Once()
			},
			wantStatus: http.StatusInternalServerError,
			writer:     httptest.NewRecorder(),
			check:      func(t *testing.T, w http.ResponseWriter) {},
		},
		{
			name: "writejson error",
			body: dto.PostLoginJSONRequestBody{Email: "test@test.com", Password: "password"},
			setup: func(f *handlerWithMocks) {
				f.appService.On("LoginUser", mock.Anything, mock.Anything).
					Return("access-token", rToken, nil).Once()
			},
			wantStatus: http.StatusInternalServerError,
			writer:     &failingWriter{},
			check:      func(t *testing.T, w http.ResponseWriter) {},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			h, f := setup(t)
			tt.setup(f)

			bodyBytes, _ := json.Marshal(tt.body)
			req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(bodyBytes))

			err := h.LoginUserHandler(tt.writer, req)

			if err != nil {
				var httpErr *HTTPError
				if errors.As(err, &httpErr) {
					assert.Equal(t, tt.wantStatus, httpErr.Code)
				}
			} else if rr, ok := tt.writer.(*httptest.ResponseRecorder); ok {
				assert.Equal(t, tt.wantStatus, rr.Code)
			}
			tt.check(t, tt.writer)
		})
	}
}

func TestHandlers_RefreshTokensHandler(t *testing.T) {
	t.Parallel()

	newRToken := &auth.RefreshToken{
		Token:     "new-refresh-token",
		ExpiresAt: time.Now().Add(time.Hour),
	}

	tests := []struct {
		name       string
		cookie     *http.Cookie
		setup      func(f *handlerWithMocks)
		wantStatus int
		writer     http.ResponseWriter
		check      func(t *testing.T, w http.ResponseWriter)
	}{
		{
			name:   "success",
			cookie: &http.Cookie{Name: refreshTokenKey, Value: "old-refresh-token"},
			setup: func(f *handlerWithMocks) {
				f.appService.On("RefreshTokens", mock.Anything, mock.Anything).
					Return("new-access-token", newRToken, nil).Once()
			},
			wantStatus: http.StatusCreated,
			writer:     httptest.NewRecorder(),
			check: func(t *testing.T, w http.ResponseWriter) {
				rr := w.(*httptest.ResponseRecorder)
				cookies := rr.Result().Cookies()
				assert.Len(t, cookies, 1)
				assert.Equal(t, refreshTokenKey, cookies[0].Name)
				assert.Equal(t, "new-refresh-token", cookies[0].Value)
			},
		},
		{
			name:       "no cookie",
			cookie:     nil,
			setup:      func(f *handlerWithMocks) {},
			wantStatus: http.StatusUnauthorized,
			writer:     httptest.NewRecorder(),
			check:      func(t *testing.T, w http.ResponseWriter) {},
		},
		{
			name:   "service error",
			cookie: &http.Cookie{Name: refreshTokenKey, Value: "old-refresh-token"},
			setup: func(f *handlerWithMocks) {
				f.appService.On("RefreshTokens", mock.Anything, mock.Anything).
					Return("", nil, assert.AnError).Once()
			},
			wantStatus: http.StatusInternalServerError,
			writer:     httptest.NewRecorder(),
			check:      func(t *testing.T, w http.ResponseWriter) {},
		},
		{
			name:   "writejson error",
			cookie: &http.Cookie{Name: refreshTokenKey, Value: "old-refresh-token"},
			setup: func(f *handlerWithMocks) {
				f.appService.On("RefreshTokens", mock.Anything, mock.Anything).
					Return("new-access-token", newRToken, nil).Once()
			},
			wantStatus: http.StatusInternalServerError,
			writer:     &failingWriter{},
			check:      func(t *testing.T, w http.ResponseWriter) {},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			h, f := setup(t)
			tt.setup(f)

			req := httptest.NewRequest(http.MethodPost, "/tokens/refresh", nil)
			if tt.cookie != nil {
				req.AddCookie(tt.cookie)
			}

			err := h.RefreshTokensHandler(tt.writer, req)

			if err != nil {
				var httpErr *HTTPError
				if errors.As(err, &httpErr) {
					assert.Equal(t, tt.wantStatus, httpErr.Code)
				}
			} else if rr, ok := tt.writer.(*httptest.ResponseRecorder); ok {
				assert.Equal(t, tt.wantStatus, rr.Code)
			}
			tt.check(t, tt.writer)
		})
	}
}
