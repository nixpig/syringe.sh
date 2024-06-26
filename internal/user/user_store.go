package user

import "database/sql"

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

type UserStore interface {
	InsertUser(username, email, status string) (*User, error)
	InsertKey(userID int, publicKey string) (*Key, error)
}

type SqliteUserStore struct {
	db *sql.DB
}

func NewSqliteUserStore(db *sql.DB) SqliteUserStore {
	return SqliteUserStore{db}
}

func (s SqliteUserStore) InsertUser(username, email, status string) (*User, error) {
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

func (s SqliteUserStore) InsertKey(userID int, publicKey string) (*Key, error) {
	query := `
	insert into keys_ (user_id_, ssh_public_key_)
	values ($userID, $publicKey)
	returning id_, user_id_, ssh_public_key_, created_at_
	`

	row := s.db.QueryRow(
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
