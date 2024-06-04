package user

import (
	_ "github.com/go-playground/validator/v10"
)

type User struct {
	Id        int
	Username  string
	Email     string
	CreatedAt string
	Status    string
}
