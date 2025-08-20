package http

import (
	"testing"

	"github.com/google/uuid"
	"github.com/oapi-codegen/runtime/types"
	"github.com/stretchr/testify/assert"
)

func TestNewValidator(t *testing.T) {
	t.Parallel()

	validator := NewValidator()

	type testStruct struct {
		ID types.UUID `validate:"oapi_uuid"`
	}

	type testStructWrongType struct {
		ID string `validate:"oapi_uuid"`
	}

	testCases := []struct {
		name      string
		data      any
		expectErr bool
	}{
		{
			name: "valid uuid",
			data: testStruct{
				ID: types.UUID(uuid.New()),
			},
			expectErr: false,
		},
		{
			name: "empty uuid",
			data: testStruct{
				ID: types.UUID{},
			},
			expectErr: true,
		},
		{
			name: "wrong type",
			data: testStructWrongType{
				ID: "not-a-uuid",
			},
			expectErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			err := validator.Struct(tc.data)
			if tc.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
