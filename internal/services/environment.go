package services

import (
	"github.com/go-playground/validator/v10"
	"github.com/nixpig/syringe.sh/server/internal/stores"
)

type AddEnvironmentRequest struct {
	Name        string `validate:"required,min=1,max=256"`
	ProjectName string `validate:"required,min=1,max=256"`
}

type EnvironmentService interface {
	Add(environment AddEnvironmentRequest) error
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

func (e EnvironmentServiceImpl) Add(environment AddEnvironmentRequest) error {
	if err := e.validate.Struct(environment); err != nil {
		return err
	}

	if err := e.store.Add(environment.Name, environment.ProjectName); err != nil {
		return err
	}

	return nil
}
