package services

import (
	"github.com/go-playground/validator/v10"
	"github.com/nixpig/syringe.sh/server/internal/stores"
)

type SetSecretRequest struct {
	Project     string
	Environment string
	Key         string
	Value       string
}

type GetSecretRequest struct {
	Project     string
	Environment string
	Key         string
}

type GetSecretResponse struct {
	Id          int
	Project     string
	Environment string
	Key         string
	Value       string
}

type UserService interface {
	CreateTables() error
	SetSecret(secret SetSecretRequest) error
	GetSecret(request GetSecretRequest) (*GetSecretResponse, error)
}

type UserServiceImpl struct {
	store    stores.UserStore
	validate *validator.Validate
}

func NewUserServiceImpl(
	store stores.UserStore,
	validate *validator.Validate,
) UserService {
	return UserServiceImpl{
		store:    store,
		validate: validate,
	}
}

func (e UserServiceImpl) CreateTables() error {
	if err := e.store.CreateTables(); err != nil {
		return err
	}

	return nil
}

func (e UserServiceImpl) SetSecret(secret SetSecretRequest) error {
	if err := e.validate.Struct(secret); err != nil {
		return err
	}

	return e.store.InsertSecret(
		secret.Project,
		secret.Environment,
		secret.Key,
		secret.Value,
	)
}

func (e UserServiceImpl) GetSecret(request GetSecretRequest) (*GetSecretResponse, error) {
	if err := e.validate.Struct(request); err != nil {
		return nil, err
	}

	secret, err := e.store.GetSecret(
		request.Project,
		request.Environment,
		request.Key,
	)
	if err != nil {
		return nil, err
	}

	return &GetSecretResponse{
		Id:          secret.Id,
		Project:     secret.Project,
		Environment: secret.Environment,
		Key:         secret.Key,
		Value:       secret.Value,
	}, nil
}
