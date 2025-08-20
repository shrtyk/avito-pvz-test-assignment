package http

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/shrtyk/avito-pvz-test-assignment/pkg/logger"
	"github.com/stretchr/testify/assert"
)

func TestHandle(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		h := func(w http.ResponseWriter, r *http.Request) error {
			w.WriteHeader(http.StatusOK)
			return nil
		}

		server := httptest.NewServer(Handle(h))
		defer server.Close()

		resp, err := http.Get(server.URL)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("http error", func(t *testing.T) {
		t.Parallel()
		h := Handle(func(w http.ResponseWriter, r *http.Request) error {
			return &HTTPError{
				Code:    http.StatusBadRequest,
				Message: "Bad request",
				Err:     errors.New("test error"),
			}
		})

		server := httptest.NewServer(h)
		defer server.Close()

		resp, err := http.Get(server.URL)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("default error", func(t *testing.T) {
		t.Parallel()
		h := Handle(func(w http.ResponseWriter, r *http.Request) error {
			return errors.New("test error")
		})

		l, _ := logger.NewTestLogger()
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		r = r.WithContext(logger.ToCtx(context.Background(), l))

		w := httptest.NewRecorder()
		h(w, r)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}
