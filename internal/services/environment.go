package services

import (
	"reflect"

	"github.com/go-playground/validator/v10"
	"github.com/nixpig/syringe.sh/server/internal/stores"
)

type AddEnvironmentRequest struct {
	Name        string `name:"environment name" validate:"required,min=1,max=256"`
	ProjectName string `name:"project name" validate:"required,min=1,max=256"`
}

type RemoveEnvironmentRequest struct {
	Name        string `name:"environment name" validate:"required,min=1,max=256"`
	ProjectName string `name:"project name" validate:"required,min=1,max=256"`
}

type RenameEnvironmentRequest struct {
	Name        string `name:"environment name" validate:"required,min=1,max=256"`
	NewName     string `name:"new environment name" validate:"required,min=1,max=256"`
	ProjectName string `name:"project name" validate:"required,min=1,max=256"`
}

type EnvironmentService interface {
	Add(environment AddEnvironmentRequest) error
	Remove(environment RemoveEnvironmentRequest) error
	Rename(environment RenameEnvironmentRequest) error
}

func NewEnvironmentServiceImpl(
	store stores.EnvironmentStore,
	validate *validator.Validate,
) EnvironmentService {
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		return fld.Tag.Get("name")
	})

	return EnvironmentServiceImpl{
		store:    store,
		validate: validate,
	}
}

type EnvironmentServiceImpl struct {
	store    stores.EnvironmentStore
	validate *validator.Validate
}

func (e EnvironmentServiceImpl) Add(
	environment AddEnvironmentRequest,
) error {
	if err := e.validate.Struct(environment); err != nil {
		return err
	}

	if err := e.store.Add(
		environment.Name,
		environment.ProjectName,
	); err != nil {
		return err
	}

	return nil
}

func (e EnvironmentServiceImpl) Remove(
	environment RemoveEnvironmentRequest,
) error {
	if err := e.validate.Struct(environment); err != nil {
		return err
	}

	if err := e.store.Remove(
		environment.Name,
		environment.ProjectName,
	); err != nil {
		return err
	}

	return nil
}

func (e EnvironmentServiceImpl) Rename(
	environment RenameEnvironmentRequest,
) error {
	if err := e.validate.Struct(environment); err != nil {
		return err
	}

	if err := e.store.Rename(
		environment.Name,
		environment.NewName,
		environment.ProjectName,
	); err != nil {
		return err
	}

	return nil
}
