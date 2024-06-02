package user

import (
	"time"

	_ "github.com/go-playground/validator/v10"
)

type User struct {
	Id        int
	Username  string
	Email     string
	CreatedAt time.Time
	Password  string // for database??
}
