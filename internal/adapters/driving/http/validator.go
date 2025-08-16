package http

import (
	"github.com/go-playground/validator/v10"
	"github.com/oapi-codegen/runtime/types"
)

func NewValidator() *validator.Validate {
	v := validator.New()

	registerUUIDValidator(v)

	return v
}

func registerUUIDValidator(v *validator.Validate) {
	v.RegisterValidation("oapi_uuid", func(fl validator.FieldLevel) bool {
		u, ok := fl.Field().Interface().(types.UUID)
		if !ok {
			return false
		}
		return u != types.UUID{}
	})
}
