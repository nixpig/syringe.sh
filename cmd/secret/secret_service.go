package secret

import (
	"github.com/go-playground/validator/v10"
	"github.com/nixpig/syringe.sh/server/pkg/serrors"
)

type SetSecretRequest struct {
	Project     string `name:"project name" validate:"required,min=1,max=256"`
	Environment string `name:"environment name" validate:"required,min=1,max=256"`
	Key         string `name:"secret key" validate:"required,min=1,max=256"`
	Value       string `name:"secret name" validate:"required,min=1,max=256"`
}

type GetSecretRequest struct {
	Project     string `name:"project name" validate:"required,min=1,max=256"`
	Environment string `name:"environment name" validate:"required,min=1,max=256"`
	Key         string `name:"secret key" validate:"required,min=1,max=256"`
}

type ListSecretsRequest struct {
	Project     string `name:"project name" validate:"required,min=1,max=256"`
	Environment string `name:"environment name" validate:"required,min=1,max=256"`
}

type RemoveSecretRequest struct {
	Project     string `name:"project name" validate:"required,min=1,max=256"`
	Environment string `name:"environment name" validate:"required,min=1,max=256"`
	Key         string `name:"secret key" validate:"required,min=1,max=256"`
}

type GetSecretResponse struct {
	ID          int
	Project     string
	Environment string
	Key         string
	Value       string
}

type ListSecretsResponse struct {
	Project     string
	Environment string
	Secrets     []struct {
		ID    int
		Key   string
		Value string
	}
}

type SecretService interface {
	CreateTables() error
	Set(secret SetSecretRequest) error
	Get(request GetSecretRequest) (*GetSecretResponse, error)
	List(request ListSecretsRequest) (*ListSecretsResponse, error)
	Remove(request RemoveSecretRequest) error
}

type SecretServiceImpl struct {
	store    SecretStore
	validate *validator.Validate
}

func NewSecretServiceImpl(
	store SecretStore,
	validate *validator.Validate,
) SecretService {
	return SecretServiceImpl{
		store:    store,
		validate: validate,
	}
}

func (s SecretServiceImpl) CreateTables() error {
	if err := s.store.CreateTables(); err != nil {
		return err
	}

	return nil
}

func (s SecretServiceImpl) Set(secret SetSecretRequest) error {
	if err := s.validate.Struct(secret); err != nil {
		return serrors.ValidationError(err)
	}

	if err := s.store.Set(
		secret.Project,
		secret.Environment,
		secret.Key,
		secret.Value,
	); err != nil {
		return err
	}

	return nil
}

func (s SecretServiceImpl) Get(request GetSecretRequest) (*GetSecretResponse, error) {
	if err := s.validate.Struct(request); err != nil {
		return nil, serrors.ValidationError(err)
	}

	secret, err := s.store.Get(
		request.Project,
		request.Environment,
		request.Key,
	)
	if err != nil {
		return nil, err
	}

	return &GetSecretResponse{
		ID:          secret.ID,
		Project:     secret.Project,
		Environment: secret.Environment,
		Key:         secret.Key,
		Value:       secret.Value,
	}, nil
}

func (s SecretServiceImpl) List(request ListSecretsRequest) (*ListSecretsResponse, error) {
	if err := s.validate.Struct(request); err != nil {
		return nil, serrors.ValidationError(err)
	}

	secrets, err := s.store.List(request.Project, request.Environment)
	if err != nil {
		return nil, err
	}

	var secretsResponseList []struct {
		ID    int
		Key   string
		Value string
	}

	for _, sv := range *secrets {
		secretsResponseList = append(secretsResponseList, struct {
			ID    int
			Key   string
			Value string
		}{
			ID:    sv.ID,
			Key:   sv.Key,
			Value: sv.Value,
		})
	}

	return &ListSecretsResponse{
		Project:     request.Project,
		Environment: request.Environment,
		Secrets:     secretsResponseList,
	}, nil
}

func (s SecretServiceImpl) Remove(request RemoveSecretRequest) error {
	if err := s.validate.Struct(request); err != nil {
		return serrors.ValidationError(err)
	}

	if err := s.store.Remove(
		request.Project,
		request.Environment,
		request.Key,
	); err != nil {
		return err
	}

	return nil
}
