package stores

import (
	"database/sql"
	"fmt"
)

type Environment struct {
	ID        int
	Name      string
	ProjectID int
}

type EnvironmentStore interface {
	Add(name, projectName string) error
	Remove(name, projectName string) error
	// get all (for project id/name)
	// rename
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
		return err
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
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		fmt.Println("done fucked up!")
	}

	return nil
}
