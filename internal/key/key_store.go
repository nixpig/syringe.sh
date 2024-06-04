package key

import "database/sql"

type KeyStore interface {
	Insert(userId int, publicKey string) (*Key, error)
}

type SqliteKeyStore struct {
	db *sql.DB
}

func NewSqliteKeyStore(db *sql.DB) SqliteKeyStore {
	return SqliteKeyStore{db}
}

func (k SqliteKeyStore) Insert(userId int, publicKey string) (*Key, error) {
	query := `
		insert into keys_ (user_id_, ssh_public_key_)
		values ($userId, $publicKey)
		returning id_, user_id_, ssh_public_key_
	`

	row := k.db.QueryRow(
		query,
		sql.Named("userId", userId),
		sql.Named("publicKey", publicKey),
	)

	var insertedKey Key

	if err := row.Scan(
		&insertedKey.Id,
		&insertedKey.UserId,
		&insertedKey.PublicKey,
	); err != nil {
		return nil, err
	}

	return &insertedKey, nil
}
