
package http

import (
	"errors"
	"net/http"
	"testing"

	"github.com/shrtyk/avito-pvz-test-assignment/internal/core/ports/auth"
	ps "github.com/shrtyk/avito-pvz-test-assignment/internal/core/ports/service"
	xerr "github.com/shrtyk/avito-pvz-test-assignment/pkg/xerrors"
	"github.com/stretchr/testify/assert"
)

func Test_mapAppServiceErrsToHTTP(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		err        error
		wantStatus int
	}{
		{
			name:       "unexpected error",
			err:        xerr.NewErr("op", ps.Unexpected),
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:       "active reception exists",
			err:        xerr.NewErr("op", ps.ActiveReceptionExists),
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "default error",
			err:        errors.New("some error"),
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			httpErr := mapAppServiceErrsToHTTP(tt.err)
			assert.Equal(t, tt.wantStatus, httpErr.Code)
		})
	}
}

func Test_mapAuthServiceErrsToHTTP(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		err        error
		wantStatus int
	}{
		{
			name:       "jwt creation error",
			err:        xerr.NewErr("op", auth.JwtCreation),
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:       "invalid jwt",
			err:        xerr.NewErr("op", auth.InvalidJwt),
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "not authorized",
			err:        xerr.NewErr("op", auth.NotAuthorized),
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "default error",
			err:        errors.New("some error"),
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			httpErr := mapAuthServiceErrsToHTTP(tt.err)
			assert.Equal(t, tt.wantStatus, httpErr.Code)
		})
	}
}
