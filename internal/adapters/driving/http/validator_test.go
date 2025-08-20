package http

import (
	"errors"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/oapi-codegen/runtime/types"
	"github.com/stretchr/testify/assert"
)

func TestMustNewValidator(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		t.Parallel()
		assert.NotPanics(t, func() {
			v := MustNewValidator()
			assert.NotNil(t, v)
		})
	})

	t.Run("panic", func(t *testing.T) {
		originalImpl := registerUUIDValidatorImpl
		defer func() {
			registerUUIDValidatorImpl = originalImpl
		}()

		registerUUIDValidatorImpl = func(v *validator.Validate) error {
			return errors.New("mocked error")
		}

		assert.Panics(t, func() {
			MustNewValidator()
		})
	})
}

func TestOapiUUIDValidation(t *testing.T) {
	t.Parallel()

	v := MustNewValidator()

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
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			err := v.Struct(tc.data)
			if tc.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRegisterUUIDValidator(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		v := validator.New()
		err := registerUUIDValidator(v)
		assert.NoError(t, err)
	})
}
