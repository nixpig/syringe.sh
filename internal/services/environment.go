package services

import (
	"github.com/go-playground/validator/v10"
	"github.com/nixpig/syringe.sh/server/internal/stores"
	"github.com/nixpig/syringe.sh/server/pkg"
)

type AddEnvironmentRequest struct {
	Name    string `name:"environment name" validate:"required,min=1,max=256"`
	Project string `name:"project name" validate:"required,min=1,max=256"`
}

type RemoveEnvironmentRequest struct {
	Name    string `name:"environment name" validate:"required,min=1,max=256"`
	Project string `name:"project name" validate:"required,min=1,max=256"`
}

type RenameEnvironmentRequest struct {
	Name    string `name:"environment name" validate:"required,min=1,max=256"`
	NewName string `name:"new environment name" validate:"required,min=1,max=256"`
	Project string `name:"project name" validate:"required,min=1,max=256"`
}

type ListEnvironmentRequest struct {
	Project string `name:"project name" validate:"required,min=1,max=256"`
}

type ListEnvironmentsResponse struct {
	Project      string
	Environments []struct {
		ID   int
		Name string
	}
}

type EnvironmentService interface {
	Add(environment AddEnvironmentRequest) error
	Remove(environment RemoveEnvironmentRequest) error
	Rename(environment RenameEnvironmentRequest) error
	List(project ListEnvironmentRequest) (*ListEnvironmentsResponse, error)
}

func NewEnvironmentServiceImpl(
	store stores.EnvironmentStore,
	validate *validator.Validate,
) EnvironmentService {
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
		return pkg.ValidationError(err)
	}

	if err := e.store.Add(
		environment.Name,
		environment.Project,
	); err != nil {
		return err
	}

	return nil
}

func (e EnvironmentServiceImpl) Remove(
	environment RemoveEnvironmentRequest,
) error {
	if err := e.validate.Struct(environment); err != nil {
		return pkg.ValidationError(err)
	}

	if err := e.store.Remove(
		environment.Name,
		environment.Project,
	); err != nil {
		return err
	}

	return nil
}

func (e EnvironmentServiceImpl) Rename(
	environment RenameEnvironmentRequest,
) error {
	if err := e.validate.Struct(environment); err != nil {
		return pkg.ValidationError(err)
	}

	if err := e.store.Rename(
		environment.Name,
		environment.NewName,
		environment.Project,
	); err != nil {
		return err
	}

	return nil
}

func (e EnvironmentServiceImpl) List(
	request ListEnvironmentRequest,
) (*ListEnvironmentsResponse, error) {
	if err := e.validate.Struct(request); err != nil {
		return nil, pkg.ValidationError(err)
	}

	environments, err := e.store.List(request.Project)
	if err != nil {
		return nil, err
	}

	var environmentsResponseList []struct {
		ID   int
		Name string
	}

	for _, ev := range *environments {
		environmentsResponseList = append(environmentsResponseList, struct {
			ID   int
			Name string
		}{
			ID:   ev.ID,
			Name: ev.Name,
		},
		)
	}

	return &ListEnvironmentsResponse{
		Project:      request.Project,
		Environments: environmentsResponseList,
	}, nil
}
