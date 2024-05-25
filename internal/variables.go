package internal

import "database/sql"

type Variable struct {
	Id              int    `validate:"omitempty"`
	Key             string `validate:"required"`
	Value           string `validate:"required"`
	Secret          bool   `validate:"required"`
	ProjectName     string `validate:"required"`
	EnvironmentName string `validate:"required"`
}

type VariableStore interface {
	Set(variable Variable) error
	// Get() (Variable, error)
	// GetAll() ([]Variable, error)
	// Delete(variable Variable) error
}

type VariableSqliteStore struct {
	db *sql.DB
}

func NewVariableSqliteStore(db *sql.DB) VariableSqliteStore {
	return VariableSqliteStore{db}
}

func (v VariableSqliteStore) Set(variable Variable) error {
	query := `insert into variables_ (key_, value_, secret_, project_name_, environment_name_) values ($1, $2, $3, $4, $5)`

	_, err := v.db.Exec(
		query,
		variable.Key,
		variable.Value,
		variable.Secret,
		variable.ProjectName,
		variable.EnvironmentName,
	)
	if err != nil {
		return err
	}

	return nil
}
