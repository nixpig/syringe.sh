package services

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/nixpig/syringe.sh/server/internal/stores"
)

type RegisterUserRequest struct {
	Username  string `json:"username" validate:"required,min=3,max=30"`
	Email     string `json:"email" validate:"required,email"`
	PublicKey string `json:"public_key" validate:"required"`
}

type RegisterUserResponse struct {
	Id           int    `json:"id"`
	Username     string `json:"username"`
	Email        string `json:"email"`
	CreatedAt    string `json:"created_at"`
	PublicKey    string `json:"public_key"`
	DatabaseName string `json:"database_name"`
}

type AddPublicKeyRequest struct {
	PublicKey string `json:"public_key" validate:"required"`
	UserId    int    `json:"user_id" validate:"required"`
}

type AddPublicKeyResponse struct {
	Id        int    `json:"id"`
	UserId    int    `json:"user_id"`
	PublicKey string `json:"public_key"`
	CreatedAt string `json:"created_at"`
}

type CreateDatabaseRequest struct {
	Name          string `json:"name" validate:"required"`
	UserId        int    `json:"user_id" validate:"required"`
	DatabaseGroup string `json:"database_group" validate:"required"`
	DatabaseOrg   string `json:"database_org" validate:"required"`
}

type CreateDatabaseResponse struct {
	Id        int    `json:"id"`
	Name      string `json:"name"`
	UserId    int    `json:"user_id"`
	CreatedAt string `json:"created_at"`
}

type TursoApiSettings struct {
	Url   string
	Token string
}

type AppService interface {
	RegisterUser(user RegisterUserRequest) (*RegisterUserResponse, error)
	AddPublicKey(publicKey AddPublicKeyRequest) (*AddPublicKeyResponse, error)
	CreateDatabase(databaseDetails CreateDatabaseRequest) (*CreateDatabaseResponse, error)
}

type AppServiceImpl struct {
	store            stores.AppStore
	validate         *validator.Validate
	httpClient       http.Client
	tursoApiSettings TursoApiSettings
}

func NewAppServiceImpl(
	store stores.AppStore,
	validate *validator.Validate,
	httpClient http.Client,
	tursoApiSettings TursoApiSettings,
) AppServiceImpl {
	return AppServiceImpl{
		store:            store,
		validate:         validate,
		httpClient:       httpClient,
		tursoApiSettings: tursoApiSettings,
	}
}

func (a AppServiceImpl) RegisterUser(user RegisterUserRequest) (*RegisterUserResponse, error) {
	if err := a.validate.Struct(user); err != nil {
		return nil, err
	}

	insertedUser, err := a.store.InsertUser(
		user.Username,
		user.Email,
		"active",
	)
	if err != nil {
		return nil, err
	}

	insertedKey, err := a.AddPublicKey(AddPublicKeyRequest{
		UserId:    insertedUser.Id,
		PublicKey: user.PublicKey,
	})
	if err != nil {
		return nil, err
	}

	insertedDatabase, err := a.CreateDatabase(
		CreateDatabaseRequest{
			Name:          insertedUser.Username + "-" + strconv.Itoa(insertedUser.Id),
			UserId:        insertedUser.Id,
			DatabaseOrg:   os.Getenv("DATABASE_ORG"),
			DatabaseGroup: os.Getenv("DATABASE_GROUP"),
		})
	if err != nil {
		return nil, err
	}

	return &RegisterUserResponse{
		Id:           insertedUser.Id,
		Username:     insertedUser.Username,
		Email:        insertedUser.Email,
		CreatedAt:    insertedUser.CreatedAt,
		PublicKey:    insertedKey.PublicKey,
		DatabaseName: insertedDatabase.Name,
	}, nil
}

func (a AppServiceImpl) AddPublicKey(addKeyDetails AddPublicKeyRequest) (*AddPublicKeyResponse, error) {
	if err := a.validate.Struct(addKeyDetails); err != nil {
		return nil, err
	}

	addedKeyDetails, err := a.store.InsertKey(addKeyDetails.UserId, addKeyDetails.PublicKey)
	if err != nil {
		return nil, err
	}

	return &AddPublicKeyResponse{
		Id:        addedKeyDetails.Id,
		UserId:    addedKeyDetails.UserId,
		PublicKey: addedKeyDetails.PublicKey,
		CreatedAt: addedKeyDetails.CreatedAt,
	}, nil
}

func (a AppServiceImpl) CreateDatabase(databaseDetails CreateDatabaseRequest) (*CreateDatabaseResponse, error) {
	if err := a.validate.Struct(databaseDetails); err != nil {
		return nil, err
	}

	createDatabaseUrl := a.tursoApiSettings.Url + "/organizations/" + databaseDetails.DatabaseOrg + "/databases"

	body := []byte(fmt.Sprintf(`{
		"name": "%s",
		"group": "%s"
	}`, databaseDetails.Name, databaseDetails.DatabaseGroup))

	req, err := http.NewRequest("POST", createDatabaseUrl, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", a.tursoApiSettings.Token))

	res, err := a.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	if res.StatusCode < 200 || res.StatusCode > 299 {
		return nil, errors.New(fmt.Sprintf("api error: %v", res))
	}

	createdDatabaseDetails, err := a.store.InsertDatabase(
		databaseDetails.Name,
		databaseDetails.UserId,
	)

	if err != nil {
		return nil, err
	}

	return &CreateDatabaseResponse{
		Id:        createdDatabaseDetails.Id,
		Name:      createdDatabaseDetails.Name,
		UserId:    createdDatabaseDetails.UserId,
		CreatedAt: createdDatabaseDetails.CreatedAt,
	}, nil
}
