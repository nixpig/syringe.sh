package validation

import (
	"reflect"

	"github.com/go-playground/validator/v10"
)

func NewValidator() *validator.Validate {
	v := validator.
		New(validator.WithRequiredStructEnabled())

	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		return fld.Tag.Get("name")
	})

	return v
}
