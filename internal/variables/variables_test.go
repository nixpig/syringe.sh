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

		"get variable from store (success)":          testVariableStoreGetVariableSuccess,
		"get variable from store (success - empty)":  testVariableStoreGetVariableSuccessEmpty,
		"get variable from store (error - row scan)": testVariableStoreGetVariableRowScanError,

		"delete variable from store (success)":          testVariableStoreDeleteVariableSuccess,
		"delete variable from store (error - database)": testVariableStoreDeleteVariableDatabaseError,

		"get all variables for project and environment (success)":       testVariableStoreGetAllSuccess,
		"get all variables for project and environment (error - query)": testVariableStoreGetAllQueryError,
		"get all variables for project and environment (error - row)":   testVariableStoreGetAllRowError,
	}

	for scenario, fn := range scenarios {
		t.Run(scenario, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatal("failed to create mock database")
			}

			store := NewVariableStoreSqlite(db)

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

	secret := false

	err := store.Set(Variable{
		Key:             "KEY",
		Value:           "NAME",
		Secret:          &secret,
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

	secret := false

	err := store.Set(Variable{
		Key:             "KEY",
		Value:           "NAME",
		Secret:          &secret,
		ProjectName:     "project name",
		EnvironmentName: "dev",
	})

	require.EqualError(t, err, "database_error", "should return database error")
	require.NoError(t, mock.ExpectationsWereMet(), "should meet expectations")
}

func testVariableStoreGetVariableSuccess(t *testing.T, mock sqlmock.Sqlmock, store VariableStore) {
	query := `select value_ from variables_ where project_name_ = $1 and environment_name_ = $2 and key_ = $3`

	mockRow := mock.NewRows([]string{"value_"}).AddRow("var_value")

	mock.ExpectQuery(regexp.QuoteMeta(query)).WithArgs(
		"project_name",
		"environment_name",
		"VAR_KEY",
	).WillReturnRows(mockRow)

	variable, err := store.Get("project_name", "environment_name", "VAR_KEY")

	require.NoError(t, err, "should not return an error")
	require.Equal(t, "var_value", variable, "should return variable value")
	require.NoError(t, mock.ExpectationsWereMet(), "should meet expectations")
}

func testVariableStoreGetVariableSuccessEmpty(t *testing.T, mock sqlmock.Sqlmock, store VariableStore) {
	query := `select value_ from variables_ where project_name_ = $1 and environment_name_ = $2 and key_ = $3`

	mockRow := mock.NewRows([]string{"value_"})

	mock.ExpectQuery(regexp.QuoteMeta(query)).WithArgs(
		"project_name",
		"environment_name",
		"VAR_KEY",
	).WillReturnRows(mockRow)

	variable, err := store.Get("project_name", "environment_name", "VAR_KEY")

	require.NoError(t, err, "should not return an error")
	require.Equal(t, "", variable, "should return empty string")
	require.NoError(t, mock.ExpectationsWereMet(), "should meet expectations")
}

func testVariableStoreGetVariableRowScanError(t *testing.T, mock sqlmock.Sqlmock, store VariableStore) {
	query := `select value_ from variables_ where project_name_ = $1 and environment_name_ = $2 and key_ = $3`

	mockRow := mock.
		NewRows([]string{"value_"}).
		RowError(0, errors.New("row_error")).
		AddRow("var_value")

	mock.ExpectQuery(regexp.QuoteMeta(query)).WillReturnRows(mockRow)

	variable, err := store.Get("project_name", "environment_name", "VAR_KEY")

	require.Empty(t, variable, "should return empty string")
	require.EqualError(t, err, "row_error", "should return row error")
	require.NoError(t, mock.ExpectationsWereMet(), "should meet expectations")
}

func testVariableStoreDeleteVariableSuccess(t *testing.T, mock sqlmock.Sqlmock, store VariableStore) {
	query := `delete from variables_ where project_name_ = $1 and environment_name_ = $2 and key_ = $3`

	mock.ExpectExec(regexp.QuoteMeta(query)).WithArgs(
		"project_name",
		"environment_name",
		"VAR_KEY",
	).WillReturnResult(sqlmock.NewResult(0, 0))

	err := store.Delete("project_name", "environment_name", "VAR_KEY")

	require.NoError(t, err, "should not return error")
	require.NoError(t, mock.ExpectationsWereMet(), "should meet expectations")
}

func testVariableStoreDeleteVariableDatabaseError(t *testing.T, mock sqlmock.Sqlmock, store VariableStore) {
	query := `delete from variables_ where project_name_ = $1 and environment_name_ = $2 and key_ = $3`

	mock.ExpectExec(regexp.QuoteMeta(query)).WithArgs(
		"project_name",
		"environment_name",
		"VAR_KEY",
	).WillReturnError(errors.New("database_error"))

	err := store.Delete("project_name", "environment_name", "VAR_KEY")

	require.EqualError(t, err, "database_error", "should return database error")
	require.NoError(t, mock.ExpectationsWereMet(), "should meet expectations")
}

func testVariableStoreGetAllSuccess(t *testing.T, mock sqlmock.Sqlmock, store VariableStore) {
	query := `select key_, value_, secret_ from variables_ where project_name_ = $1 and environment_name_ = $2`

	mockRows := mock.
		NewRows([]string{"key_", "value_", "secret_"}).
		AddRow("KEY_1", "value_1", true).
		AddRow("KEY_2", "value_2", false)

	mock.
		ExpectQuery(regexp.QuoteMeta(query)).
		WithArgs("project_name", "environment_name").
		WillReturnRows(mockRows)

	variables, err := store.GetAll("project_name", "environment_name")

	require.NoError(t, err, "should not return error")

	truePtr := true
	falsePtr := false

	require.Equal(t, []Variable{
		{
			Key:    "KEY_1",
			Value:  "value_1",
			Secret: &truePtr,
		},
		{
			Key:    "KEY_2",
			Value:  "value_2",
			Secret: &falsePtr,
		},
	}, variables, "should return variables for project and environment")

	require.NoError(t, mock.ExpectationsWereMet(), "all expectations should be met")
}

func testVariableStoreGetAllQueryError(t *testing.T, mock sqlmock.Sqlmock, store VariableStore) {
	query := `select key_, value_, secret_ from variables_ where project_name_ = $1 and environment_name_ = $2`

	mock.
		ExpectQuery(regexp.QuoteMeta(query)).
		WithArgs("project_name", "environment_name").
		WillReturnError(errors.New("query_error"))

	variables, err := store.GetAll("project_name", "environment_name")

	require.Nil(t, variables, "should not return variables")
	require.EqualError(t, err, "query_error", "should return error from query")
	require.NoError(t, mock.ExpectationsWereMet(), "all expectations should be met")
}

func testVariableStoreGetAllRowError(t *testing.T, mock sqlmock.Sqlmock, store VariableStore) {
	query := `select key_, value_, secret_ from variables_ where project_name_ = $1 and environment_name_ = $2`

	mockRows := mock.
		NewRows([]string{"key_", "value_", "secret_"}).
		AddRow("VAR_KEY_1", "var_value_1", false).
		AddRow(nil, "var_value_2", true).
		RowError(2, errors.New("row_error"))

	mock.
		ExpectQuery(regexp.QuoteMeta(query)).
		WithArgs("project_name", "environment_name").
		WillReturnRows(mockRows)

	variables, err := store.GetAll("project_name", "environment_name")

	require.Nil(t, variables, "should not return variables")
	require.Error(t, err, "should return row error")
	require.NoError(t, mock.ExpectationsWereMet(), "all expectations should be met")
}
