package secret

import (
	"database/sql"
	"errors"
	"fmt"
)

type Secret struct {
	ID          int
	Key         string
	Value       string
	Project     string
	Environment string
}

type SecretStore interface {
	Set(project, environment, key, value string) error
	Get(project, environment, key string) (*Secret, error)
	List(project, environment string) (*[]Secret, error)
	Remove(project, environment, key string) error
}

type SqliteSecretStore struct {
	db *sql.DB
}

func NewSqliteSecretStore(db *sql.DB) SqliteSecretStore {
	return SqliteSecretStore{db}
}

func (s SqliteSecretStore) Set(project, environment, key, value string) error {
	query := `
		insert into secrets_ 
		(key_, value_, environment_id_) 
		values (
			$key,
			$value,
			(
				select e.id_ from 
					environments_ e
					inner join 
					projects_ p 
					on e.project_id_ = p.id_ 
					where p.name_ = $project 
					and e.name_ = $environment
			)
		) 
		on conflict(key_)
		do update set value_ = $value
	`

	if _, err := s.db.Exec(
		query,
		sql.Named("project", project),
		sql.Named("environment", environment),
		sql.Named("key", key),
		sql.Named("value", value),
	); err != nil {
		return fmt.Errorf("secret set database error: %w", err)
	}

	return nil
}

func (s SqliteSecretStore) Get(project, environment, key string) (*Secret, error) {
	query := `
		select s.id_, s.key_, s.value_, p.name_, e.name_
		from secrets_ s
		inner join
		environments_ e
		on s.environment_id_ = e.id_
		inner join
		projects_ p
		on p.id_ = e.project_id_
		where p.name_ = $project
		and e.name_ = $environment
		and s.key_ = $key
	`

	row := s.db.QueryRow(
		query,
		sql.Named("project", project),
		sql.Named("environment", environment),
		sql.Named("key", key),
	)

	var secret Secret

	if err := row.Scan(
		&secret.ID,
		&secret.Key,
		&secret.Value,
		&secret.Project,
		&secret.Environment,
	); err != nil {
		return nil, fmt.Errorf("secret get database error: %w", err)
	}

	return &secret, nil
}

func (s SqliteSecretStore) List(project, environment string) (*[]Secret, error) {
	query := `
		select s.id_, s.key_, s.value_, p.name_, e.name_
		from secrets_ s
		inner join
		environments_ e
		on s.environment_id_ = e.id_
		inner join
		projects_ p
		on p.id_ = e.project_id_
		where p.name_ = $project
		and e.name_ = $environment
	`

	rows, err := s.db.Query(
		query,
		sql.Named("project", project),
		sql.Named("environment", environment),
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New("no secrets found")
	}
	if err != nil {
		return nil, fmt.Errorf("secret list database error: %w", err)
	}

	var secrets []Secret

	for rows.Next() {
		var secret Secret

		if err := rows.Scan(
			&secret.ID,
			&secret.Key,
			&secret.Value,
			&secret.Project,
			&secret.Environment,
		); err != nil {
			return nil, err
		}

		secrets = append(secrets, secret)
	}

	return &secrets, nil
}

func (s SqliteSecretStore) Remove(project, environment, key string) error {
	query := `
		delete from secrets_ 
		where id_ in (
			select s.id_ from secrets_ s
			inner join
			environments_ e
			on s.environment_id_ = e.id_
			inner join
			projects_ p
			on e.project_id_ = p.id_
			where p.name_ = $projectName
			and e.name_ = $environmentName
			and s.key_ = $key
		)
	`

	res, err := s.db.Exec(
		query,
		sql.Named("projectName", project),
		sql.Named("environmentName", environment),
		sql.Named("key", key),
	)
	if err != nil {
		return fmt.Errorf("secret remove database error: %w", err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return errors.New(fmt.Sprintf("secret '%s' not found", key))
	}

	return nil
}
