package user

import (
	"time"

	"github.com/go-playground/validator/v10"
)

type RegisterUserRequestJsonDto struct {
	Username string `json:"username" validate:"required,min=3,max=30"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type UserDetailsResponseJsonDto struct {
	Id        int       `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

type UserService interface {
	Create(user RegisterUserRequestJsonDto) (*UserDetailsResponseJsonDto, error)
}

type JsonUserService struct {
	store    UserStore
	validate *validator.Validate
}

func NewJsonUserService(store UserStore, validate *validator.Validate) JsonUserService {
	return JsonUserService{
		store:    store,
		validate: validate,
	}
}

func (u JsonUserService) Create(user RegisterUserRequestJsonDto) (*UserDetailsResponseJsonDto, error) {
	if err := u.validate.Struct(user); err != nil {
		return nil, err
	}

	insertedUser, err := u.store.Insert(user.Username, user.Email, user.Password)
	if err != nil {
		return nil, err
	}

	return &UserDetailsResponseJsonDto{
		Id:        insertedUser.Id,
		Username:  insertedUser.Username,
		Email:     insertedUser.Email,
		CreatedAt: insertedUser.CreatedAt,
	}, nil
}
