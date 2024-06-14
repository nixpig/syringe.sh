package services

import (
	"github.com/go-playground/validator/v10"
	"github.com/nixpig/syringe.sh/server/internal/stores"
)

type ProjectService interface {
	Add(projectName string) error
	Remove(projectName string) error
	Rename(originalName, newName string) error
	List() ([]string, error)
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

func (p ProjectServiceImpl) Add(projectName string) error {
	if err := p.store.Add(projectName); err != nil {
		return err
	}

	return nil
}

func (p ProjectServiceImpl) Remove(projectName string) error {
	if err := p.store.Remove(projectName); err != nil {
		return err
	}

	return nil
}

func (p ProjectServiceImpl) Rename(originalName, newName string) error {
	if err := p.store.Rename(originalName, newName); err != nil {
		return err
	}

	return nil
}

func (p ProjectServiceImpl) List() ([]string, error) {
	projects, err := p.store.List()
	if err != nil {
		return nil, err
	}

	return projects, nil
}
