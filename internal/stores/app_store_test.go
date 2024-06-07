package stores

import (
	"database/sql"
	"errors"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/nixpig/syringe.sh/server/internal/models"
	"github.com/stretchr/testify/require"
)

func TestSqliteAppStore(t *testing.T) {
	scenarios := map[string]func(t *testing.T, mock sqlmock.Sqlmock, store AppStore){
		"test sqlite app store insert user (success)":    testSqliteAppStoreInsertUserSuccess,
		"test sqlite app store insert user (scan error)": testSqliteAppStoreInsertUserScanError,

		"test sqlite app store get user by username (success)":           testSqliteAppStoreGetUserByUsernameSuccess,
		"test sqlite app store get user by username (success - no user)": testSqliteAppStoreGetUserByUsernameSuccessNoUser,
		"test sqlite app store get user by username (scan error)":        testSqliteAppStoreGetUserByUsernameScanError,

		"test sqlite app store delete user by username (success)":         testSqliteAppStoreDeleteUserByUsernameSuccess,
		"test sqlite app store delete user by username (query error)":     testSqliteAppStoreDeleteUserByUsernameQueryError,
		"test sqlite app store delete user by username (rows error)":      testSqliteAppStoreDeleteUserByUsernameRowsError,
		"test sqlite app store delete user by username (no user deleted)": testSqliteAppStoreDeleteUserByUsernameNoUserError,

		"test sqlite app store update user (success)":    testSqliteAppStoreUpdateUserSuccess,
		"test sqlite app store update user (scan error)": testSqliteAppStoreUpdateUserScanError,

		"test sqlite app store insert key (success)":    testSqliteAppStoreInsertKeySuccess,
		"test sqlite app store insert key (scan error)": testSqliteAppStoreInsertKeyScanError,

		"test sqlite app store insert database (success)":    testSqliteAppStoreInsertDatabaseSuccess,
		"test sqlite app store insert database (scan error)": testSqliteAppStoreInsertDatabaseScanError,
	}

	for scenario, fn := range scenarios {
		t.Run(scenario, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatal("unable to create new mock sqlite database")
			}

			store := NewSqliteAppStore(db)

			fn(t, mock, store)
		})
	}
}

func testSqliteAppStoreInsertUserSuccess(t *testing.T, mock sqlmock.Sqlmock, store AppStore) {
	query := `
		insert into users_ (username_, email_, status_) 
		values ($username, $email, $status) 
		returning id_, username_, email_, status_, created_at_
	`

	createdAt := "2024-06-05 05:29:16"

	mockRow := mock.
		NewRows([]string{"id_", "username_", "email_", "status_", "created_at_"}).
		AddRow(23, "janedoe", "jane@example.org", "active", createdAt)

	mock.
		ExpectQuery(regexp.QuoteMeta(query)).
		WithArgs("janedoe", "jane@example.org", "active").
		WillReturnRows(mockRow)

	insertedUser, err := store.InsertUser("janedoe", "jane@example.org", "active")

	require.NoError(t, err, "should not return error")
	require.Equal(t, &models.User{
		Id:        23,
		Username:  "janedoe",
		Email:     "jane@example.org",
		Status:    "active",
		CreatedAt: createdAt,
	}, insertedUser, "should return inserted user record")

	if expected := mock.ExpectationsWereMet(); expected != nil {
		t.Errorf("database was not called as expected:\n%s", expected)
	}
}

func testSqliteAppStoreInsertUserScanError(t *testing.T, mock sqlmock.Sqlmock, store AppStore) {
	query := `
		insert into users_ (username_, email_, status_) 
		values ($username, $email, $status) 
		returning id_, username_, email_, status_, created_at_
	`

	createdAt := "2024-06-05 05:29:16"

	mockRow := mock.
		NewRows([]string{"id_", "username_", "email_", "status_", "created_at_"}).
		AddRow(23, "janedoe", "jane@example.org", "active", createdAt).
		RowError(0, errors.New("row_error"))

	mock.
		ExpectQuery(regexp.QuoteMeta(query)).
		WithArgs("janedoe", "jane@example.org", "active").
		WillReturnRows(mockRow)

	insertedUser, err := store.InsertUser("janedoe", "jane@example.org", "active")

	require.EqualError(t, err, "row_error", "should return row error from database")
	require.Empty(t, insertedUser, "should return empty inserted user record")

	if expected := mock.ExpectationsWereMet(); expected != nil {
		t.Errorf("database was not called as expected:\n%s", expected)
	}
}

