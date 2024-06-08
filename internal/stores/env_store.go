package stores

import (
	"database/sql"
	"errors"
	"fmt"
)

type EnvStore interface {
	CreateTables() error
}

type SqliteEnvStore struct {
	db *sql.DB
}

func NewSqliteEnvStore(db *sql.DB) SqliteEnvStore {
	return SqliteEnvStore{db}
}

func (s SqliteEnvStore) CreateTables() error {
	query := `
		create table if not exists envs_ (
			id_ integer primary key autoincrement,
			key_ text not null,
			value_ text not null,
			project_ varchar(256) not null,
			environment_ varchar(256) not null
		)
	`

	if _, err := s.db.Exec(query); err != nil {
		return errors.New(fmt.Sprintf("failed to exec: %s", err))
	}

	return nil
}
