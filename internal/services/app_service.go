package services

import (
	"crypto/sha1"
	"errors"
	"fmt"
	"net/http"
	"os"
	"slices"
	"time"

	"github.com/charmbracelet/ssh"
	"github.com/go-playground/validator/v10"
	"github.com/nixpig/syringe.sh/server/internal/database"
	"github.com/nixpig/syringe.sh/server/internal/stores"
	"github.com/nixpig/syringe.sh/server/pkg/turso"
	gossh "golang.org/x/crypto/ssh"
)

type RegisterUserRequest struct {
	Username  string        `json:"username" validate:"required,min=3,max=30"`
	Email     string        `json:"email" validate:"required,email"`
	PublicKey ssh.PublicKey `json:"public_key" validate:"required"`
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
	Name     string `json:"name"`
	HostName string `json:"HostName"`
}

type UserAuthRequest struct {
	Username  string        `json:"username" validate:"required"`
	PublicKey ssh.PublicKey `json:"public_key" validate:"required"`
}

type UserAuthResponse struct {
	Auth bool `json:"auth"`
}

type TursoApiSettings struct {
	Url   string
	Token string
}

type AppService interface {
	RegisterUser(user RegisterUserRequest) (*RegisterUserResponse, error)
	AddPublicKey(publicKey AddPublicKeyRequest) (*AddPublicKeyResponse, error)
	CreateDatabase(databaseDetails CreateDatabaseRequest) (*CreateDatabaseResponse, error)
	AuthenticateUser(authDetails UserAuthRequest) (*UserAuthResponse, error)
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
) AppService {
	return AppServiceImpl{
		store:            store,
		validate:         validate,
		httpClient:       httpClient,
		tursoApiSettings: tursoApiSettings,
	}
}

func (a AppServiceImpl) RegisterUser(
	user RegisterUserRequest,
) (*RegisterUserResponse, error) {
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

	marshalledKey := gossh.MarshalAuthorizedKey(user.PublicKey)

	insertedKey, err := a.AddPublicKey(AddPublicKeyRequest{
		UserId:    insertedUser.Id,
		PublicKey: string(marshalledKey),
	})
	if err != nil {
		return nil, err
	}

	fmt.Println("creating user database")
	insertedDatabase, err := a.CreateDatabase(
		CreateDatabaseRequest{
			Name:          fmt.Sprintf("%x", sha1.Sum(marshalledKey)),
			UserId:        insertedUser.Id,
			DatabaseOrg:   os.Getenv("DATABASE_ORG"),
			DatabaseGroup: os.Getenv("DATABASE_GROUP"),
		})
	if err != nil {
		return nil, err
	}
	fmt.Println("created user database: ", insertedDatabase)

	return &RegisterUserResponse{
		Id:           insertedUser.Id,
		Username:     insertedUser.Username,
		Email:        insertedUser.Email,
		CreatedAt:    insertedUser.CreatedAt,
		PublicKey:    insertedKey.PublicKey,
		DatabaseName: insertedDatabase.Name,
	}, nil
}

func (a AppServiceImpl) AddPublicKey(
	addKeyDetails AddPublicKeyRequest,
) (*AddPublicKeyResponse, error) {
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

func (a AppServiceImpl) CreateDatabase(
	databaseDetails CreateDatabaseRequest,
) (*CreateDatabaseResponse, error) {
	if err := a.validate.Struct(databaseDetails); err != nil {
		return nil, err
	}

	api := turso.New(databaseDetails.DatabaseOrg, a.tursoApiSettings.Token, a.httpClient)

	list, err := api.ListDatabases()
	if err != nil {
		return nil, err
	}

	exists := slices.IndexFunc(list.Databases, func(db turso.TursoDatabase) bool {
		return db.Name == databaseDetails.Name
	})

	if exists != -1 {
		return nil, errors.New("database already exists in returned list!!")
	}

	createdDatabaseDetails, err := api.CreateDatabase(databaseDetails.Name, databaseDetails.DatabaseGroup)
	if err != nil {
		return nil, err
	}

	createdToken, err := api.CreateToken(createdDatabaseDetails.Database.Name, "5m")
	if err != nil {
		return nil, err
	}

	userDb, err := database.Connection(
		"libsql://"+createdDatabaseDetails.Database.HostName,
		createdToken.Jwt,
	)
	if err != nil {
		return nil, err
	}

	envStore := stores.NewSqliteEnvStore(userDb)
	envService := NewEnvServiceImpl(envStore, validator.New())

	var count time.Duration
	increment := time.Second * 30
	timeout := time.Second * 360
	for err := envService.CreateTables(); err != nil; err = envService.CreateTables() {
		time.Sleep(increment)
		count = count + increment
		if count >= timeout {
			break
		}
	}

	return &CreateDatabaseResponse{Name: createdDatabaseDetails.Database.Name}, nil
}

func (a AppServiceImpl) AuthenticateUser(
	authDetails UserAuthRequest,
) (*UserAuthResponse, error) {
	if err := a.validate.Struct(authDetails); err != nil {
		return nil, err
	}

	publicKeysDetails, err := a.store.GetUserPublicKeys(authDetails.Username)
	if err != nil {
		return nil, err
	}

	for _, v := range *publicKeysDetails {
		parsed, _, _, _, err := ssh.ParseAuthorizedKey([]byte(v.PublicKey))
		fmt.Println()
		if err != nil {
			return nil, err
		}

		if ssh.KeysEqual(
			authDetails.PublicKey,
			parsed,
		) {
			return &UserAuthResponse{Auth: true}, nil
		}
	}

	return &UserAuthResponse{Auth: false}, nil
}
