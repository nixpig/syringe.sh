package project

import (
	"database/sql"
	"errors"
	"fmt"
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
		return fmt.Errorf("project add database error: %w", err)
	}

	return nil
}

func (s SqliteProjectStore) Remove(name string) error {
	query := `
		delete from projects_ where name_ = $name
	`

	res, err := s.db.Exec(query, sql.Named("name", name))
	if err != nil {
		return fmt.Errorf("project remove database error: %w", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("project '%s' not found", name)
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
		return fmt.Errorf("project rename database error: %w", err)
	}

	return nil
}

func (s SqliteProjectStore) List() (*[]Project, error) {
	query := `
		select id_, name_ from projects_
	`

	rows, err := s.db.Query(query)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New("no projects found")
	}
	if err != nil {
		return nil, fmt.Errorf("project list database error: %w", err)
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
