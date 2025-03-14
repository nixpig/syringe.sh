package stores

import (
	"database/sql"
	"fmt"
	"sync"
)

type SystemStore struct {
	db *sql.DB
	mu sync.Mutex
}

func NewSystemStore(db *sql.DB) *SystemStore {
	return &SystemStore{
		db: db,
		mu: sync.Mutex{},
	}
}

func (s *SystemStore) GetUser(username string) (*User, error) {
	query := `select u.id_, u.username_, u.email_, u.verified_, k.public_key_sha1_
		from users_ u inner join public_keys_ k on u.id_ = k.user_id_ where u.username_ = $username`

	row := s.db.QueryRow(
		query,
		sql.Named("username", username),
	)

	var user User

	if err := row.Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.Verified,
		&user.PublicKeySHA1,
	); err != nil {
		return nil, fmt.Errorf("scan user: %w", err)
	}

	return &user, nil
}

func (s *SystemStore) CreateUser(user *User) (int, error) {
	userQuery := `insert into users_ (username_, email_, verified_)
		values ($username, $email, $verified) returning id_`

	tx, err := s.db.Begin()
	if err != nil {
		return 0, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	row := tx.QueryRow(
		userQuery,
		sql.Named("username", user.Username),
		sql.Named("email", user.Email),
		sql.Named("verified", user.Verified),
		sql.Named("publicKeySHA1", user.PublicKeySHA1),
	)

	var userID int

	if err := row.Scan(&userID); err != nil {
		return 0, fmt.Errorf("scan user id: %w", err)
	}

	keyQuery := `insert into public_keys_ (public_key_sha1_, user_id_)
		values ($publicKeySHA1, $userID)`
	if _, err := tx.Exec(
		keyQuery,
		sql.Named("publicKeySHA1", user.PublicKeySHA1),
		sql.Named("userID", userID),
	); err != nil {
		return 0, fmt.Errorf("create public key: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("commit create user transaction: %w", err)
	}

	return userID, nil
}
