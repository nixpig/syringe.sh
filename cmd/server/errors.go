package main

import "errors"

var (
	ErrServer  = errors.New("ErrServer")
	ErrCmd     = errors.New("ErrCmd")
	ErrTimeout = errors.New("ErrTimeout")
)
