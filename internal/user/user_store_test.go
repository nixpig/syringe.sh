package user

import (
	"database/sql"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
)

func TestSqliteUserStore(t *testing.T) {
	scenarios := map[string]func(t *testing.T, mock sqlmock.Sqlmock, store UserStore){
		"test sqlite user store insert user (success)":    testSqliteUserStoreInsertUserSuccess,
		"test sqlite user store insert user (scan error)": testSqliteUserStoreInsertUserScanError,

		"test sqlite user store get user by username (success)":           testSqliteUserStoreGetUserByUsernameSuccess,
		"test sqlite user store get user by username (success - no user)": testSqliteUserStoreGetUserByUsernameSuccessNoUser,
		"test sqlite user store get user by username (scan error)":        testSqliteUserStoreGetUserByUsernameScanError,

		"test sqlite user store delete user by username (success)":         testSqliteUserStoreDeleteUserByUsernameSuccess,
		"test sqlite user store delete user by username (query error)":     testSqliteUserStoreDeleteUserByUsernameQueryError,
		"test sqlite user store delete user by username (rows error)":      testSqliteUserStoreDeleteUserByUsernameRowsError,
		"test sqlite user store delete user by username (no user deleted)": testSqliteUserStoreDeleteUserByUsernameNoUserError,

		"test sqlite user store update user (success)":    testSqliteUserStoreUpdateUserSuccess,
		"test sqlite user store update user (scan error)": testSqliteUserStoreUpdateUserScanError,
	}

	for scenario, fn := range scenarios {
		t.Run(scenario, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatal("unable to create new mock sqlite database")
			}

			store := NewSqliteUserStore(db)

			fn(t, mock, store)
		})
	}
}

func testSqliteUserStoreInsertUserSuccess(t *testing.T, mock sqlmock.Sqlmock, store UserStore) {
	query := `
		insert into users_ (username_, email_, status_, created_at_) 
		values ($username, $email, $status, $createdAt) 
		returning id_, username_, email_, status_, created_at_
	`

	createdAt := time.Now().UTC()

	mockRow := mock.
		NewRows([]string{"id_", "username_", "email_", "status_", "created_at_"}).
		AddRow(23, "janedoe", "jane@example.org", "p4ssw0rd", createdAt)

	mock.
		ExpectQuery(regexp.QuoteMeta(query)).
		WithArgs("janedoe", "jane@example.org", "p4ssw0rd", sqlmock.AnyArg()).
		WillReturnRows(mockRow)

	insertedUser, err := store.Insert("janedoe", "jane@example.org", "p4ssw0rd")

	require.NoError(t, err, "should not return error")
	require.Equal(t, &User{
		Id:        23,
		Username:  "janedoe",
		Email:     "jane@example.org",
		Status:    "p4ssw0rd",
		CreatedAt: createdAt,
	}, insertedUser, "should return inserted user record")

	if expected := mock.ExpectationsWereMet(); expected != nil {
		t.Error("database was not called as expected")
	}
}

func testSqliteUserStoreInsertUserScanError(t *testing.T, mock sqlmock.Sqlmock, store UserStore) {
	query := `
		insert into users_ (username_, email_, status_, created_at_) 
		values ($username, $email, $status, $createdAt) 
		returning id_, username_, email_, status_, created_at_
	`

	createdAt := time.Now().UTC()

	mockRow := mock.
		NewRows([]string{"id_", "username_", "email_", "status_", "created_at_"}).
		AddRow(23, "janedoe", "jane@example.org", "p4ssw0rd", createdAt).
		RowError(0, errors.New("row_error"))

	mock.
		ExpectQuery(regexp.QuoteMeta(query)).
		WithArgs("janedoe", "jane@example.org", "p4ssw0rd", sqlmock.AnyArg()).
		WillReturnRows(mockRow)

	insertedUser, err := store.Insert("janedoe", "jane@example.org", "p4ssw0rd")

	require.EqualError(t, err, "row_error", "should return row error from database")
	require.Empty(t, insertedUser, "should return empty inserted user record")

	if expected := mock.ExpectationsWereMet(); expected != nil {
		t.Errorf("database was not called as expected:\n%s", err)
	}
}

