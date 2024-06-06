package services

import (
	"github.com/go-playground/validator/v10"
	"github.com/nixpig/syringe.sh/server/internal/stores"
)

type RegisterUserRequestDto struct {
	Username  string `json:"username" validate:"required,min=3,max=30"`
	Email     string `json:"email" validate:"required,email"`
	PublicKey string `json:"public_key" validate:"required"`
}

type RegisterUserResponseDto struct {
	Id        int    `json:"id"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	CreatedAt string `json:"created_at"`
	PublicKey string `json:"public_key"`
}

type AddPublicKeyRequestDto struct {
	PublicKey string `json:"public_key" validate:"required"`
	UserId    int    `json:"user_id" validate:"required"`
}

type AddPublicKeyResponseDto struct {
	Id        int    `json:"id"`
	UserId    int    `json:"user_id"`
	PublicKey string `json:"public_key"`
	CreatedAt string `json:"created_at"`
}

type AppService interface {
	RegisterUser(user RegisterUserRequestDto) (*RegisterUserResponseDto, error)
	AddPublicKey(publicKey AddPublicKeyRequestDto) (*AddPublicKeyResponseDto, error)
}

type AppServiceImpl struct {
	store    stores.AppStore
	validate *validator.Validate
}

func NewAppServiceImpl(store stores.AppStore, validate *validator.Validate) AppServiceImpl {
	return AppServiceImpl{
		store:    store,
		validate: validate,
	}
}

func (a AppServiceImpl) RegisterUser(user RegisterUserRequestDto) (*RegisterUserResponseDto, error) {
	if err := a.validate.Struct(user); err != nil {
		return nil, err
	}

	insertedUser, err := a.store.InsertUser(
		user.Username,
		user.Email,
		"active",
	)
	if err != nil {
		return nil, err
	}

	insertedKey, err := a.store.InsertKey(
		insertedUser.Id,
		user.PublicKey,
	)
	if err != nil {
		return nil, err
	}

	return &RegisterUserResponseDto{
		Id:        insertedUser.Id,
		Username:  insertedUser.Username,
		Email:     insertedUser.Email,
		CreatedAt: insertedUser.CreatedAt,
		PublicKey: insertedKey.PublicKey,
	}, nil
}

func (a AppServiceImpl) AddPublicKey(addKeyDetails AddPublicKeyRequestDto) (*AddPublicKeyResponseDto, error) {
	if err := a.validate.Struct(addKeyDetails); err != nil {
		return nil, err
	}

	addedKeyDetails, err := a.store.InsertKey(addKeyDetails.UserId, addKeyDetails.PublicKey)
	if err != nil {
		return nil, err
	}

	return &AddPublicKeyResponseDto{
		Id:        addedKeyDetails.Id,
		UserId:    addedKeyDetails.UserId,
		PublicKey: addedKeyDetails.PublicKey,
		CreatedAt: addedKeyDetails.CreatedAt,
	}, nil
}