func testSqliteAppStoreGetUserByUsernameSuccess(t *testing.T, mock sqlmock.Sqlmock, store AppStore) {
	query := `
		select id_, username_, email_, status_, created_at_ 
		from users_ 
		where username_ = $1
	`
	createdAt := "2024-06-05 05:29:16"

	mockRow := mock.
		NewRows([]string{"id_", "username_", "email_", "status_", "created_at_"}).
		AddRow(23, "janedoe", "jane@example.org", "active", createdAt)

	mock.
		ExpectQuery(regexp.QuoteMeta(query)).
		WithArgs("janedoe").
		WillReturnRows(mockRow)

	user, err := store.GetUserByUsername("janedoe")

	require.NoError(t, err, "should not return error")
	require.Equal(t, &models.User{
		Id:        23,
		Username:  "janedoe",
		Email:     "jane@example.org",
		Status:    "active",
		CreatedAt: createdAt,
	}, user, "should return user record")

	if expected := mock.ExpectationsWereMet(); expected != nil {
		t.Errorf("database was not queried as expected:\n%s", expected)
	}
}

func testSqliteAppStoreGetUserByUsernameSuccessNoUser(t *testing.T, mock sqlmock.Sqlmock, store AppStore) {
	query := `
		select id_, username_, email_, status_, created_at_ 
		from users_ 
		where username_ = $1
	`

	mockRow := mock.
		NewRows([]string{"id_", "username_", "email_", "status_", "created_at_"})

	mock.
		ExpectQuery(regexp.QuoteMeta(query)).
		WithArgs("janedoe").
		WillReturnRows(mockRow)

	user, err := store.GetUserByUsername("janedoe")

	require.ErrorIs(t, err, sql.ErrNoRows, "should return no rows error")
	require.Empty(t, user, "should return empty user record")

	if expected := mock.ExpectationsWereMet(); expected != nil {
		t.Errorf("database was not queried as expected:\n%s", expected)
	}

}

func testSqliteAppStoreGetUserByUsernameScanError(t *testing.T, mock sqlmock.Sqlmock, store AppStore) {
	query := `
		select id_, username_, email_, status_, created_at_ 
		from users_ 
		where username_ = $1
	`

	createdAt := "2024-06-05 05:29:16"

	mockRows := mock.
		NewRows([]string{"id_", "username_", "email_", "status_", "created_at_"}).
		AddRow(23, "janedoe", "jane@example.org", "active", createdAt).
		RowError(0, errors.New("row_error"))

	mock.
		ExpectQuery(regexp.QuoteMeta(query)).
		WithArgs("janedoe").
		WillReturnRows(mockRows)

	user, err := store.GetUserByUsername("janedoe")

	require.Empty(t, user, "should return empty user result")
	require.EqualError(t, err, "row_error", "should return row error from database")

	if expected := mock.ExpectationsWereMet(); expected != nil {
		t.Errorf("did not query database as expected:\n%s", expected)
	}
}

func testSqliteAppStoreDeleteUserByUsernameSuccess(t *testing.T, mock sqlmock.Sqlmock, store AppStore) {
	query := `delete from users_ where username_ = $1`

	mockResult := sqlmock.NewResult(23, 1)

	mock.
		ExpectExec(regexp.QuoteMeta(query)).
		WithArgs("janedoe").
		WillReturnResult(mockResult)

	err := store.DeleteUserByUsername("janedoe")

	require.NoError(t, err, "should not return error")

	if expected := mock.ExpectationsWereMet(); expected != nil {
		t.Errorf("did not query database as expected:\n%s", expected)
	}
}

func testSqliteAppStoreDeleteUserByUsernameQueryError(t *testing.T, mock sqlmock.Sqlmock, store AppStore) {
	query := `delete from users_ where username_ = $1`

	mock.
		ExpectExec(regexp.QuoteMeta(query)).
		WithArgs("janedoe").
		WillReturnError(errors.New("query_error"))

	err := store.DeleteUserByUsername("janedoe")

	require.EqualError(t, err, "query_error", "should return error from query")

	if expected := mock.ExpectationsWereMet(); expected != nil {
		t.Errorf("did not query database as expected:\n%s", expected)
	}
}

