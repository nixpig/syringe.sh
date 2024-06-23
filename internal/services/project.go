package services

import (
	"github.com/go-playground/validator/v10"
	"github.com/nixpig/syringe.sh/server/internal/stores"
	"github.com/nixpig/syringe.sh/server/pkg"
)

type AddProjectRequest struct {
	Name string `name:"project name" validate:"required,min=1,max=256"`
}

type RemoveProjectRequest struct {
	Name string `name:"project name" validate:"required,min=1,max=256"`
}

type RenameProjectRequest struct {
	Name    string `name:"project name" validate:"required,min=1,max=256"`
	NewName string `name:"new project name" validate:"required,min=1,max=256"`
}

type ListProjectsResponse struct {
	Projects []struct {
		ID   int
		Name string
	}
}

type ProjectService interface {
	Add(project AddProjectRequest) error
	Remove(project RemoveProjectRequest) error
	Rename(project RenameProjectRequest) error
	List() (*ListProjectsResponse, error)
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

func (p ProjectServiceImpl) Add(project AddProjectRequest) error {
	if err := p.validate.Struct(project); err != nil {
		return pkg.ValidationError(err)
	}

	if err := p.store.Add(project.Name); err != nil {
		return err
	}

	return nil
}

func (p ProjectServiceImpl) Remove(project RemoveProjectRequest) error {
	if err := p.validate.Struct(project); err != nil {
		return pkg.ValidationError(err)
	}

	if err := p.store.Remove(project.Name); err != nil {
		return err
	}

	return nil
}

func (p ProjectServiceImpl) Rename(project RenameProjectRequest) error {
	if err := p.validate.Struct(project); err != nil {
		return pkg.ValidationError(err)
	}

	if err := p.store.Rename(
		project.Name,
		project.NewName,
	); err != nil {
		return err
	}

	return nil
}

func (p ProjectServiceImpl) List() (*ListProjectsResponse, error) {
	projects, err := p.store.List()
	if err != nil {
		return nil, err
	}

	var projectsResponseList []struct {
		ID   int
		Name string
	}

	for _, pv := range *projects {
		projectsResponseList = append(projectsResponseList, struct {
			ID   int
			Name string
		}{
			ID:   pv.ID,
			Name: pv.Name,
		})
	}

	return &ListProjectsResponse{Projects: projectsResponseList}, nil
}
