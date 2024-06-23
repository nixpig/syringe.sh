package pkg

import (
	"errors"
	"fmt"

	"github.com/go-playground/validator/v10"
)

var (
	ErrNoProjectsFound     = fmt.Errorf("no projects found")
	ErrNoEnvironmentsFound = fmt.Errorf("no environments found")
	ErrNoSecretsFound      = fmt.Errorf("no secrets found")
)

type ErrValidation struct{ msg string }

func (ve ErrValidation) Error() string {
	return ve.msg
}

func ValidationError(err error) error {
	switch t := err.(type) {
	case validator.ValidationErrors:
		var errs []error

		for _, e := range t {
			switch tag := e.Tag(); tag {
			case "max":
				errs = append(errs, ErrValidation{msg: fmt.Sprintf(
					"\"%s\" exceeds max length of %s characters",
					e.Field(),
					e.Param(),
				)})
			}
		}

		return errors.Join(errs...)
	}

	return err
}
