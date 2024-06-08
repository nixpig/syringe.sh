package services

import (
	"github.com/go-playground/validator/v10"
	"github.com/nixpig/syringe.sh/server/internal/stores"
)

type EnvService interface {
	CreateTables() error
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
