package user

import (
	"crypto/sha1"
	"errors"
	"fmt"
	"net/http"
	"os"

	"github.com/charmbracelet/ssh"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "github.com/mattn/go-sqlite3"
	"github.com/nixpig/syringe.sh/internal/database"
	"github.com/nixpig/syringe.sh/migrations"
	"github.com/nixpig/syringe.sh/pkg/validation"
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

type UserService interface {
	RegisterUser(user RegisterUserRequest) (*RegisterUserResponse, error)
	AddPublicKey(publicKey AddPublicKeyRequest) (*AddPublicKeyResponse, error)
	CreateDatabase(databaseDetails CreateDatabaseRequest) (*CreateDatabaseResponse, error)
}

type UserServiceImpl struct {
	store      UserStore
	validate   validation.Validator
	httpClient http.Client
}

func NewUserServiceImpl(
	store UserStore,
	validate validation.Validator,
) UserServiceImpl {
	return UserServiceImpl{
		store:    store,
		validate: validate,
	}
}

func (u UserServiceImpl) RegisterUser(
	user RegisterUserRequest,
) (*RegisterUserResponse, error) {
	if err := u.validate.Struct(user); err != nil {
		return nil, err
	}

	exists, err := u.store.Exists(user.Username)
	if err != nil {
		return nil, err
	}

	if exists {
		return nil, fmt.Errorf("user '%s' already exists", user.Username)
	}

	marshalledKey := gossh.MarshalAuthorizedKey(user.PublicKey)
	databaseName := fmt.Sprintf("%x", sha1.Sum(marshalledKey))

	if _, err := os.Stat(databaseName); err == nil {
		return nil, fmt.Errorf("user '%s' already exists", user.Username)
	}

	insertedUser, err := u.store.InsertUser(
		user.Username,
		user.Email,
		"active",
	)
	if err != nil {
		return nil, err
	}

	insertedKey, err := u.AddPublicKey(AddPublicKeyRequest{
		UserID:    insertedUser.ID,
		PublicKey: string(marshalledKey),
	})
	if err != nil {
		return nil, err
	}

	insertedDatabase, err := u.CreateDatabase(
		CreateDatabaseRequest{
			Name:          databaseName,
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

	userDB, err := database.NewConnection(
		database.GetDatabasePath(fmt.Sprintf("%s.db", databaseDetails.Name)),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to open user database: %w", err)
	}

	migrations, err := iofs.New(migrations.User, "user")
	if err != nil {
		return nil, err
	}

	m, err := database.NewMigration(
		userDB,
		migrations,
	)
	if err != nil {
		return nil, err
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return nil, err
	}

	return &CreateDatabaseResponse{Name: databaseDetails.Name}, nil
}
