package internal

import (
	"errors"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
)

func TestVariablesStore(t *testing.T) {
	scenarios := map[string]func(t *testing.T, mock sqlmock.Sqlmock, store VariableStore){
		"set new variable in store (success)":          testVariableStoreSetNewSuccess,
		"set new variable in store (error - database)": testVariableStoreSetNewDatabaseErrror,
	}

	for scenario, fn := range scenarios {
		t.Run(scenario, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatal("failed to create mock database")
			}

			store := NewVariableSqliteStore(db)

			fn(t, mock, store)
		})
	}
}

func testVariableStoreSetNewSuccess(t *testing.T, mock sqlmock.Sqlmock, store VariableStore) {
	query := `insert into variables_ (key_, value_, secret_, project_name_, environment_name_) values ($1, $2, $3, $4, $5)`

	mockResult := sqlmock.NewResult(23, 1)

	mock.ExpectExec(regexp.QuoteMeta(query)).WithArgs(
		"KEY",
		"NAME",
		false,
		"project name",
		"dev",
	).WillReturnResult(mockResult)

	err := store.Set(Variable{
		Key:             "KEY",
		Value:           "NAME",
		Secret:          false,
		ProjectName:     "project name",
		EnvironmentName: "dev",
	})

	require.NoError(t, err, "should not return error")
	require.NoError(t, mock.ExpectationsWereMet(), "all expectations should be met")
}

func testVariableStoreSetNewDatabaseErrror(t *testing.T, mock sqlmock.Sqlmock, store VariableStore) {
	query := `insert into variables_ (key_, value_, secret_, project_name_, environment_name_) values ($1, $2, $3, $4, $5)`

	mock.ExpectExec(regexp.QuoteMeta(query)).WithArgs(
		"KEY",
		"NAME",
		false,
		"project name",
		"dev",
	).WillReturnError(errors.New("database_error"))

	err := store.Set(Variable{
		Key:             "KEY",
		Value:           "NAME",
		Secret:          false,
		ProjectName:     "project name",
		EnvironmentName: "dev",
	})

	require.EqualError(t, err, "database_error", "should return database error")

	require.NoError(t, mock.ExpectationsWereMet(), "should meet expectations")
}
