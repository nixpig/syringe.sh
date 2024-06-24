package serrors

import (
	"errors"
	"fmt"

	"github.com/go-playground/validator/v10"
)

func NewError(err error, msg string) Error {
	return Error{
		err: fmt.Errorf("%s: %w", msg, err),
		msg: msg,
	}
}

type Error struct {
	err error
	msg string
}

func (e Error) Error() string {
	return e.msg
}

func (e Error) Unwrap() error {
	return e.err
}

func ErrDatabaseExec(err error) error {
	return NewError(err, "database exec error")
}

func ErrDatabaseQuery(err error) error {
	return NewError(err, "database query error")
}

func ErrNoProjects(err error) error {
	return NewError(err, "no projects found")
}

var (
	ErrNoProjectsFound     = fmt.Errorf("no projects found")
	ErrNoEnvironmentsFound = fmt.Errorf("no environments found")
	ErrNoSecretsFound      = fmt.Errorf("no secrets found")
	ErrProjectNotFound     = fmt.Errorf("project not found")
	ErrEnvironmentNotFound = fmt.Errorf("environment not found")
	ErrSecretNotFound      = fmt.Errorf("secret not found")
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