func testSqliteUserStoreGetUserByUsernameSuccess(t *testing.T, mock sqlmock.Sqlmock, store UserStore) {
	query := `
		select id_, username_, email_, status_, created_at_ 
		from users_ 
		where username_ = $1
	`
	createdAt := time.Now()

	mockRow := mock.
		NewRows([]string{"id_", "username_", "email_", "status_", "created_at_"}).
		AddRow(23, "janedoe", "jane@example.org", "p4ssw0rd", createdAt)

	mock.
		ExpectQuery(regexp.QuoteMeta(query)).
		WithArgs("janedoe").
		WillReturnRows(mockRow)

	user, err := store.GetByUsername("janedoe")

	require.NoError(t, err, "should not return error")
	require.Equal(t, &User{
		Id:        23,
		Username:  "janedoe",
		Email:     "jane@example.org",
		Status:    "p4ssw0rd",
		CreatedAt: createdAt,
	}, user, "should return user record")

	if expected := mock.ExpectationsWereMet(); expected != nil {
		t.Errorf("database was not queried as expected:\n%s", err)
	}
}

func testSqliteUserStoreGetUserByUsernameSuccessNoUser(t *testing.T, mock sqlmock.Sqlmock, store UserStore) {
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

	user, err := store.GetByUsername("janedoe")

	require.ErrorIs(t, err, sql.ErrNoRows, "should return no rows error")
	require.Empty(t, user, "should return empty user record")

	if expected := mock.ExpectationsWereMet(); expected != nil {
		t.Errorf("database was not queried as expected:\n%s", err)
	}

}

func testSqliteUserStoreGetUserByUsernameScanError(t *testing.T, mock sqlmock.Sqlmock, store UserStore) {
	query := `
		select id_, username_, email_, status_, created_at_ 
		from users_ 
		where username_ = $1
	`

	createdAt := time.Now()

	mockRows := mock.
		NewRows([]string{"id_", "username_", "email_", "status_", "created_at_"}).
		AddRow(23, "janedoe", "jane@example.org", "p4ssw0rd", createdAt).
		RowError(0, errors.New("row_error"))

	mock.
		ExpectQuery(regexp.QuoteMeta(query)).
		WithArgs("janedoe").
		WillReturnRows(mockRows)

	user, err := store.GetByUsername("janedoe")

	require.Empty(t, user, "should return empty user result")
	require.EqualError(t, err, "row_error", "should return row error from database")

	if expected := mock.ExpectationsWereMet(); expected != nil {
		t.Errorf("did not query database as expected:\n%s", expected)
	}
}

func testSqliteUserStoreDeleteUserByUsernameSuccess(t *testing.T, mock sqlmock.Sqlmock, store UserStore) {
	query := `delete from users_ where username_ = $1`

	mockResult := sqlmock.NewResult(23, 1)

	mock.
		ExpectExec(regexp.QuoteMeta(query)).
		WithArgs("janedoe").
		WillReturnResult(mockResult)

	err := store.DeleteByUsername("janedoe")

	require.NoError(t, err, "should not return error")

	if expected := mock.ExpectationsWereMet(); expected != nil {
		t.Errorf("did not query database as expected:\n%s", expected)
	}
}

func testSqliteUserStoreDeleteUserByUsernameQueryError(t *testing.T, mock sqlmock.Sqlmock, store UserStore) {
	query := `delete from users_ where username_ = $1`

	mock.
		ExpectExec(regexp.QuoteMeta(query)).
		WithArgs("janedoe").
		WillReturnError(errors.New("query_error"))

	err := store.DeleteByUsername("janedoe")

	require.EqualError(t, err, "query_error", "should return error from query")

	if expected := mock.ExpectationsWereMet(); expected != nil {
		t.Errorf("did not query database as expected:\n%s", expected)
	}
}

