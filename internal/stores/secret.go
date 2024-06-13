package stores

import (
	"context"
	"database/sql"
)

type Secret struct {
	Id          int
	Project     string
	Environment string
	Key         string
	Value       string
}

type SecretStore interface {
	CreateTables() error
	InsertSecret(project, environment, key, value string) error
	GetSecret(project, environment, key string) (*Secret, error)
}

type SqliteSecretStore struct {
	db *sql.DB
}

func NewSqliteEnvStore(db *sql.DB) SqliteSecretStore {
	return SqliteSecretStore{db}
}

func (s SqliteSecretStore) CreateTables() error {
	projectsQuery := `
		create table if not exists projects_ (
			id_ integer primary key autoincrement,
			name_ varchar(256) unique not null
		)
	`

	environmentsQuery := `
		create table if not exists environments_ (
			id_ integer primary key autoincrement,
			name_ varchar(256) not null,
			project_id_ integer not null,

			foreign key (project_id_) references projects_(id_)
		)
	`
	secretsQuery := `
		create table if not exists secrets_ (
			id_ integer primary key autoincrement,
			key_ text not null,
			value_ text not null,
			environment_id_ integer not null,

			foreign key (environment_id_) references environments_(id_)
		)
	`

	ctx := context.Background()

	trx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	trx.Exec(projectsQuery)
	trx.Exec(environmentsQuery)
	trx.Exec(secretsQuery)

	if err := trx.Commit(); err != nil {
		return err
	}

	return nil
}

func (s SqliteSecretStore) InsertSecret(project, environment, key, value string) error {
	query := `
		insert into secrets_ 
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

func (s SqliteSecretStore) GetSecret(project, environment, key string) (*Secret, error) {
	query := `
		select id_, project_, environment_, key_, value_
		from secrets_
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
