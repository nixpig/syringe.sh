package stores

import (
	"database/sql"
	"errors"
	"fmt"
)

type EnvStore interface {
	CreateTables() error
	InsertSecret(project, environment, key, value string) error
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

func (s SqliteEnvStore) InsertSecret(project, environment, key, value string) error {
	query := `
		insert into envs_ 
		(project_, environment_, key_, value_) 
		values ($project, $environment, $key, $value)
	`

	fmt.Println("stats in store: ", s.db.Stats())

	if err := s.db.Ping(); err != nil {
		fmt.Println("failed to ping database in store")
		return err
	}

	fmt.Println("execing...")
	fmt.Println("project: ", project)
	fmt.Println("environment: ", environment)
	fmt.Println("key: ", key)
	fmt.Println("value: ", value)
	if _, err := s.db.Query(
		query,
		sql.Named("project", project),
		sql.Named("environment", environment),
		sql.Named("key", key),
		sql.Named("value", value),
	); err != nil {
		fmt.Println("this is the exec error: ", err)
		return err
	}

	fmt.Println("executed!")

	return nil
}
