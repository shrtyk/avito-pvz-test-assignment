package http

import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHTTPError(t *testing.T) {
	t.Parallel()

	t.Run("Error method", func(t *testing.T) {
		t.Parallel()
		err := &HTTPError{
			Code:    http.StatusInternalServerError,
			Message: "Internal error",
			Err:     errors.New("test error"),
		}
		want := fmt.Sprintf("HTTP Code: %d. Message: %s. Error: %s", http.StatusInternalServerError, "Internal error", "test error")
		assert.Equal(t, want, err.Error())
	})

	t.Run("Unwrap method", func(t *testing.T) {
		t.Parallel()
		originalErr := errors.New("test error")
		err := &HTTPError{
			Err: originalErr,
		}
		assert.Equal(t, originalErr, err.Unwrap())
	})
}

func TestErrorConstructors(t *testing.T) {
	t.Parallel()

	originalErr := errors.New("test error")

	testCases := []struct {
		name          string
		constructor   func(err error) *HTTPError
		expectedCode  int
		expectedMsg   string
		expectedError error
	}{
		{
			name:          "BadRequestBodyError",
			constructor:   BadRequestBodyError,
			expectedCode:  http.StatusBadRequest,
			expectedMsg:   "Badly formed request body",
			expectedError: originalErr,
		},
		{
			name:          "BadRequestQueryParamsError",
			constructor:   BadRequestQueryParamsError,
			expectedCode:  http.StatusBadRequest,
			expectedMsg:   "Wrong formated url query params",
			expectedError: originalErr,
		},
		{
			name:          "InternalError",
			constructor:   InternalError,
			expectedCode:  http.StatusInternalServerError,
			expectedMsg:   "Internal error",
			expectedError: originalErr,
		},
		{
			name:          "ValidationError",
			constructor:   ValidationError,
			expectedCode:  http.StatusBadRequest,
			expectedMsg:   "Invalid data in request body",
			expectedError: originalErr,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			err := tc.constructor(originalErr)
			assert.Equal(t, tc.expectedCode, err.Code)
			assert.Equal(t, tc.expectedMsg, err.Message)
			assert.Equal(t, tc.expectedError, err.Err)
		})
	}
}
