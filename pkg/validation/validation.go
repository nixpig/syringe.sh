package validation

import (
	"reflect"

	"github.com/go-playground/validator/v10"
)

type Validator interface {
	Struct(s interface{}) error
}

func New() Validate {
	v := validator.
		New(validator.WithRequiredStructEnabled())

	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		return fld.Tag.Get("name")
	})

	return Validate{
		structValidator: v.Struct,
	}
}

type Validate struct {
	structValidator func(s interface{}) error
}

func (v Validate) Struct(s interface{}) error {
	return v.structValidator(s)
}
