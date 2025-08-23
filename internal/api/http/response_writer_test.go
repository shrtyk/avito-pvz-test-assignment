package http

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/shrtyk/pvz-service/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockResponseWriter struct {
	header     http.Header
	statusCode int
	writeErr   error
}

func (m *mockResponseWriter) Header() http.Header {
	if m.header == nil {
		m.header = make(http.Header)
	}
	return m.header
}

func (m *mockResponseWriter) Write(b []byte) (int, error) {
	return 0, m.writeErr
}

func (m *mockResponseWriter) WriteHeader(statusCode int) {
	m.statusCode = statusCode
}

func TestWriteHTTPError(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name               string
		err                *HTTPError
		expectedStatusCode int
		expectedBody       string
		expectedLog        string
		w                  http.ResponseWriter
	}{
		{
			name: "server error",
			err: &HTTPError{
				Code:    http.StatusInternalServerError,
				Message: "Internal error",
				Err:     errors.New("test error"),
			},
			expectedStatusCode: http.StatusInternalServerError,
			expectedBody:       `{"message":"Internal error"}`,
			expectedLog:        "Server error",
			w:                  httptest.NewRecorder(),
		},
		{
			name: "client error",
			err: &HTTPError{
				Code:    http.StatusBadRequest,
				Message: "Bad request",
				Err:     errors.New("test error"),
			},
			expectedStatusCode: http.StatusBadRequest,
			expectedBody:       `{"message":"Bad request"}`,
			expectedLog:        "Client error",
			w:                  httptest.NewRecorder(),
		},
		{
			name: "write json error",
			err: &HTTPError{
				Code:    http.StatusInternalServerError,
				Message: "Internal error",
				Err:     errors.New("test error"),
			},
			expectedStatusCode: http.StatusInternalServerError,
			expectedBody:       "",
			expectedLog:        "Failed to response with error",
			w:                  &mockResponseWriter{writeErr: errors.New("write error")},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			r := httptest.NewRequest(http.MethodGet, "/", nil)
			l, logs := logger.NewTestLogger()
			r = r.WithContext(logger.ToCtx(context.Background(), l))

			WriteHTTPError(tc.w, r, tc.err)

			if recorder, ok := tc.w.(*httptest.ResponseRecorder); ok {
				assert.Equal(t, tc.expectedStatusCode, recorder.Code)
				if tc.expectedBody != "" {
					assert.JSONEq(t, tc.expectedBody, recorder.Body.String())
				}
			}

			if mrw, ok := tc.w.(*mockResponseWriter); ok {
				assert.Equal(t, tc.expectedStatusCode, mrw.statusCode)
			}

			assert.Contains(t, logs.String(), tc.expectedLog)
		})
	}
}

func TestWriteJSON(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		data           any
		status         int
		headers        http.Header
		expectedErr    bool
		assertResponse func(t *testing.T, w *httptest.ResponseRecorder)
	}{
		{
			name:   "success",
			data:   map[string]string{"foo": "bar"},
			status: http.StatusOK,
			headers: http.Header{
				"X-Test": []string{"true"},
			},
			expectedErr: false,
			assertResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusOK, w.Code)
				assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
				assert.Equal(t, "true", w.Header().Get("X-Test"))
				assert.JSONEq(t, `{"foo":"bar"}`, w.Body.String())
			},
		},
		{
			name:           "marshal error",
			data:           map[string]any{"foo": func() {}},
			status:         http.StatusOK,
			headers:        nil,
			expectedErr:    true,
			assertResponse: func(t *testing.T, w *httptest.ResponseRecorder) {},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			w := httptest.NewRecorder()

			err := WriteJSON(w, tc.data, tc.status, tc.headers)

			if tc.expectedErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			tc.assertResponse(t, w)
		})
	}
}
