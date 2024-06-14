package stores

import (
	"database/sql"
	"fmt"
)

type ProjectStore interface {
	Add(name string) error
	Remove(name string) error
	Rename(originalName, newName string) error
	List() ([]string, error)
}

type SqliteProjectStore struct {
	db *sql.DB
}

func NewSqliteProjectStore(db *sql.DB) SqliteProjectStore {
	return SqliteProjectStore{db}
}

func (s SqliteProjectStore) Add(name string) error {
	query := `
		insert into projects_ (name_) values ($name)
	`

	if _, err := s.db.Exec(query, sql.Named("name", name)); err != nil {
		return err
	}

	return nil
}

func (s SqliteProjectStore) Remove(name string) error {
	query := `
		delete from projects_ where name_ = $name
	`

	res, err := s.db.Exec(query, sql.Named("name", name))
	if err != nil {
		return err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("nothing deleted")
	}

	return nil
}

func (s SqliteProjectStore) Rename(originalName, newName string) error {
	query := `
		update projects_ set name_ = $newName where name_ = $originalName
	`

	if _, err := s.db.Exec(
		query,
		sql.Named("originalName", originalName),
		sql.Named("newName", newName),
	); err != nil {
		return err
	}

	return nil
}

func (s SqliteProjectStore) List() ([]string, error) {
	query := `
		select name_ from projects_
	`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}

	var projects []string

	for rows.Next() {
		var project string

		if err := rows.Scan(&project); err != nil {
			return projects, err
		}

		projects = append(projects, project)
	}

	return projects, nil
}
