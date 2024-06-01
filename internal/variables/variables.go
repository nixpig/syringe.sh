package internal

import (
	"database/sql"
)

const (
	MSG_SET_VAR_SUCCESS = "Sucessfully set: \nProject: %s\nEnvironment: %s\nKey: %s\nValue: %s\nSecret: %v\n"
)

type Variable struct {
	Key             string `validate:"required"`
	Value           string `validate:"required"`
	Secret          *bool  `validate:"required"`
	ProjectName     string `validate:"required"`
	EnvironmentName string `validate:"required"`
}

type VariableStore interface {
	Set(variable Variable) error
	Get(projectName, environmentName, key string) (string, error)
	Delete(projectName, environmentName, key string) error
	GetAll(projectName, environmentName string) ([]Variable, error)
}

type VariableStoreSqlite struct {
	db *sql.DB
}

func NewVariableStoreSqlite(db *sql.DB) VariableStoreSqlite {
	return VariableStoreSqlite{db}
}

func (v VariableStoreSqlite) Set(variable Variable) error {
	query := `insert into 
		variables_ (key_, value_, secret_, project_name_, environment_name_) 
		values ($1, $2, $3, $4, $5) 
		on conflict (key_, project_name_, environment_name_) 
		do update set value_ = $2, secret_ = $3
	`

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

func (v VariableStoreSqlite) Get(projectName, environmentName, key string) (string, error) {
	query := `select value_ from variables_ where project_name_ = $1 and environment_name_ = $2 and key_ = $3`

	row := v.db.QueryRow(query, projectName, environmentName, key)

	var variable string

	if err := row.Scan(&variable); err != nil {
		if err == sql.ErrNoRows {
			return "", nil
		}

		return "", err
	}

	return variable, nil
}

func (v VariableStoreSqlite) Delete(projectName, environmentName, key string) error {
	query := `delete from variables_ where project_name_ = $1 and environment_name_ = $2 and key_ = $3`

	_, err := v.db.Exec(query, projectName, environmentName, key)
	if err != nil {
		return err
	}

	return nil
}

func (v VariableStoreSqlite) GetAll(projectName, environmentName string) ([]Variable, error) {
	query := `select key_, value_, secret_ from variables_ where project_name_ = $1 and environment_name_ = $2`

	rows, err := v.db.Query(query, projectName, environmentName)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var variables []Variable

	for rows.Next() {
		var variable Variable

		if err := rows.Scan(&variable.Key, &variable.Value, &variable.Secret); err != nil {
			return nil, err
		}

		variables = append(variables, variable)
	}

	return variables, nil
}
