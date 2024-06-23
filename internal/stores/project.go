package stores

import (
	"database/sql"
	"fmt"

	"github.com/nixpig/syringe.sh/server/pkg"
)

type Project struct {
	ID   int
	Name string
}

type ProjectStore interface {
	Add(name string) error
	Remove(name string) error
	Rename(originalName, newName string) error
	List() (*[]Project, error)
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
		return fmt.Errorf("project not found")
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

func (s SqliteProjectStore) List() (*[]Project, error) {
	query := `
		select id_, name_ from projects_
	`

	rows, err := s.db.Query(query)
	if err == sql.ErrNoRows {
		return nil, pkg.ErrNoProjectsFound
	}
	if err != nil {
		return nil, err
	}

	var projects []Project

	for rows.Next() {
		var project Project

		if err := rows.Scan(&project.ID, &project.Name); err != nil {
			return nil, err
		}

		projects = append(projects, project)
	}

	return &projects, nil
}