func testSqliteAppStoreDeleteUserByUsernameRowsError(t *testing.T, mock sqlmock.Sqlmock, store AppStore) {
	query := `delete from users_ where username_ = $1`

	mockRes := sqlmock.NewErrorResult(errors.New("rows_error"))

	mock.
		ExpectExec(regexp.QuoteMeta(query)).
		WithArgs("janedoe").
		WillReturnResult(mockRes)

	err := store.DeleteUserByUsername("janedoe")

	require.EqualError(t, err, "rows_error", "should return error from query")

	if expected := mock.ExpectationsWereMet(); expected != nil {
		t.Errorf("did not query database as expected:\n%s", expected)
	}
}

func testSqliteAppStoreDeleteUserByUsernameNoUserError(t *testing.T, mock sqlmock.Sqlmock, store AppStore) {
	query := `delete from users_ where username_ = $1`

	mockRes := sqlmock.NewResult(0, 0)

	mock.
		ExpectExec(regexp.QuoteMeta(query)).
		WithArgs("janedoe").
		WillReturnResult(mockRes)

	err := store.DeleteUserByUsername("janedoe")

	require.EqualError(t, err, "no user deleted", "should return error due to zero results")

	if expected := mock.ExpectationsWereMet(); expected != nil {
		t.Errorf("did not query database as expected:\n%s", expected)
	}
}

func testSqliteAppStoreUpdateUserSuccess(t *testing.T, mock sqlmock.Sqlmock, store AppStore) {
	query := `
		update users_ set email_ = $2, set status_ = $3
		where username_ = $1 
		returning id_, username_, email_, status_, created_at_
	`

	createdAt := "2024-06-05 05:29:16"

	mockRow := mock.
		NewRows([]string{"id_", "username_", "email_", "status_", "created_at_"}).
		AddRow(23, "janedoe", "jane@example.org", "active", createdAt)

	mock.
		ExpectQuery(regexp.QuoteMeta(query)).
		WithArgs("janedoe", "jane@example.org", "active").
		WillReturnRows(mockRow)

	user, err := store.UpdateUser(models.User{
		Id:        23,
		Username:  "janedoe",
		Email:     "jane@example.org",
		Status:    "active",
		CreatedAt: createdAt,
	})

	require.NoError(t, err, "should not return error")
	require.Equal(t, &models.User{
		Id:        23,
		Username:  "janedoe",
		Email:     "jane@example.org",
		Status:    "active",
		CreatedAt: createdAt,
	}, user, "should return user record")

	if expected := mock.ExpectationsWereMet(); expected != nil {
		t.Errorf("database was not queried as expected:\n%s", expected)
	}
}

func testSqliteAppStoreUpdateUserScanError(t *testing.T, mock sqlmock.Sqlmock, store AppStore) {
	query := `
		update users_ set email_ = $2, set status_ = $3
		where username_ = $1 
		returning id_, username_, email_, status_, created_at_
	`

	createdAt := "2024-06-05 05:29:16"

	mockRow := mock.
		NewRows([]string{"id_", "username_", "email_", "status_", "created_at_"}).
		AddRow(23, "janedoe", "jane@example.org", "active", createdAt).
		RowError(0, errors.New("row_error"))

	mock.
		ExpectQuery(regexp.QuoteMeta(query)).
		WithArgs("janedoe", "jane@example.org", "active").
		WillReturnRows(mockRow)

	user, err := store.UpdateUser(models.User{
		Id:        23,
		Username:  "janedoe",
		Email:     "jane@example.org",
		Status:    "active",
		CreatedAt: createdAt,
	})

	require.EqualError(t, err, "row_error", "should return row error from database")
	require.Empty(t, user, "should return user record")

	if expected := mock.ExpectationsWereMet(); expected != nil {
		t.Errorf("database was not queried as expected:\n%s", expected)
	}
}

