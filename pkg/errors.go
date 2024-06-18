package pkg

import (
	"errors"
	"fmt"

	"github.com/go-playground/validator/v10"
)

var (
	ErrNoProjectsFound = fmt.Errorf("no projects found")
)

type FormattedError struct {
	Err error
}

func (fe FormattedError) Error() string {
	return formatError(fe.Err).Error()
}

func formatError(err error) error {
	switch t := err.(type) {
	case validator.ValidationErrors:
		var errs []error

		for _, e := range t {
			switch tag := e.Tag(); tag {
			case "max":
				errs = append(errs, fmt.Errorf(
					"\"%s\" exceeds max length of %s characters",
					e.Field(),
					e.Param(),
				))
			}
		}

		return errors.Join(errs...)

	}

	return err
}
