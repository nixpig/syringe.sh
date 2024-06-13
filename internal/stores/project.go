package stores

import "database/sql"

type Project struct {
	ID   int
	Name string
}

type ProjectStore interface {
	InsertProject(name string) error
}

type SqliteProjectStore struct {
	db *sql.DB
}

func NewSqliteProjectStore(db *sql.DB) SqliteProjectStore {
	return SqliteProjectStore{db}
}

func (s SqliteProjectStore) InsertProject(name string) error {
	query := `
		insert into projects_ (name_) values ($name)
	`

	if _, err := s.db.Exec(query, sql.Named("name", name)); err != nil {
		return err
	}

	return nil
}