func testSqliteUserStoreDeleteUserByUsernameRowsError(t *testing.T, mock sqlmock.Sqlmock, store UserStore) {
	query := `delete from users_ where username_ = $1`

	mockRes := sqlmock.NewErrorResult(errors.New("rows_error"))

	mock.
		ExpectExec(regexp.QuoteMeta(query)).
		WithArgs("janedoe").
		WillReturnResult(mockRes)

	err := store.DeleteByUsername("janedoe")

	require.EqualError(t, err, "rows_error", "should return error from query")

	if expected := mock.ExpectationsWereMet(); expected != nil {
		t.Errorf("did not query database as expected:\n%s", expected)
	}
}

func testSqliteUserStoreDeleteUserByUsernameNoUserError(t *testing.T, mock sqlmock.Sqlmock, store UserStore) {
	query := `delete from users_ where username_ = $1`

	mockRes := sqlmock.NewResult(0, 0)

	mock.
		ExpectExec(regexp.QuoteMeta(query)).
		WithArgs("janedoe").
		WillReturnResult(mockRes)

	err := store.DeleteByUsername("janedoe")

	require.EqualError(t, err, "no user deleted", "should return error due to zero results")

	if expected := mock.ExpectationsWereMet(); expected != nil {
		t.Errorf("did not query database as expected:\n%s", expected)
	}
}

func testSqliteUserStoreUpdateUserSuccess(t *testing.T, mock sqlmock.Sqlmock, store UserStore) {
	query := `
		update users_ set email_ = $2, set status_ = $3
		where username_ = $1 
		returning id_, username_, email_, status_, created_at_
	`

	createdAt := time.Now()

	mockRow := mock.
		NewRows([]string{"id_", "username_", "email_", "status_", "created_at_"}).
		AddRow(23, "janedoe", "jane@example.org", "p4ssw0rd", createdAt)

	mock.
		ExpectQuery(regexp.QuoteMeta(query)).
		WithArgs("janedoe", "jane@example.org", "p4ssw0rd").
		WillReturnRows(mockRow)

	user, err := store.Update(User{
		Id:        23,
		Username:  "janedoe",
		Email:     "jane@example.org",
		Status:    "p4ssw0rd",
		CreatedAt: createdAt,
	})

	require.NoError(t, err, "should not return error")
	require.Equal(t, &User{
		Id:        23,
		Username:  "janedoe",
		Email:     "jane@example.org",
		Status:    "p4ssw0rd",
		CreatedAt: createdAt,
	}, user, "should return user record")

	if expected := mock.ExpectationsWereMet(); expected != nil {
		t.Errorf("database was not queried as expected:\n%s", err)
	}
}

func testSqliteUserStoreUpdateUserScanError(t *testing.T, mock sqlmock.Sqlmock, store UserStore) {
	query := `
		update users_ set email_ = $2, set status_ = $3
		where username_ = $1 
		returning id_, username_, email_, status_, created_at_
	`

	createdAt := time.Now()

	mockRow := mock.
		NewRows([]string{"id_", "username_", "email_", "status_", "created_at_"}).
		AddRow(23, "janedoe", "jane@example.org", "p4ssw0rd", createdAt).
		RowError(0, errors.New("row_error"))

	mock.
		ExpectQuery(regexp.QuoteMeta(query)).
		WithArgs("janedoe", "jane@example.org", "p4ssw0rd").
		WillReturnRows(mockRow)

	user, err := store.Update(User{
		Id:        23,
		Username:  "janedoe",
		Email:     "jane@example.org",
		Status:    "p4ssw0rd",
		CreatedAt: createdAt,
	})

	require.EqualError(t, err, "row_error", "should return row error from database")
	require.Empty(t, user, "should return user record")

	if expected := mock.ExpectationsWereMet(); expected != nil {
		t.Errorf("database was not queried as expected:\n%s", err)
	}
}
