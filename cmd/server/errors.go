package main

import (
	"errors"
	"fmt"
)

var (
	ErrServer  = errors.New("ErrServer")
	ErrCmd     = errors.New("ErrCmd")
	ErrTimeout = errors.New("ErrTimeout")
)

func newError(err error, id string) error {
	return fmt.Errorf("%w %s", err, id)
}
