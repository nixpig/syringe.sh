package auth

import (
	"database/sql"
)

type UserKey struct {
	ID        int
	PublicKey string
	UserID    int
	CreatedAt string
}

type AuthStore interface {
	GetUserPublicKeys(username string) (*[]UserKey, error)
}

type SqliteAuthStore struct {
	appDB *sql.DB
}

func NewSqliteAuthStore(appDB *sql.DB) SqliteAuthStore {
	return SqliteAuthStore{appDB}
}

func (s SqliteAuthStore) GetUserPublicKeys(username string) (*[]UserKey, error) {
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

	var keys []UserKey

	for rows.Next() {
		var key UserKey

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
