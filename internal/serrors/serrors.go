package serrors

import (
	"errors"
	"fmt"
)

var serrors = map[string]error{
	"server":  errors.New("ErrServer"),
	"cmd":     errors.New("ErrCmd"),
	"timeout": errors.New("ErrTimeout"),
	"user":    errors.New("ErrUserCreate"),
}

func New(errType, msg, id string) error {
	e, ok := serrors[errType]
	if !ok {
		e = errors.New("ErrUnknown")
	}

	return fmt.Errorf("%w: %s (id: %s)", e, msg, id)
}
