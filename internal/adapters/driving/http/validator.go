package http

import (
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/oapi-codegen/runtime/types"
)

var (
	registerUUIDValidatorImpl = registerUUIDValidator
)

func MustNewValidator() *validator.Validate {
	v := validator.New()

	err := registerUUIDValidatorImpl(v)
	if err != nil {
		panic(err)
	}

	return v
}

func registerUUIDValidator(v *validator.Validate) error {
	err := v.RegisterValidation("oapi_uuid", func(fl validator.FieldLevel) bool {
		u, ok := fl.Field().Interface().(types.UUID)
		if !ok {
			return false
		}
		return u != types.UUID{}
	})
	if err != nil {
		return fmt.Errorf("failed to register oapi_uuid validator: %w", err)
	}
	return nil
}
