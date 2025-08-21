package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReadJson(t *testing.T) {
	t.Parallel()

	type request struct {
		Foo string `json:"foo"`
	}

	tests := []struct {
		name     string
		body     string
		dst      any
		expErr   string
		checkErr func(t *testing.T, err error, expErr string)
	}{
		{
			name:     "success",
			body:     `{"foo":"bar"}`,
			dst:      &request{},
			checkErr: func(t *testing.T, err error, expErr string) { require.NoError(t, err) },
		},
		{
			name:     "invalid json",
			body:     `{"foo":"bar"`,
			dst:      &request{},
			checkErr: func(t *testing.T, err error, expErr string) { assert.Error(t, err) },
		},
		{
			name:     "empty body",
			body:     "",
			dst:      &request{},
			expErr:   "body must not be empty",
			checkErr: func(t *testing.T, err error, expErr string) { assert.EqualError(t, err, expErr) },
		},
		{
			name:     "unknown field",
			body:     `{"bar":"baz"}`,
			dst:      &request{},
			expErr:   "body contains unknown key",
			checkErr: func(t *testing.T, err error, expErr string) { assert.Contains(t, err.Error(), expErr) },
		},
		{
			name:     "multiple json values",
			body:     `{"foo":"bar"}{"bar":"baz"}`,
			dst:      &request{},
			expErr:   "body must contain a single JSON value",
			checkErr: func(t *testing.T, err error, expErr string) { assert.EqualError(t, err, expErr) },
		},
		{
			name:     "syntax error",
			body:     `{"foo":"bar"`,
			dst:      &request{},
			expErr:   "body contains badly-formed JSON",
			checkErr: func(t *testing.T, err error, expErr string) { assert.Contains(t, err.Error(), expErr) },
		},
		{
			name:     "unmarshal type error",
			body:     `{"foo":123}`,
			dst:      &request{},
			expErr:   "body contains incorrect JSON type for field",
			checkErr: func(t *testing.T, err error, expErr string) { assert.Contains(t, err.Error(), expErr) },
		},
		{
			name:     "max bytes error",
			body:     `{"foo":"` + strings.Repeat("a", 1_048_577) + `"}`,
			dst:      &request{},
			expErr:   "body must not be larger than",
			checkErr: func(t *testing.T, err error, expErr string) { assert.Contains(t, err.Error(), expErr) },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(tt.body))

			err := ReadJson(w, r, tt.dst)

			tt.checkErr(t, err, tt.expErr)
		})
	}
}

func TestIdParam(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		r := httptest.NewRequest(http.MethodGet, "/123", nil)
		chiCtx := chi.NewRouteContext()
		chiCtx.URLParams.Add("id", "123")
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, chiCtx))

		id, err := IdParam(r)

		require.NoError(t, err)
		assert.Equal(t, int64(123), id)
	})

	t.Run("invalid id", func(t *testing.T) {
		t.Parallel()
		r := httptest.NewRequest(http.MethodGet, "/abc", nil)
		chiCtx := chi.NewRouteContext()
		chiCtx.URLParams.Add("id", "abc")
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, chiCtx))

		_, err := IdParam(r)

		assert.Error(t, err)
	})
}

func TestPvzIdParam(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		pvzID := uuid.New()
		r := httptest.NewRequest(http.MethodGet, "/"+pvzID.String(), nil)
		chiCtx := chi.NewRouteContext()
		chiCtx.URLParams.Add("pvzId", pvzID.String())
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, chiCtx))

		id, err := PvzIdParam(r)

		require.NoError(t, err)
		assert.Equal(t, &pvzID, id)
	})

	t.Run("invalid id", func(t *testing.T) {
		t.Parallel()
		r := httptest.NewRequest(http.MethodGet, "/abc", nil)
		chiCtx := chi.NewRouteContext()
		chiCtx.URLParams.Add("pvzId", "abc")
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, chiCtx))

		_, err := PvzIdParam(r)

		assert.Error(t, err)
	})
}

func TestBearerToken(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		header        string
		expectedToken string
		expectErr     bool
	}{
		{
			name:          "success",
			header:        "Bearer my-token",
			expectedToken: "my-token",
			expectErr:     false,
		},
		{
			name:      "no header",
			header:    "",
			expectErr: true,
		},
		{
			name:      "invalid header",
			header:    "my-token",
			expectErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			r := httptest.NewRequest(http.MethodGet, "/", nil)
			if tc.header != "" {
				r.Header.Set("Authorization", tc.header)
			}

			token, err := BearerToken(r)

			if tc.expectErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.expectedToken, token)
			}
		})
	}
}

func TestUserAgentAndIP(t *testing.T) {
	t.Parallel()

	eua, eip := "test-agent", "1.2.3.4"
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("User-Agent", eua)
	r.Header.Set("X-Real-IP", eip)

	ua, ip := UserAgentAndIP(r)

	assert.Equal(t, eua, ua)
	assert.Equal(t, eip, ip)
}

func TestPvzParamsFromURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		url      string
		checkErr func(t *testing.T, err error)
	}{
		{
			name:     "success",
			url:      "/?startDate=2025-01-01T00:00:00Z&endDate=2025-01-02T00:00:00Z&page=2&limit=20",
			checkErr: func(t *testing.T, err error) { require.NoError(t, err) },
		},
		{
			name:     "invalid start date",
			url:      "/?startDate=invalid-date",
			checkErr: func(t *testing.T, err error) { assert.Error(t, err) },
		},
		{
			name:     "invalid end date",
			url:      "/?endDate=invalid-date",
			checkErr: func(t *testing.T, err error) { assert.Error(t, err) },
		},
		{
			name:     "invalid page",
			url:      "/?page=abc",
			checkErr: func(t *testing.T, err error) { assert.Error(t, err) },
		},
		{
			name:     "invalid limit",
			url:      "/?limit=abc",
			checkErr: func(t *testing.T, err error) { assert.Error(t, err) },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			r := httptest.NewRequest(http.MethodGet, tt.url, nil)

			_, err := PvzParamsFromURL(r)

			tt.checkErr(t, err)
		})
	}
}
