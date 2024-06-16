package stores

import (
	"database/sql"
	"fmt"
)

type User struct {
	ID        int
	Username  string
	Email     string
	CreatedAt string
	Status    string
}

type Key struct {
	ID        int
	PublicKey string
	UserID    int
	CreatedAt string
}

type AppStore interface {
	InsertUser(username, email, status string) (*User, error)
	GetUserByUsername(username string) (*User, error)
	DeleteUserByUsername(username string) error
	GetUserPublicKeys(username string) (*[]Key, error)
	InsertKey(userID int, publicKey string) (*Key, error)
}

type SqliteAppStore struct {
	appDB *sql.DB
}

func NewSqliteAppStore(appDB *sql.DB) SqliteAppStore {
	return SqliteAppStore{appDB}
}

func (s SqliteAppStore) InsertUser(username, email, status string) (*User, error) {
	query := `
		insert into users_ (username_, email_, status_) 
		values ($username, $email, $status) 
		returning id_, username_, email_, status_, created_at_
	`

	row := s.appDB.QueryRow(
		query,
		sql.Named("username", username),
		sql.Named("email", email),
		sql.Named("status", status),
	)

	var insertedUser User

	if err := row.Scan(
		&insertedUser.ID,
		&insertedUser.Username,
		&insertedUser.Email,
		&insertedUser.Status,
		&insertedUser.CreatedAt,
	); err != nil {
		return nil, err
	}

	return &insertedUser, nil
}

func (s SqliteAppStore) GetUserByUsername(username string) (*User, error) {
	query := `
		select id_, username_, email_, status_, created_at_ 
		from users_ 
		where username_ = $1
	`

	row := s.appDB.QueryRow(query, username)

	var user User

	if err := row.Scan(
		&user.ID,
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

	res, err := s.appDB.Exec(query, username)
	if err != nil {
		return err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("no user deleted")
	}

	return nil
}

func (s SqliteAppStore) InsertKey(userID int, publicKey string) (*Key, error) {
	query := `
		insert into keys_ (user_id_, ssh_public_key_)
		values ($userID, $publicKey)
		returning id_, user_id_, ssh_public_key_, created_at_
	`

	row := s.appDB.QueryRow(
		query,
		sql.Named("userID", userID),
		sql.Named("publicKey", publicKey),
	)

	var insertedKey Key

	if err := row.Scan(
		&insertedKey.ID,
		&insertedKey.UserID,
		&insertedKey.PublicKey,
		&insertedKey.CreatedAt,
	); err != nil {
		return nil, err
	}

	return &insertedKey, nil
}

func (s SqliteAppStore) GetUserPublicKeys(username string) (*[]Key, error) {
	query := `
		select k.id_, k.user_id_, k.ssh_public_key_, k.created_at_
		from keys_ k 
		inner join
		users_ u
		on k.user_id_ = u.id_
		where u.username_ = $username
	`

	rows, err := s.appDB.Query(
		query,
		sql.Named("username", username),
	)
	if err != nil {
		return nil, err
	}

	var keys []Key

	for rows.Next() {
		var key Key

		if err := rows.Scan(
			&key.ID,
			&key.UserID,
			&key.PublicKey,
			&key.CreatedAt,
		); err != nil {
			return nil, err
		}

		keys = append(keys, key)
	}

	return &keys, nil
}
