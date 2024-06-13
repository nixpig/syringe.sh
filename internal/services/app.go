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
	Username  string
	Email     string
	PublicKey ssh.PublicKey
}

type RegisterUserResponse struct {
	ID           int
	Username     string
	Email        string
	CreatedAt    string
	PublicKey    string
	DatabaseName string
}

type AddPublicKeyRequest struct {
	PublicKey string
	UserID    int
}

type AddPublicKeyResponse struct {
	ID        int
	UserID    int
	PublicKey string
	CreatedAt string
}

type CreateDatabaseRequest struct {
	Name          string
	UserID        int
	DatabaseGroup string
	DatabaseOrg   string
}

type CreateDatabaseResponse struct {
	Name     string
	HostName string
}

type UserAuthRequest struct {
	Username  string
	PublicKey ssh.PublicKey
}

type UserAuthResponse struct {
	Auth bool
}

type TursoAPISettings struct {
	URL   string
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
	tursoAPISettings TursoAPISettings
}

func NewAppService(
	store stores.AppStore,
	validate *validator.Validate,
	httpClient http.Client,
	tursoAPISettings TursoAPISettings,
) AppServiceImpl {
	return AppServiceImpl{
		store:            store,
		validate:         validate,
		httpClient:       httpClient,
		tursoAPISettings: tursoAPISettings,
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
		UserID:    insertedUser.ID,
		PublicKey: string(marshalledKey),
	})
	if err != nil {
		return nil, err
	}

	insertedDatabase, err := a.CreateDatabase(
		CreateDatabaseRequest{
			Name:          fmt.Sprintf("%x", sha1.Sum(marshalledKey)),
			UserID:        insertedUser.ID,
			DatabaseOrg:   os.Getenv("DATABASE_ORG"),
			DatabaseGroup: os.Getenv("DATABASE_GROUP"),
		})
	if err != nil {
		return nil, err
	}

	return &RegisterUserResponse{
		ID:           insertedUser.ID,
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

	addedKeyDetails, err := a.store.InsertKey(addKeyDetails.UserID, addKeyDetails.PublicKey)
	if err != nil {
		return nil, err
	}

	return &AddPublicKeyResponse{
		ID:        addedKeyDetails.ID,
		UserID:    addedKeyDetails.UserID,
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

	api := turso.New(databaseDetails.DatabaseOrg, a.tursoAPISettings.Token, a.httpClient)

	list, err := api.ListDatabases()
	if err != nil {
		return nil, err
	}

	exists := slices.IndexFunc(list.Databases, func(db turso.TursoDatabase) bool {
		return db.Name == databaseDetails.Name
	})

	if exists != -1 {
		return nil, errors.New("database already exists in returned list")
	}

	createdDatabaseDetails, err := api.CreateDatabase(databaseDetails.Name, databaseDetails.DatabaseGroup)
	if err != nil {
		return nil, err
	}

	createdToken, err := api.CreateToken(createdDatabaseDetails.Database.Name, "5m")
	if err != nil {
		return nil, err
	}

	userDB, err := database.Connection(
		"libsql://"+createdDatabaseDetails.Database.HostName,
		createdToken.Jwt,
	)
	if err != nil {
		return nil, err
	}

	envStore := stores.NewSqliteSecretStore(userDB)
	envService := NewSecretServiceImpl(envStore, validator.New())

	var count time.Duration
	increment := time.Second * 5
	timeout := time.Second * 60
	fmt.Println("creating tables...")
	for err := envService.CreateTables(); err != nil; err = envService.CreateTables() {
		fmt.Printf("sleep for %d seconds\n", increment/time.Second)
		time.Sleep(increment)
		count = count + increment
		if count >= timeout {
			return nil, fmt.Errorf(fmt.Sprintf("timed out after %d seconds trying to create tables", timeout/time.Second))
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
