package auth_test

// import (
// 	"crypto/rand"
// 	"crypto/rsa"
// 	"database/sql"
// 	"errors"
// 	"regexp"
// 	"strings"
// 	"testing"
// 	"time"
//
// 	"github.com/DATA-DOG/go-sqlmock"
// 	"github.com/charmbracelet/ssh"
// 	"github.com/go-playground/validator/v10"
// 	"github.com/nixpig/syringe.sh/internal/auth"
// 	"github.com/stretchr/testify/require"
// 	gossh "golang.org/x/crypto/ssh"
// )
//
// func TestAuthInternalPkg(t *testing.T) {
// 	scenarios := map[string]func(t *testing.T, mock sqlmock.Sqlmock, db *sql.DB, service auth.AuthService){
// 		"test authenticate user with matching key":       testAuthUserWithMatchingKey,
// 		"test authenticate user with non-matching key":   testAuthUserWithNonMatchingKey,
// 		"test authenticate user with no keys":            testAuthUserWithNoKeys,
// 		"test authenticate user when key parsing errors": testAuthUserKeyParsingError,
// 		"test authenticate user db query error":          testAuthUserDBQueryError,
// 		"test authenticate user db scan error":           testAuthUserDBScanError,
// 	}
//
// 	for scenario, fn := range scenarios {
// 		t.Run(scenario, func(t *testing.T) {
// 			db, mock, err := sqlmock.New()
// 			if err != nil {
// 				t.Error("failed to create mock db")
// 			}
//
// 			store := auth.NewSqliteAuthStore(db)
//
// 			service := auth.NewAuthService(
// 				store,
// 				validator.New(),
// 			)
//
// 			fn(t, mock, db, service)
// 		})
// 	}
// }
//
// func generatePublicKey() (ssh.PublicKey, error) {
// 	privateKey, err := rsa.GenerateKey(rand.Reader, 4096)
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	publicKey, err := gossh.NewPublicKey(&privateKey.PublicKey)
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	charmPublicKey, ok := publicKey.(ssh.PublicKey)
// 	if !ok {
// 		return nil, errors.New("failed to cast public key")
// 	}
//
// 	return charmPublicKey, err
// }
//
// func testAuthUserWithMatchingKey(
// 	t *testing.T,
// 	mock sqlmock.Sqlmock,
// 	db *sql.DB,
// 	service auth.AuthService,
// ) {
// 	key, err := generatePublicKey()
// 	if err != nil {
// 		t.Errorf("failed to generate public key: %s", err)
// 	}
//
// 	mock.ExpectQuery(regexp.QuoteMeta(`
// 		select k.id_, k.user_id_, k.ssh_public_key_, k.created_at_
// 		from keys_ k
// 		inner join
// 		users_ u
// 		on k.user_id_ = u.id_
// 		where u.username_ = $username
// 		`)).WithArgs("janedoe").
// 		WillReturnRows(
// 			sqlmock.
// 				NewRows([]string{"id_", "user_id_", "ssh_public_key_", "created_at_"}).
// 				AddRow(23, 42, gossh.MarshalAuthorizedKey(key), time.Now().String()),
// 		)
//
// 	res, err := service.AuthenticateUser(auth.AuthenticateUserRequest{
// 		Username:  "janedoe",
// 		PublicKey: key,
// 	})
//
// 	require.NoError(t, err)
//
// 	require.Equal(
// 		t,
// 		&auth.AuthenticateUserResponse{
// 			Auth: true,
// 		},
// 		res,
// 	)
// }
//
// func testAuthUserWithNonMatchingKey(
// 	t *testing.T,
// 	mock sqlmock.Sqlmock,
// 	db *sql.DB,
// 	service auth.AuthService,
// ) {
// 	key1, err := generatePublicKey()
// 	if err != nil {
// 		t.Errorf("failed to generate public key: %s", err)
// 	}
//
// 	key2, err := generatePublicKey()
// 	if err != nil {
// 		t.Errorf("failed to generate public key: %s", err)
// 	}
//
// 	mock.ExpectQuery(regexp.QuoteMeta(`
// 		select k.id_, k.user_id_, k.ssh_public_key_, k.created_at_
// 		from keys_ k
// 		inner join
// 		users_ u
// 		on k.user_id_ = u.id_
// 		where u.username_ = $username
// 		`)).WithArgs("janedoe").
// 		WillReturnRows(
// 			sqlmock.
// 				NewRows([]string{"id_", "user_id_", "ssh_public_key_", "created_at_"}).
// 				AddRow(23, 42, gossh.MarshalAuthorizedKey(key1), time.Now().String()),
// 		)
//
// 	res, err := service.AuthenticateUser(auth.AuthenticateUserRequest{
// 		Username:  "janedoe",
// 		PublicKey: key2,
// 	})
//
// 	require.NoError(t, err)
//
// 	require.Equal(
// 		t,
// 		&auth.AuthenticateUserResponse{
// 			Auth: false,
// 		},
// 		res,
// 	)
// }
//
// func testAuthUserWithNoKeys(
// 	t *testing.T,
// 	mock sqlmock.Sqlmock,
// 	db *sql.DB,
// 	service auth.AuthService,
// ) {
// 	mock.ExpectQuery(regexp.QuoteMeta(`
// 		select k.id_, k.user_id_, k.ssh_public_key_, k.created_at_
// 		from keys_ k
// 		inner join
// 		users_ u
// 		on k.user_id_ = u.id_
// 		where u.username_ = $username
// 		`)).WithArgs("janedoe").
// 		WillReturnRows(
// 			sqlmock.
// 				NewRows([]string{"id_", "user_id_", "ssh_public_key_", "created_at_"}),
// 		)
//
// 	key, err := generatePublicKey()
// 	if err != nil {
// 		t.Errorf("failed to generate public key: %s", err)
// 	}
//
// 	res, err := service.AuthenticateUser(auth.AuthenticateUserRequest{
// 		Username:  "janedoe",
// 		PublicKey: key,
// 	})
//
// 	require.NoError(t, err)
//
// 	require.Equal(
// 		t,
// 		&auth.AuthenticateUserResponse{
// 			Auth: false,
// 		},
// 		res,
// 	)
// }
//
// func testAuthUserKeyParsingError(
// 	t *testing.T,
// 	mock sqlmock.Sqlmock,
// 	db *sql.DB,
// 	service auth.AuthService,
// ) {
// 	key, err := generatePublicKey()
// 	if err != nil {
// 		t.Errorf("failed to generate public key: %s", err)
// 	}
//
// 	mock.ExpectQuery(regexp.QuoteMeta(`
// 		select k.id_, k.user_id_, k.ssh_public_key_, k.created_at_
// 		from keys_ k
// 		inner join
// 		users_ u
// 		on k.user_id_ = u.id_
// 		where u.username_ = $username
// 		`)).WithArgs("janedoe").
// 		WillReturnRows(
// 			sqlmock.
// 				NewRows([]string{"id_", "user_id_", "ssh_public_key_", "created_at_"}).
// 				AddRow(23, 42, "invalid key", time.Now().String()),
// 		)
//
// 	res, err := service.AuthenticateUser(auth.AuthenticateUserRequest{
// 		Username:  "janedoe",
// 		PublicKey: key,
// 	})
//
// 	require.Nil(t, res)
// 	require.Error(t, err)
// 	require.EqualError(t, err, "ssh: no key found")
// }
//
// func testAuthUserDBQueryError(
// 	t *testing.T,
// 	mock sqlmock.Sqlmock,
// 	db *sql.DB,
// 	service auth.AuthService,
// ) {
// 	mock.ExpectQuery(regexp.QuoteMeta(`
// 		select k.id_, k.user_id_, k.ssh_public_key_, k.created_at_
// 		from keys_ k
// 		inner join
// 		users_ u
// 		on k.user_id_ = u.id_
// 		where u.username_ = $username
// 		`)).WithArgs("janedoe").
// 		WillReturnError(errors.New("database_error"))
//
// 	key, err := generatePublicKey()
// 	if err != nil {
// 		t.Errorf("failed to generate public key: %s", err)
// 	}
//
// 	res, err := service.AuthenticateUser(auth.AuthenticateUserRequest{
// 		Username:  "janedoe",
// 		PublicKey: key,
// 	})
//
// 	require.Nil(t, res)
// 	require.Error(t, err)
// 	require.EqualError(t, err, "database_error")
// }
//
// func testAuthUserDBScanError(
// 	t *testing.T,
// 	mock sqlmock.Sqlmock,
// 	db *sql.DB,
// 	service auth.AuthService,
// ) {
// 	key, err := generatePublicKey()
// 	if err != nil {
// 		t.Errorf("failed to generate public key: %s", err)
// 	}
//
// 	mock.ExpectQuery(regexp.QuoteMeta(`
// 		select k.id_, k.user_id_, k.ssh_public_key_, k.created_at_
// 		from keys_ k
// 		inner join
// 		users_ u
// 		on k.user_id_ = u.id_
// 		where u.username_ = $username
// 		`)).WithArgs("janedoe").
// 		WillReturnRows(
// 			sqlmock.
// 				NewRows([]string{"id_", "user_id_", "ssh_public_key_", "created_at_"}).
// 				AddRow(23, "invalid user id to trigger scan error", gossh.MarshalAuthorizedKey(key), time.Now().String()),
// 		)
//
// 	res, err := service.AuthenticateUser(auth.AuthenticateUserRequest{
// 		Username:  "janedoe",
// 		PublicKey: key,
// 	})
//
// 	require.Nil(t, res)
// 	require.Error(t, err)
// 	require.True(t, strings.HasPrefix(err.Error(), "sql: Scan error"))
// }
