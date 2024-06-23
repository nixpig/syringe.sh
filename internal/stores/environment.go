package stores

import (
	"database/sql"

	"github.com/nixpig/syringe.sh/server/pkg"
)

type Environment struct {
	ID      int
	Name    string
	Project string
}

type EnvironmentStore interface {
	Add(name, projectName string) error
	Remove(name, projectName string) error
	Rename(originalName, newName, projectName string) error
	List(projectName string) (*[]Environment, error)
}

type SqliteEnvironmentStore struct {
	db *sql.DB
}

func NewSqliteEnvironmentStore(db *sql.DB) EnvironmentStore {
	return SqliteEnvironmentStore{db}
}

func (s SqliteEnvironmentStore) Add(name, projectName string) error {
	query := `
		insert into environments_ (name_, project_id_) values (
			$name,
			(select id_ from projects_ where name_ = $projectName)
		)
	`

	if _, err := s.db.Exec(
		query,
		sql.Named("name", name),
		sql.Named("projectName", projectName),
	); err != nil {
		return pkg.ErrDatabaseExec(err)
	}

	return nil
}

func (s SqliteEnvironmentStore) Remove(name, projectName string) error {
	query := `
		delete from environments_ 
		where id_ in (
			select e.id_ from environments_ e
			inner join
			projects_ p
			on e.project_id_ = p.id_
			where p.name_ = $projectName
			and e.name_ = $name
		)
	`

	res, err := s.db.Exec(
		query,
		sql.Named("name", name),
		sql.Named("projectName", projectName),
	)
	if err != nil {
		return pkg.ErrDatabaseExec(err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return pkg.ErrEnvironmentNotFound
	}

	return nil
}

func (s SqliteEnvironmentStore) Rename(originalName, newName, projectName string) error {
	query := `
		update environments_ set name_ = $newName
		where name_ = $originalName 
		and id_ in (
			select e.id_ from environments_ e
			inner join
			projects_ p
			on e.project_id_ = p.id_
			where p.name_ = $projectName
			and e.name_ = $originalName
		)
	`

	res, err := s.db.Exec(
		query,
		sql.Named("originalName", originalName),
		sql.Named("newName", newName),
		sql.Named("projectName", projectName),
	)

	if err != nil {
		return pkg.ErrDatabaseExec(err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return pkg.ErrEnvironmentNotFound
	}

	return nil
}

func (s SqliteEnvironmentStore) List(projectName string) (*[]Environment, error) {
	query := `
		select e.id_, e.name_, p.name_ from environments_ e
		inner join projects_ p
		on e.project_id_ = p.id_ 
		where p.name_ = $projectName
	`

	rows, err := s.db.Query(
		query,
		sql.Named("projectName", projectName),
	)
	if err == sql.ErrNoRows {
		return nil, pkg.ErrNoEnvironmentsFound
	}
	if err != nil {
		return nil, pkg.ErrDatabaseQuery(err)
	}

	var environments []Environment

	for rows.Next() {
		var environment Environment

		if err := rows.Scan(
			&environment.ID,
			&environment.Name,
			&environment.Project,
		); err != nil {
			return nil, err
		}

		environments = append(environments, environment)
	}

	return &environments, nil
}
