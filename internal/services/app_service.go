package services

import (
	"github.com/go-playground/validator/v10"
)

type RegisterUserRequestDto struct {
	Username  string `json:"username" validate:"required,min=3,max=30"`
	Email     string `json:"email" validate:"required,email"`
	PublicKey string `json:"public_key" validate:"required"`
}

type UserDetailsResponseDto struct {
	Id        int    `json:"id"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	CreatedAt string `json:"created_at"`
}

type UserService interface {
	Create(user RegisterUserRequestDto) (*UserDetailsResponseDto, error)
}

type UserServiceImpl struct {
	store    UserStore
	validate *validator.Validate
}

func NewJsonUserService(store UserStore, validate *validator.Validate) UserServiceImpl {
	return UserServiceImpl{
		store:    store,
		validate: validate,
	}
}

func (u UserServiceImpl) Create(user RegisterUserRequestDto) (*UserDetailsResponseDto, error) {
	if err := u.validate.Struct(user); err != nil {
		return nil, err
	}

	insertedUser, err := u.store.Insert(user.Username, user.Email, user.PublicKey)
	if err != nil {
		return nil, err
	}

	return &UserDetailsResponseDto{
		Id:        insertedUser.Id,
		Username:  insertedUser.Username,
		Email:     insertedUser.Email,
		CreatedAt: insertedUser.CreatedAt,
	}, nil
}
