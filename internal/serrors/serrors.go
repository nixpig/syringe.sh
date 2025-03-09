package serrors

import (
	"errors"
	"fmt"
)

var serrors = map[string]error{
	"server":  errors.New("ErrServer"),
	"cmd":     errors.New("ErrCmd"),
	"timeout": errors.New("ErrTimeout"),
}

func New(err string, sid string) error {
	e, ok := serrors[err]
	if !ok {
		e = errors.New("ErrUnknown")
	}

	return fmt.Errorf("%w %s", e, sid)
}
