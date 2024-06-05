package stores

import (
	"database/sql"
	"errors"

	"github.com/nixpig/syringe.sh/server/internal/models"
)

type UserStore interface {
	InsertUser(username, email, status string) (*models.User, error)
	GetUserByUsername(username string) (*models.User, error)
	DeleteUserByUsername(username string) error
	UpdateUser(user models.User) (*models.User, error)
}

type KeyStore interface {
	InsertKey(userId int, publicKey string) (*models.Key, error)
}

type AppStore interface {
	UserStore
	KeyStore
}

type SqliteAppStore struct {
	db *sql.DB
}

func NewSqliteAppStore(db *sql.DB) SqliteAppStore {
	return SqliteAppStore{db}
}

func (s SqliteAppStore) InsertUser(username, email, status string) (*models.User, error) {
	query := `
		insert into users_ (username_, email_, status_) 
		values ($username, $email, $status) 
		returning id_, username_, email_, status_, created_at_
	`

	row := s.db.QueryRow(
		query,
		sql.Named("username", username),
		sql.Named("email", email),
		sql.Named("status", status),
	)

	var insertedUser models.User

	if err := row.Scan(
		&insertedUser.Id,
		&insertedUser.Username,
		&insertedUser.Email,
		&insertedUser.Status,
		&insertedUser.CreatedAt,
	); err != nil {
		return nil, err
	}

	return &insertedUser, nil
}

func (s SqliteAppStore) GetUserByUsername(username string) (*models.User, error) {
	query := `
		select id_, username_, email_, status_, created_at_ 
		from users_ 
		where username_ = $1
	`

	row := s.db.QueryRow(query, username)

	var user models.User

	if err := row.Scan(
		&user.Id,
		&user.Username,
		&user.Email,
		&user.Status,
		&user.CreatedAt,
	); err != nil {
		return nil, err
	}

	return &user, nil
}

func (s SqliteAppStore) DeleteUserByUsername(username string) error {
	query := `delete from users_ where username_ = $1`

	res, err := s.db.Exec(query, username)
	if err != nil {
		return err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("no user deleted")
	}

	return nil
}

func (s SqliteAppStore) UpdateUser(user models.User) (*models.User, error) {
	query := `
		update users_ set email_ = $2, set status_ = $3
		where username_ = $1 
		returning id_, username_, email_, status_, created_at_
	`

	row := s.db.QueryRow(query, user.Username, user.Email, user.Status)

	var updatedUser models.User

	if err := row.Scan(
		&updatedUser.Id,
		&updatedUser.Username,
		&updatedUser.Email,
		&updatedUser.Status,
		&updatedUser.CreatedAt,
	); err != nil {
		return nil, err
	}

	return &user, nil
}

func (s SqliteAppStore) InsertKey(userId int, publicKey string) (*models.Key, error) {
	query := `
		insert into keys_ (user_id_, ssh_public_key_)
		values ($userId, $publicKey)
		returning id_, user_id_, ssh_public_key_, created_at_
	`

	row := s.db.QueryRow(
		query,
		sql.Named("userId", userId),
		sql.Named("publicKey", publicKey),
	)

	var insertedKey models.Key

	if err := row.Scan(
		&insertedKey.Id,
		&insertedKey.UserId,
		&insertedKey.PublicKey,
		&insertedKey.CreatedAt,
	); err != nil {
		return nil, err
	}

	return &insertedKey, nil
}
