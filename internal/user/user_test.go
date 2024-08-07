package user_test

// import (
// 	"bytes"
// 	"context"
// 	"database/sql"
// 	"os"
// 	"regexp"
// 	"testing"
//
// 	"github.com/DATA-DOG/go-sqlmock"
// 	"github.com/nixpig/syringe.sh/internal/user"
// 	"github.com/nixpig/syringe.sh/pkg/ctxkeys"
// 	"github.com/nixpig/syringe.sh/pkg/validation"
// 	"github.com/nixpig/syringe.sh/test"
// 	"github.com/spf13/cobra"
// 	"github.com/stretchr/testify/require"
// 	gossh "golang.org/x/crypto/ssh"
// )
//
// func TestUserCmd(t *testing.T) {
// 	scenarios := map[string]func(
// 		t *testing.T,
// 		cmd *cobra.Command,
// 		service user.UserService,
// 		mainDBMock sqlmock.Sqlmock,
// 		userDBMock sqlmock.Sqlmock,
// 	){
// 		// only test the 'main' scenarios - there's a lot of mocking and arguably little value in testing all the edge cases
// 		"test user register happy path": testUserRegisterHappyPath,
// 	}
//
// 	for scenario, fn := range scenarios {
// 		mainDB, mainDBMock, err := sqlmock.New()
// 		if err != nil {
// 			t.Fatalf("failed to create mock db: \n%s", err)
// 		}
//
// 		userDB, userDBMock, err := sqlmock.New()
// 		if err != nil {
// 			t.Fatalf("failed to create mock db: \n%s", err)
// 		}
//
// 		var databaseConnector = func(filename, user, password string) (*sql.DB, error) {
// 			return userDB, nil
// 		}
//
// 		store := user.NewSqliteUserStore(mainDB)
//
// 		service := user.NewUserServiceImpl(
// 			store,
// 			validation.New(),
// 			databaseConnector,
// 		)
//
// 		cmd := user.NewCmdUser()
//
// 		t.Run(scenario, func(t *testing.T) {
// 			fn(
// 				t,
// 				cmd,
// 				service,
// 				mainDBMock,
// 				userDBMock,
// 			)
// 		})
// 	}
// }
//
// func testUserRegisterHappyPath(
// 	t *testing.T,
// 	cmd *cobra.Command,
// 	service user.UserService,
// 	mainDBMock sqlmock.Sqlmock,
// 	userDBMock sqlmock.Sqlmock,
// ) {
// 	var err error
//
// 	cmdIn := bytes.NewReader([]byte{})
// 	cmdOut := bytes.NewBufferString("")
// 	errOut := bytes.NewBufferString("")
//
// 	handler := user.NewHandlerUserRegister(service)
// 	cmdRegister := user.NewCmdUserRegister(handler)
//
// 	publicKey, _, err := test.GenerateKeyPair()
// 	require.NoError(t, err)
//
// 	ctx := context.Background()
// 	ctx = context.WithValue(ctx, ctxkeys.Username, "janedoe")
// 	ctx = context.WithValue(ctx, ctxkeys.PublicKey, publicKey)
//
// 	cmd.SetContext(ctx)
//
// 	cmd.AddCommand(cmdRegister)
// 	cmd.SetIn(cmdIn)
// 	cmd.SetOut(cmdOut)
// 	cmd.SetErr(errOut)
//
// 	cmd.SetArgs([]string{"register"})
//
// 	mainDBMock.
// 		ExpectQuery(regexp.QuoteMeta(`
// 			insert into users_ (username_, email_, status_)
// 			values ($username, $email, $status)
// 			returning id_, username_, email_, status_, created_at_
// 		`)).
// 		WithArgs("janedoe", "not_used_yet@example.org", "active").
// 		WillReturnRows(
// 			mainDBMock.NewRows([]string{"id_", "username_", "email_", "status_", "created_at_"}).
// 				AddRow(23, "janedoe", "not_used_yet@example.org", "active", "sometime"),
// 		)
//
// 	mainDBMock.
// 		ExpectQuery(regexp.QuoteMeta(`
// 			insert into keys_ (user_id_, ssh_public_key_)
// 			values ($userID, $publicKey)
// 			returning id_, user_id_, ssh_public_key_, created_at_
// 		`)).
// 		WithArgs(23, string(gossh.MarshalAuthorizedKey(publicKey))).
// 		WillReturnRows(
// 			mainDBMock.
// 				NewRows([]string{"id_", "user_id_", "public_key_", "created_at_"}).
// 				AddRow(23, 23, string(gossh.MarshalAuthorizedKey(publicKey)), "somsdf"),
// 		)
//
// 	os.Setenv("DATABASE_ORG", "mock_db_org")
// 	os.Setenv("DATABASE_GROUP", "mock_db_group")
//
// 	// projectsQuery := `
// 	// 	create table if not exists projects_ (
// 	// 		id_ integer primary key autoincrement,
// 	// 		name_ varchar(256) unique not null
// 	// 	)
// 	// `
// 	//
// 	// environmentsQuery := `
// 	// 	create table if not exists environments_ (
// 	// 		id_ integer primary key autoincrement,
// 	// 		name_ varchar(256) not null,
// 	// 		project_id_ integer not null,
// 	//
// 	// 		foreign key (project_id_) references projects_(id_) on delete cascade
// 	// 	)
// 	// `
// 	// secretsQuery := `
// 	// 	create table if not exists secrets_ (
// 	// 		id_ integer primary key autoincrement,
// 	// 		key_ text not null unique,
// 	// 		value_ text not null,
// 	// 		environment_id_ integer not null,
// 	//
// 	// 		foreign key (environment_id_) references environments_(id_) on delete cascade
// 	// 	)
// 	// `
//
// 	userDBMock.ExpectExec(regexp.QuoteMeta(`
// 		CREATE TABLE IF NOT EXISTS schema_migrations (version uint64,dirty bool);
//     CREATE UNIQUE INDEX IF NOT EXISTS version_unique ON schema_migrations (version);
// 	`))
//
// 	// userDBMock.ExpectBegin()
// 	// userDBMock.ExpectExec(regexp.QuoteMeta(projectsQuery))
// 	// userDBMock.ExpectExec(regexp.QuoteMeta(environmentsQuery))
// 	// userDBMock.ExpectExec(regexp.QuoteMeta(secretsQuery))
// 	// userDBMock.ExpectCommit()
//
// 	err = cmd.Execute()
// 	require.NoError(t, err)
//
// 	require.Empty(t, errOut.String())
// 	require.Equal(t, "User 'janedoe' registered successfully!", cmdOut.String())
//
// 	require.NoError(t, mainDBMock.ExpectationsWereMet())
// 	require.NoError(t, userDBMock.ExpectationsWereMet())
// }
