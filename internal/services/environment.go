package services

import (
	"github.com/go-playground/validator/v10"
	"github.com/nixpig/syringe.sh/server/internal/stores"
)

type AddEnvironmentRequest struct {
	Name        string
	ProjectName string
}

type EnvironmentService interface {
	AddEnvironment(environment AddEnvironmentRequest) error
}

func NewEnvironmentServiceImpl(
	store stores.EnvironmentStore,
	validate *validator.Validate,
) EnvironmentService {
	return EnvironmentServiceImpl{
		store:    store,
		validate: validator.New(validator.WithRequiredStructEnabled()),
	}
}

type EnvironmentServiceImpl struct {
	store    stores.EnvironmentStore
	validate *validator.Validate
}

func (e EnvironmentServiceImpl) AddEnvironment(environment AddEnvironmentRequest) error {
	if err := e.validate.Struct(environment); err != nil {
		return err
	}

	if err := e.store.InsertEnvironment(environment.Name, environment.ProjectName); err != nil {
		return err
	}

	return nil
}
