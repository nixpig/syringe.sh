package stores

import (
	"database/sql"
	"fmt"
)

type Secret struct {
	Id          int
	Project     string
	Environment string
	Key         string
	Value       string
}

type UserStore interface {
	CreateTables() error
	InsertSecret(project, environment, key, value string) error
	GetSecret(project, environment, key string) (*Secret, error)
}

type SqliteUserStore struct {
	db *sql.DB
}

func NewSqliteEnvStore(db *sql.DB) SqliteUserStore {
	return SqliteUserStore{db}
}

func (s SqliteUserStore) CreateTables() error {
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
		return fmt.Errorf(fmt.Sprintf("failed to exec: %s", err))
	}

	return nil
}

func (s SqliteUserStore) InsertSecret(project, environment, key, value string) error {
	query := `
		insert into envs_ 
		(project_, environment_, key_, value_) 
		values ($project, $environment, $key, $value)
	`

	if _, err := s.db.Exec(
		query,
		sql.Named("project", project),
		sql.Named("environment", environment),
		sql.Named("key", key),
		sql.Named("value", value),
	); err != nil {
		return err
	}

	return nil
}

func (s SqliteUserStore) GetSecret(project, environment, key string) (*Secret, error) {
	query := `
		select id_, project_, environment_, key_, value_
		from envs_
		where project_ = $project
		and environment_ = $environment
		and key_ = $key
	`

	row := s.db.QueryRow(
		query,
		sql.Named("project", project),
		sql.Named("environment", environment),
		sql.Named("key", key),
	)

	var secret Secret

	if err := row.Scan(
		&secret.Id,
		&secret.Project,
		&secret.Environment,
		&secret.Key,
		&secret.Value,
	); err != nil {
		return nil, err
	}

	return &secret, nil
}
