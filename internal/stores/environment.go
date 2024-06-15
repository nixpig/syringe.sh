package stores

import "database/sql"

type Environment struct {
	ID        int
	Name      string
	ProjectID int
}

type EnvironmentStore interface {
	Add(name string, projectName string) error
	// delete
	// get all (for project id/name)
	// rename
}

type SqliteEnvironmentStore struct {
	db *sql.DB
}

func NewSqliteEnvironmentStore(db *sql.DB) EnvironmentStore {
	return SqliteEnvironmentStore{db}
}

func (s SqliteEnvironmentStore) Add(name string, projectName string) error {
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
