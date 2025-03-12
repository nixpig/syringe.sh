package stores_test

import (
	"database/sql"
	"fmt"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/nixpig/syringe.sh/internal/stores"
	"github.com/stretchr/testify/require"
)

const (
	getUserQuery = `select u.id_, u.username_, u.email_, u.verified_, k.public_key_sha1_
		from users_ u inner join public_keys_ k on u.id_ = k.user_id_ where u.username_ = $username`
	createUserQuery = `insert into users_ (username_, email_, verified_)
		values ($username, $email, $verified) returning id_`
	createKeyQuery = `insert into public_keys_ (public_key_sha1_, user_id_)
		values ($publicKeySHA1, $userID)`
)

func TestSystemStore(t *testing.T) {
	scenarios := map[string]func(
		t *testing.T,
		store *stores.SystemStore,
		mock sqlmock.Sqlmock,
	){
		"get user from system store (success)":     testGetUserFromSystemStoreSuccess,
		"get user from system store (no user)":     testGetUserFromSystemStoreNoUser,
		"create user in system store (success)":    testCreateUserInSystemStoreSuccess,
		"create user in system store (user error)": testCreateUserInSystemStoreUserErr,
		"create user in system store (key error)":  testCreateUserInSystemStoreKeyErr,
	}

	for scenario, fn := range scenarios {
		t.Run(scenario, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("failed to create mock database: %s", err)
			}
			defer db.Close()

			store := stores.NewSystemStore(db)

			fn(t, store, mock)
		})
	}
}

func testGetUserFromSystemStoreSuccess(
	t *testing.T,
	store *stores.SystemStore,
	mock sqlmock.Sqlmock,
) {
	mock.ExpectQuery(
		regexp.QuoteMeta(getUserQuery),
	).WithArgs(
		sql.Named("username", "janedoe"),
	).WillReturnRows(
		sqlmock.
			NewRows(
				[]string{"id_", "username_", "email_", "verified_", "public_key_sha1_"},
			).AddRow(23, "janedoe", "janedoe@example.org", true, "some_public_key"),
	)

	user, err := store.GetUser("janedoe")

	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
	require.Equal(t, &stores.User{
		ID:            23,
		Username:      "janedoe",
		Email:         "janedoe@example.org",
		Verified:      true,
		PublicKeySHA1: "some_public_key",
	}, user)
}

func testGetUserFromSystemStoreNoUser(
	t *testing.T,
	store *stores.SystemStore,
	mock sqlmock.Sqlmock,
) {
	mock.ExpectQuery(
		regexp.QuoteMeta(getUserQuery),
	).WithArgs(
		sql.Named("username", "janedoe"),
	).WillReturnRows(sqlmock.NewRows(
		[]string{"id_", "username_", "email_", "verified_", "public_key_sha1_"},
	))

	user, err := store.GetUser("janedoe")

	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
	require.Nil(t, user)
}

func testGetUserFromSystemStoreRowErr(
	t *testing.T,
	store *stores.SystemStore,
	mock sqlmock.Sqlmock,
) {
	mock.ExpectQuery(
		regexp.QuoteMeta(getUserQuery),
	).WithArgs(
		sql.Named("username", "janedoe"),
	).WillReturnRows(sqlmock.NewRows(
		[]string{"id_", "username_", "email_", "verified_", "public_key_sha1_"},
	).RowError(1, fmt.Errorf("row_err")))

	user, err := store.GetUser("janedoe")

	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
	require.Nil(t, user)
}

func testCreateUserInSystemStoreSuccess(
	t *testing.T,
	store *stores.SystemStore,
	mock sqlmock.Sqlmock,
) {
	mock.ExpectBegin()
	mock.ExpectQuery(
		regexp.QuoteMeta(createUserQuery),
	).WithArgs(
		sql.Named("username", "janedoe"),
		sql.Named("email", "janedoe@example.org"),
		sql.Named("verified", true),
		sql.Named("publicKeySHA1", "some_public_key"),
	).WillReturnRows(sqlmock.NewRows(
		[]string{"id_"},
	).AddRow(23))

	mock.ExpectExec(
		regexp.QuoteMeta(createKeyQuery),
	).WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectCommit()

	userID, err := store.CreateUser(&stores.User{
		Username:      "janedoe",
		Email:         "janedoe@example.org",
		Verified:      true,
		PublicKeySHA1: "some_public_key",
	})

	require.NoError(t, err)
	require.Equal(t, 23, userID)
	require.NoError(t, mock.ExpectationsWereMet())
}

func testCreateUserInSystemStoreUserErr(
	t *testing.T,
	store *stores.SystemStore,
	mock sqlmock.Sqlmock,
) {
	mock.ExpectBegin()
	mock.ExpectQuery(
		regexp.QuoteMeta(createUserQuery),
	).WithArgs(
		sql.Named("username", "janedoe"),
		sql.Named("email", "janedoe@example.org"),
		sql.Named("verified", true),
		sql.Named("publicKeySHA1", "some_public_key"),
	).WillReturnRows(sqlmock.NewRows(
		[]string{"id_"},
	))

	mock.ExpectRollback()

	userID, err := store.CreateUser(&stores.User{
		Username:      "janedoe",
		Email:         "janedoe@example.org",
		Verified:      true,
		PublicKeySHA1: "some_public_key",
	})

	require.Error(t, err)
	require.Equal(t, 0, userID)
	require.NoError(t, mock.ExpectationsWereMet())
}

func testCreateUserInSystemStoreKeyErr(
	t *testing.T,
	store *stores.SystemStore,
	mock sqlmock.Sqlmock,
) {
	mock.ExpectBegin()
	mock.ExpectQuery(
		regexp.QuoteMeta(createUserQuery),
	).WithArgs(
		sql.Named("username", "janedoe"),
		sql.Named("email", "janedoe@example.org"),
		sql.Named("verified", true),
		sql.Named("publicKeySHA1", "some_public_key"),
	).WillReturnRows(sqlmock.NewRows(
		[]string{"id_"},
	).AddRow(23))

	mock.ExpectExec(
		regexp.QuoteMeta(createKeyQuery),
	).WillReturnError(fmt.Errorf("key_err"))

	mock.ExpectRollback()

	userID, err := store.CreateUser(&stores.User{
		Username:      "janedoe",
		Email:         "janedoe@example.org",
		Verified:      true,
		PublicKeySHA1: "some_public_key",
	})

	require.Error(t, err)
	require.Equal(t, 0, userID)
	require.NoError(t, mock.ExpectationsWereMet())
}
