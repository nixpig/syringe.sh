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

type EnvService interface {
	CreateTables() error
	SetSecret(secret SetSecretRequest) error
}

type EnvServiceImpl struct {
	store    stores.EnvStore
	validate *validator.Validate
}

func NewEnvServiceImpl(
	store stores.EnvStore,
	validate *validator.Validate,
) EnvService {
	return EnvServiceImpl{
		store:    store,
		validate: validate,
	}
}

func (e EnvServiceImpl) CreateTables() error {
	if err := e.store.CreateTables(); err != nil {
		return err
	}

	return nil
}

func (e EnvServiceImpl) SetSecret(secret SetSecretRequest) error {
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