func testSqliteAppStoreInsertKeySuccess(t *testing.T, mock sqlmock.Sqlmock, store AppStore) {
	query := `
		insert into keys_ (user_id_, ssh_public_key_)
		values ($userId, $publicKey)
		returning id_, user_id_, ssh_public_key_, created_at_
	`

	createdAt := "2024-06-05 05:29:16"

	mockRow := sqlmock.
		NewRows([]string{"id_", "user_id_", "ssh_public_key_", "created_at_"}).
		AddRow(42, 23, "some_public_key", createdAt)

	mock.
		ExpectQuery(regexp.QuoteMeta(query)).
		WithArgs(42, "some_public_key").
		WillReturnRows(mockRow)

	insertedKeyDetails, err := store.InsertKey(42, "some_public_key")
	require.NoError(t, err, "should not return error")
	require.Equal(t, &models.Key{
		Id:        42,
		UserId:    23,
		PublicKey: "some_public_key",
		CreatedAt: createdAt,
	}, insertedKeyDetails, "should return inserted key details")

	if expected := mock.ExpectationsWereMet(); expected != nil {
		t.Errorf("database was not queried as expected:\n%s", expected)
	}
}

func testSqliteAppStoreInsertKeyScanError(t *testing.T, mock sqlmock.Sqlmock, store AppStore) {
	query := `
		insert into keys_ (user_id_, ssh_public_key_)
		values ($userId, $publicKey)
		returning id_, user_id_, ssh_public_key_, created_at_
	`

	createdAt := "2024-06-05 05:29:16"

	mockRow := sqlmock.
		NewRows([]string{"id_", "user_id_", "ssh_public_key_", "created_at_"}).
		AddRow(42, 23, "some_public_key", createdAt).
		RowError(0, errors.New("row_error"))

	mock.
		ExpectQuery(regexp.QuoteMeta(query)).
		WithArgs(42, "some_public_key").
		WillReturnRows(mockRow)

	insertedKeyDetails, err := store.InsertKey(42, "some_public_key")
	require.EqualError(t, err, "row_error", "should return row error")
	require.Empty(t, insertedKeyDetails, "should return empty key details")

	if expected := mock.ExpectationsWereMet(); expected != nil {
		t.Errorf("database was not queried as expected:\n%s", expected)
	}
}

func testSqliteAppStoreInsertDatabaseSuccess(t *testing.T, mock sqlmock.Sqlmock, store AppStore) {
	query := `
		insert into databases_ (name_, user_id_)
		values($name, $userId)
		returning id_, name_, user_id_, created_at_
	`

	createdAt := "2024-06-05 05:29:16"

	mockRow := sqlmock.
		NewRows([]string{"id_", "name_", "user_id_", "created_at_"}).
		AddRow(23, "dbname", 42, createdAt)

	mock.
		ExpectQuery(regexp.QuoteMeta(query)).
		WithArgs("dbname", 42).
		WillReturnRows(mockRow)

	insertedDatabaseDetails, err := store.InsertDatabase("dbname", 42)
	require.NoError(t, err, "should not return row error")
	require.Equal(t, &models.Database{
		Id:        23,
		Name:      "dbname",
		UserId:    42,
		CreatedAt: createdAt,
	}, insertedDatabaseDetails, "should return created database details")

	if expected := mock.ExpectationsWereMet(); expected != nil {
		t.Errorf("database was not queried as expected:\n%s", expected)
	}
}

func testSqliteAppStoreInsertDatabaseScanError(t *testing.T, mock sqlmock.Sqlmock, store AppStore) {
	query := `
		insert into databases_ (name_, user_id_)
		values($name,  $userId)
		returning id_, name_, user_id_, created_at_
	`

	createdAt := "2024-06-05 05:29:16"

	mockRow := sqlmock.
		NewRows([]string{"id_", "name_", "user_id_", "created_at_"}).
		AddRow(23, "dbname", 42, createdAt).
		RowError(0, errors.New("row_error"))

	mock.
		ExpectQuery(regexp.QuoteMeta(query)).
		WithArgs("dbname", 42).
		WillReturnRows(mockRow)

	insertedDatabaseDetails, err := store.InsertDatabase("dbname", 42)
	require.EqualError(t, err, "row_error", "should return row error")
	require.Empty(t, insertedDatabaseDetails, "should return empty database details")

	if expected := mock.ExpectationsWereMet(); expected != nil {
		t.Errorf("database was not queried as expected:\n%s", expected)
	}
}
