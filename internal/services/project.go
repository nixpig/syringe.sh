package services

import (
	"github.com/go-playground/validator/v10"
	"github.com/nixpig/syringe.sh/server/internal/stores"
)

type AddProjectRequest struct {
	Name string
}

type ProjectService interface {
	AddProject(project AddProjectRequest) error
}

func NewProjectServiceImpl(
	store stores.ProjectStore,
	validate *validator.Validate,
) ProjectService {
	return ProjectServiceImpl{
		store:    store,
		validate: validate,
	}
}

type ProjectServiceImpl struct {
	store    stores.ProjectStore
	validate *validator.Validate
}

func (p ProjectServiceImpl) AddProject(project AddProjectRequest) error {
	if err := p.validate.Struct(project); err != nil {
		return err
	}

	if err := p.store.InsertProject(project.Name); err != nil {
		return err
	}

	return nil
}
