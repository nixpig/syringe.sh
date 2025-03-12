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
		"get user from system store (success)": testGetUserFromSystemStoreSuccess,
		"get user from system store (no user)": testGetUserFromSystemStoreNoUser,
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
