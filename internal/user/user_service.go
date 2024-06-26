package user

import (
	"crypto/sha1"
	"fmt"
	"net/http"
	"os"
	"slices"
	"time"

	"github.com/charmbracelet/ssh"
	"github.com/go-playground/validator/v10"
	"github.com/nixpig/syringe.sh/internal/database"
	"github.com/nixpig/syringe.sh/internal/secret"
	"github.com/nixpig/syringe.sh/pkg/turso"
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

type TursoAPISettings struct {
	URL   string
	Token string
}

type UserService interface {
	RegisterUser(user RegisterUserRequest) (*RegisterUserResponse, error)
	AddPublicKey(publicKey AddPublicKeyRequest) (*AddPublicKeyResponse, error)
	CreateDatabase(databaseDetails CreateDatabaseRequest) (*CreateDatabaseResponse, error)
}

type UserServiceImpl struct {
	store            UserStore
	validate         *validator.Validate
	httpClient       http.Client
	tursoAPISettings TursoAPISettings
}

func NewUserServiceImpl(
	store UserStore,
	validate *validator.Validate,
	httpClient http.Client,
	tursoAPISettings TursoAPISettings,
) UserServiceImpl {
	return UserServiceImpl{
		store:            store,
		validate:         validate,
		httpClient:       httpClient,
		tursoAPISettings: tursoAPISettings,
	}
}

func (u UserServiceImpl) RegisterUser(
	user RegisterUserRequest,
) (*RegisterUserResponse, error) {
	if err := u.validate.Struct(user); err != nil {
		return nil, err
	}

	insertedUser, err := u.store.InsertUser(
		user.Username,
		user.Email,
		"active",
	)
	if err != nil {
		return nil, err
	}

	marshalledKey := gossh.MarshalAuthorizedKey(user.PublicKey)

	insertedKey, err := u.AddPublicKey(AddPublicKeyRequest{
		UserID:    insertedUser.ID,
		PublicKey: string(marshalledKey),
	})
	if err != nil {
		return nil, err
	}

	insertedDatabase, err := u.CreateDatabase(
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

func (u UserServiceImpl) AddPublicKey(
	addKeyDetails AddPublicKeyRequest,
) (*AddPublicKeyResponse, error) {
	if err := u.validate.Struct(addKeyDetails); err != nil {
		return nil, err
	}

	addedKeyDetails, err := u.store.InsertKey(addKeyDetails.UserID, addKeyDetails.PublicKey)
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

func (u UserServiceImpl) CreateDatabase(
	databaseDetails CreateDatabaseRequest,
) (*CreateDatabaseResponse, error) {
	if err := u.validate.Struct(databaseDetails); err != nil {
		return nil, err
	}

	api := turso.New(databaseDetails.DatabaseOrg, u.tursoAPISettings.Token, u.httpClient)

	list, err := api.ListDatabases()
	if err != nil {
		return nil, err
	}

	exists := slices.IndexFunc(list.Databases, func(db turso.TursoDatabase) bool {
		return db.Name == databaseDetails.Name
	})

	if exists != -1 {
		return nil, fmt.Errorf("database already exists in returned list")
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

	envStore := secret.NewSqliteSecretStore(userDB)
	envService := secret.NewSecretServiceImpl(envStore, validator.New())

	var count time.Duration
	increment := time.Second * 5
	timeout := time.Second * 60
	for err := envService.CreateTables(); err != nil; err = envService.CreateTables() {
		time.Sleep(increment)
		count = count + increment
		if count >= timeout {
			return nil, fmt.Errorf(
				fmt.Sprintf(
					"timed out after %d seconds trying to create tables",
					timeout/time.Second,
				),
			)
		}
	}

	return &CreateDatabaseResponse{Name: createdDatabaseDetails.Database.Name}, nil
}
