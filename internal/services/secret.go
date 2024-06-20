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

type GetSecretRequest struct {
	Project     string
	Environment string
	Key         string
}

type GetSecretResponse struct {
	ID          int
	Project     string
	Environment string
	Key         string
	Value       string
}

type ListSecretsRequest struct {
	Project     string
	Environment string
}

type ListSecretsResponse struct {
	Project     string
	Environment string
	Secrets     []*struct {
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
}

type SecretServiceImpl struct {
	store    stores.SecretStore
	validate *validator.Validate
}

func NewSecretServiceImpl(
	store stores.SecretStore,
	validate *validator.Validate,
) SecretService {
	return SecretServiceImpl{
		store:    store,
		validate: validate,
	}
}

func (e SecretServiceImpl) CreateTables() error {
	if err := e.store.CreateTables(); err != nil {
		return err
	}

	return nil
}

func (e SecretServiceImpl) Set(secret SetSecretRequest) error {
	if err := e.validate.Struct(secret); err != nil {
		return err
	}

	return e.store.Set(
		secret.Project,
		secret.Environment,
		secret.Key,
		secret.Value,
	)
}

func (e SecretServiceImpl) Get(request GetSecretRequest) (*GetSecretResponse, error) {
	if err := e.validate.Struct(request); err != nil {
		return nil, err
	}

	secret, err := e.store.Get(
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

func (e SecretServiceImpl) List(request ListSecretsRequest) (*ListSecretsResponse, error) {
	if err := e.validate.Struct(request); err != nil {
		return nil, err
	}

	secrets, err := e.store.List(request.Project, request.Environment)
	if err != nil {
		return nil, err
	}

	var secretsResponseList []*struct {
		ID    int
		Key   string
		Value string
	}

	for _, s := range secrets {
		secretsResponseList = append(secretsResponseList, &struct {
			ID    int
			Key   string
			Value string
		}{
			Key:   s.Key,
			Value: s.Value,
		})
	}

	return &ListSecretsResponse{
		Project:     request.Project,
		Environment: request.Environment,
		Secrets:     secretsResponseList,
	}, nil
}
