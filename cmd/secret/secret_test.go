package secret_test

import (
	"bytes"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/nixpig/syringe.sh/server/cmd"
	"github.com/nixpig/syringe.sh/server/cmd/secret"
	"github.com/nixpig/syringe.sh/server/test"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

func TestSecretCmd(t *testing.T) {
	scenarios := map[string]func(t *testing.T, mock sqlmock.Sqlmock, db *sql.DB){
		"test secret set command happy path":          testSecretSetCmdHappyPath,
		"test secret set command missing project":     testSecretSetCmdMissingProject,
		"test secret set command missing environment": testSecretSetCmdMissingEnvironment,
		"test secret set command too few args":        testSecretSetCmdTooFewArgs,
		"test secret set command too many args":       testSecretSetCmdTooManyArgs,
		"test secret set command database error":      testSecretSetCmdDatabaseError,
		"test secret set command validation error":    testSecretSetCmdValidationError,

		"test secret get command happy path":          testSecretGetCmdHappyPath,
		"test secret get command missing project":     testSecretGetCmdMissingProject,
		"test secret get command missing environment": testSecretGetCmdMissingEnvironment,
		"test secret get command missing key":         testSecretGetCmdMissingKey,
		"test secret get command database error":      testSecretGetCmdDatabaseError,
		"test secret get command validation error":    testSecretGetCmdValidationError,

		"test secret list command happy path":          testSecretListCmdHappyPath,
		"test secret list command database error":      testSecretListCmdDatabaseError,
		"test secret list command missing project":     testSecretListCmdMissingProject,
		"test secret list command missing environment": testSecretListCmdMissingEnvironment,
		"test secret list command validation error":    testSecretListCmdValidationError,

		"test secret remove command happy path":          testSecretRemoveCmdHappyPath,
		"test secret remove command database error":      testSecretRemoveCmdDatabaseError,
		"test secret remove command missing project":     testSecretRemoveCmdMissingProject,
		"test secret remove command missing environment": testSecretRemoveCmdMissingEnvironment,
		"test secret remove command missing key":         testSecretRemoveCmdMissingKey,
		"test secret remove command validation error":    testSecretRemoveCmdValidationError,
	}

	for scenario, fn := range scenarios {
		t.Run(scenario, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("unable to create mock database:\n%s", err)
			}

			fn(t, mock, db)
		})
	}
}

func testSecretSetCmdHappyPath(t *testing.T, mock sqlmock.Sqlmock, db *sql.DB) {
	cmdIn := bytes.NewReader([]byte{})
	cmdOut := bytes.NewBufferString("")

	query := `
		insert into secrets_ 
		(key_, value_, environment_id_) 
		values (
			$key,
			$value,
			(
				select e.id_ from 
					environments_ e
					inner join 
					projects_ p 
					on e.project_id_ = p.id_ 
					where p.name_ = $project 
					and e.name_ = $environment
			)
		)
	`

	mock.ExpectExec(regexp.QuoteMeta(query)).
		WithArgs(
			"my_cool_project",
			"staging",
			"secret_key",
			"secret_value",
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := cmd.Execute(
		[]*cobra.Command{secret.SecretCommand()},
		[]string{
			"secret",
			"set",
			"-p",
			"my_cool_project",
			"-e",
			"staging",
			"secret_key",
			"secret_value",
		},
		cmdIn,
		cmdOut,
		os.Stderr,
		db,
	)

	require.NoError(t, err)

	out, err := io.ReadAll(cmdOut)
	if err != nil {
		t.Errorf("failed to read from out")
	}

	require.Equal(
		t,
		"",
		string(out),
		"should not output anything",
	)

	require.NoError(t, mock.ExpectationsWereMet())
}

func testSecretSetCmdMissingProject(t *testing.T, mock sqlmock.Sqlmock, db *sql.DB) {
	cmdIn := bytes.NewReader([]byte{})
	cmdOut := bytes.NewBufferString("")
	errOut := bytes.NewBufferString("")

	err := cmd.Execute(
		[]*cobra.Command{secret.SecretCommand()},
		[]string{
			"secret",
			"set",
			"-e",
			"staging",
			"secret_key",
			"secret_value",
		},
		cmdIn,
		cmdOut,
		errOut,
		db,
	)

	require.Error(t, err)

	out, err := io.ReadAll(errOut)
	if err != nil {
		t.Error("failed to read from err out")
	}

	require.Equal(
		t,
		test.ErrorMsg(test.RequiredFlagsErrorMsg("project")),
		string(out),
	)
}

func testSecretSetCmdMissingEnvironment(t *testing.T, mock sqlmock.Sqlmock, db *sql.DB) {
	cmdIn := bytes.NewReader([]byte{})
	cmdOut := bytes.NewBufferString("")
	errOut := bytes.NewBufferString("")

	err := cmd.Execute(
		[]*cobra.Command{secret.SecretCommand()},
		[]string{
			"secret",
			"set",
			"-p",
			"my_cool_project",
			"secret_key",
			"secret_value",
		},
		cmdIn,
		cmdOut,
		errOut,
		db,
	)

	require.Error(t, err)

	out, err := io.ReadAll(errOut)
	if err != nil {
		t.Error("failed to read from err out")
	}

	require.Equal(t, test.ErrorMsg(test.RequiredFlagsErrorMsg("environment")), string(out))
}

func testSecretSetCmdTooFewArgs(t *testing.T, mock sqlmock.Sqlmock, db *sql.DB) {
	cmdIn := bytes.NewReader([]byte{})
	cmdOut := bytes.NewBufferString("")
	errOut := bytes.NewBufferString("")

	err := cmd.Execute(
		[]*cobra.Command{secret.SecretCommand()},
		[]string{
			"secret",
			"set",
			"-p",
			"my_cool_project",
			"secret_key",
		},
		cmdIn,
		cmdOut,
		errOut,
		db,
	)

	require.Error(t, err)

	out, err := io.ReadAll(errOut)
	if err != nil {
		t.Error("failed to read from err out")
	}

	require.Equal(
		t,
		test.ErrorMsg(test.IncorrectNumberOfArgsErrorMsg(2, 1)),
		string(out),
	)
}

func testSecretSetCmdTooManyArgs(t *testing.T, mock sqlmock.Sqlmock, db *sql.DB) {
	cmdIn := bytes.NewReader([]byte{})
	cmdOut := bytes.NewBufferString("")
	errOut := bytes.NewBufferString("")

	err := cmd.Execute(
		[]*cobra.Command{secret.SecretCommand()},
		[]string{
			"secret",
			"set",
			"-p",
			"my_cool_project",
			"secret_key",
			"secret_value",
			"foo",
		},
		cmdIn,
		cmdOut,
		errOut,
		db,
	)

	require.Error(t, err)

	out, err := io.ReadAll(errOut)
	if err != nil {
		t.Error("failed to read from err out")
	}

	require.Equal(
		t,
		test.ErrorMsg(test.IncorrectNumberOfArgsErrorMsg(2, 3)),
		string(out),
	)
}

func testSecretSetCmdDatabaseError(t *testing.T, mock sqlmock.Sqlmock, db *sql.DB) {
	cmdIn := bytes.NewReader([]byte{})
	cmdOut := bytes.NewBufferString("")
	errOut := bytes.NewBufferString("")

	query := `
		insert into secrets_ 
		(key_, value_, environment_id_) 
		values (
			$key,
			$value,
			(
				select e.id_ from 
					environments_ e
					inner join 
					projects_ p 
					on e.project_id_ = p.id_ 
					where p.name_ = $project 
					and e.name_ = $environment
			)
		)
	`

	mock.ExpectExec(regexp.QuoteMeta(query)).
		WithArgs(
			"my_cool_project",
			"staging",
			"secret_key",
			"secret_value",
		).
		WillReturnError(fmt.Errorf("database_error"))

	err := cmd.Execute(
		[]*cobra.Command{secret.SecretCommand()},
		[]string{
			"secret",
			"set",
			"-p",
			"my_cool_project",
			"-e",
			"staging",
			"secret_key",
			"secret_value",
		},
		cmdIn,
		cmdOut,
		errOut,
		db,
	)

	require.Error(t, err)

	out, err := io.ReadAll(errOut)
	if err != nil {
		t.Error("failed to read from err out")
	}

	require.Equal(t, test.ErrorMsg("database_error\n"), string(out))

	require.NoError(t, mock.ExpectationsWereMet())
}

func testSecretSetCmdValidationError(t *testing.T, mock sqlmock.Sqlmock, db *sql.DB) {
	cmdIn := bytes.NewReader([]byte{})
	cmdOut := bytes.NewBufferString("")
	errOut := bytes.NewBufferString("")

	err := cmd.Execute(
		[]*cobra.Command{secret.SecretCommand()},
		[]string{
			"secret",
			"set",
			"-p",
			"my_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_project",
			"-e",
			"sstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingtagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstaging",
			"ssecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keyecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_key",
			"ssecret_valuesecret_valuesecret_valuesecret_valuesecret_valuesecret_valuesecret_valuesecret_valuesecret_valuesecret_valuesecret_valuesecret_valuesecret_valuesecret_valuesecret_valuesecret_valuesecret_valuesecret_valuesecret_valuesecret_valuesecret_valuesecret_valuesecret_valuesecret_valuesecret_valuesecret_valuesecret_valuesecret_valuesecret_valuesecret_valuesecret_valuesecret_valuesecret_valuesecret_valuesecret_valuesecret_valuesecret_valuesecret_valuesecret_valuesecret_valuesecret_valuesecret_valuesecret_valuesecret_valuesecret_valuesecret_valuesecret_valuesecret_valuesecret_valuesecret_valuesecret_valuesecret_valuesecret_valuesecret_valuesecret_valuesecret_valuesecret_valuesecret_valuesecret_valuesecret_valueecret_valuesecret_valuesecret_valuesecret_valuesecret_valuesecret_valuesecret_valuesecret_valuesecret_valuesecret_valuesecret_valuesecret_valuesecret_valuesecret_valuesecret_valuesecret_valuesecret_valuesecret_valuesecret_valuesecret_valuesecret_valuesecret_valuesecret_valuesecret_valuesecret_valuesecret_valuesecret_valuesecret_valuesecret_valuesecret_value",
		},
		cmdIn,
		cmdOut,
		errOut,
		db,
	)

	require.Error(t, err)

	out, err := io.ReadAll(errOut)
	if err != nil {
		t.Error("failed to read from err out")
	}

	require.Equal(
		t,
		test.ErrorMsg(strings.Join(
			[]string{
				test.MaxLengthValidationErrorMsg("project name", 256),
				test.MaxLengthValidationErrorMsg("environment name", 256),
				test.MaxLengthValidationErrorMsg("secret key", 256),
				test.MaxLengthValidationErrorMsg("secret name", 256),
			},
			"",
		)),
		string(out),
	)
}

func testSecretGetCmdHappyPath(t *testing.T, mock sqlmock.Sqlmock, db *sql.DB) {
	cmdIn := bytes.NewReader([]byte{})
	cmdOut := bytes.NewBufferString("")

	query := `
		select s.id_, s.key_, s.value_, p.name_, e.name_
		from secrets_ s
		inner join
		environments_ e
		on s.environment_id_ = e.id_
		inner join
		projects_ p
		on p.id_ = e.project_id_
		where p.name_ = $project
		and e.name_ = $environment
		and s.key_ = $key
	`

	mock.ExpectQuery(regexp.QuoteMeta(query)).
		WithArgs(
			"my_cool_project",
			"staging",
			"secret_key",
		).
		WillReturnRows(mock.NewRows([]string{
			"id_",
			"key_",
			"value_",
			"project_name_",
			"environment_name_",
		}).AddRow(
			23,
			"secret_key",
			"secret_value",
			"my_cool_project",
			"staging",
		))

	err := cmd.Execute(
		[]*cobra.Command{secret.SecretCommand()},
		[]string{
			"secret",
			"get",
			"-p",
			"my_cool_project",
			"-e",
			"staging",
			"secret_key",
		},
		cmdIn,
		cmdOut,
		os.Stderr,
		db,
	)

	require.NoError(t, err)

	out, err := io.ReadAll(cmdOut)
	if err != nil {
		t.Errorf("failed to read from out")
	}

	require.Equal(
		t,
		"&{23 secret_key secret_value my_cool_project staging}",
		string(out),
		"should not output anything",
	)

	require.NoError(t, mock.ExpectationsWereMet())
}

func testSecretGetCmdMissingProject(t *testing.T, mock sqlmock.Sqlmock, db *sql.DB) {
	cmdIn := bytes.NewReader([]byte{})
	cmdOut := bytes.NewBufferString("")
	errOut := bytes.NewBufferString("")

	err := cmd.Execute(
		[]*cobra.Command{secret.SecretCommand()},
		[]string{
			"secret",
			"get",
			"-e",
			"staging",
			"secret_key",
		},
		cmdIn,
		cmdOut,
		errOut,
		db,
	)

	require.Error(t, err)

	out, err := io.ReadAll(errOut)
	if err != nil {
		t.Error("failed to read from err out")
	}

	require.Equal(t, test.ErrorMsg(test.RequiredFlagsErrorMsg("project")), string(out))
}

func testSecretGetCmdMissingEnvironment(t *testing.T, mock sqlmock.Sqlmock, db *sql.DB) {
	cmdIn := bytes.NewReader([]byte{})
	cmdOut := bytes.NewBufferString("")
	errOut := bytes.NewBufferString("")

	err := cmd.Execute(
		[]*cobra.Command{secret.SecretCommand()},
		[]string{
			"secret",
			"get",
			"-p",
			"my_cool_project",
			"secret_key",
		},
		cmdIn,
		cmdOut,
		errOut,
		db,
	)

	require.Error(t, err)

	out, err := io.ReadAll(errOut)
	if err != nil {
		t.Error("failed to read from err out")
	}

	require.Equal(t, test.ErrorMsg(test.RequiredFlagsErrorMsg("environment")), string(out))
}

func testSecretGetCmdMissingKey(t *testing.T, mock sqlmock.Sqlmock, db *sql.DB) {
	cmdIn := bytes.NewReader([]byte{})
	cmdOut := bytes.NewBufferString("")
	errOut := bytes.NewBufferString("")

	err := cmd.Execute(
		[]*cobra.Command{secret.SecretCommand()},
		[]string{
			"secret",
			"get",
			"-p",
			"my_cool_project",
			"-e",
			"staging",
		},
		cmdIn,
		cmdOut,
		errOut,
		db,
	)

	require.Error(t, err)

	out, err := io.ReadAll(errOut)
	if err != nil {
		t.Error("failed to read from err out")
	}

	require.Equal(t, test.ErrorMsg(test.IncorrectNumberOfArgsErrorMsg(1, 0)), string(out))
}

func testSecretGetCmdDatabaseError(t *testing.T, mock sqlmock.Sqlmock, db *sql.DB) {
	cmdIn := bytes.NewReader([]byte{})
	cmdOut := bytes.NewBufferString("")
	errOut := bytes.NewBufferString("")

	query := `
		select s.id_, s.key_, s.value_, p.name_, e.name_
		from secrets_ s
		inner join
		environments_ e
		on s.environment_id_ = e.id_
		inner join
		projects_ p
		on p.id_ = e.project_id_
		where p.name_ = $project
		and e.name_ = $environment
		and s.key_ = $key
	`

	mock.ExpectQuery(regexp.QuoteMeta(query)).
		WithArgs(
			"my_cool_project",
			"staging",
			"secret_key",
		).
		WillReturnError(fmt.Errorf("database_error"))

	err := cmd.Execute(
		[]*cobra.Command{secret.SecretCommand()},
		[]string{
			"secret",
			"get",
			"-p",
			"my_cool_project",
			"-e",
			"staging",
			"secret_key",
		},
		cmdIn,
		cmdOut,
		errOut,
		db,
	)

	require.Error(t, err)

	out, err := io.ReadAll(errOut)
	if err != nil {
		t.Error("failed to read from err out")
	}

	require.Equal(t, test.ErrorMsg("database_error\n"), string(out))

	require.NoError(t, mock.ExpectationsWereMet())
}

func testSecretGetCmdValidationError(t *testing.T, mock sqlmock.Sqlmock, db *sql.DB) {
	cmdIn := bytes.NewReader([]byte{})
	cmdOut := bytes.NewBufferString("")
	errOut := bytes.NewBufferString("")

	err := cmd.Execute(
		[]*cobra.Command{secret.SecretCommand()},
		[]string{
			"secret",
			"get",
			"-p",
			"my_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_project",
			"-e",
			"stagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstaging",
			"secret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_key",
		},
		cmdIn,
		cmdOut,
		errOut,
		db,
	)

	require.Error(t, err)

	out, err := io.ReadAll(errOut)
	if err != nil {
		t.Error("failed to read from err out")
	}

	require.Equal(
		t,
		test.ErrorMsg(strings.Join(
			[]string{
				test.MaxLengthValidationErrorMsg("project name", 256),
				test.MaxLengthValidationErrorMsg("environment name", 256),
				test.MaxLengthValidationErrorMsg("secret key", 256),
			}, "")),
		string(out),
	)
}

func testSecretListCmdHappyPath(t *testing.T, mock sqlmock.Sqlmock, db *sql.DB) {
	cmdIn := bytes.NewReader([]byte{})
	cmdOut := bytes.NewBufferString("")
	errOut := bytes.NewBufferString("")

	query := `
		select s.id_, s.key_, s.value_, p.name_, e.name_
		from secrets_ s
		inner join
		environments_ e
		on s.environment_id_ e.id_
		inner join
		projects_ p
		on e.project_id_ p.id_
		where p.name_ = $projectName
		and e.name_ = $environmentName
	`

	mock.
		ExpectQuery(regexp.QuoteMeta(query)).
		WithArgs("my_cool_project", "staging").
		WillReturnRows(
			sqlmock.NewRows([]string{
				"id_",
				"key_",
				"value_",
				"project_name_",
				"environment_name_",
			}).
				AddRow(1, "key_1", "value_1", "my_cool_project", "staging").
				AddRow(2, "key_2", "value_2", "my_cool_project", "staging"),
		)

	err := cmd.Execute(
		[]*cobra.Command{secret.SecretCommand()},
		[]string{"secret", "list", "-p", "my_cool_project", "-e", "staging"},
		cmdIn,
		cmdOut,
		errOut,
		db,
	)

	require.NoError(t, err)

	out, err := io.ReadAll(cmdOut)
	if err != nil {
		t.Errorf("unable to read from cmd out")
	}

	require.Equal(t, "1 key_1 value_1\n2 key_2 value_2\n", string(out))

	require.NoError(t, mock.ExpectationsWereMet())
}

func testSecretRemoveCmdHappyPath(t *testing.T, mock sqlmock.Sqlmock, db *sql.DB) {
	cmdIn := bytes.NewReader([]byte{})
	cmdOut := bytes.NewBufferString("")
	errOut := bytes.NewBufferString("")

	query := `
		delete from secrets_ 
		where id_ in (
			select s.id_ from secrets_ s
			inner join
			environments_ e
			on s.environment_id_ = e.id_
			inner join
			projects_ p
			on e.project_id_ = p.id_
			where p.name_ = $projectName
			and e.name_ = $environmentName
			and s.key_ = $key
		)
	`

	mock.ExpectExec(regexp.QuoteMeta(query)).WithArgs(
		"my_cool_project", "staging", "key_1",
	).WillReturnResult(sqlmock.NewResult(23, 1))

	err := cmd.Execute(
		[]*cobra.Command{secret.SecretCommand()},
		[]string{
			"secret",
			"remove",
			"-p",
			"my_cool_project",
			"-e",
			"staging",
			"key_1",
		},
		cmdIn,
		cmdOut,
		errOut,
		db,
	)

	require.NoError(t, err)

	require.NoError(t, mock.ExpectationsWereMet())
}

func testSecretListCmdDatabaseError(t *testing.T, mock sqlmock.Sqlmock, db *sql.DB) {
	cmdIn := bytes.NewReader([]byte{})
	cmdOut := bytes.NewBufferString("")
	errOut := bytes.NewBufferString("")

	query := `
		select s.id_, s.key_, s.value_, p.name_, e.name_
		from secrets_ s
		inner join
		environments_ e
		on s.environment_id_ e.id_
		inner join
		projects_ p
		on e.project_id_ p.id_
		where p.name_ = $projectName
		and e.name_ = $environmentName
	`

	mock.ExpectQuery(regexp.QuoteMeta(query)).WithArgs(
		"my_cool_project", "staging",
	).WillReturnError(errors.New("database_error"))

	err := cmd.Execute(
		[]*cobra.Command{secret.SecretCommand()},
		[]string{
			"secret",
			"list",
			"-p",
			"my_cool_project",
			"-e",
			"staging",
		},
		cmdIn,
		cmdOut,
		errOut,
		db,
	)

	require.Error(t, err)

	out, err := io.ReadAll(errOut)
	if err != nil {
		t.Error("failed to read from err out")
	}

	require.Equal(t, test.ErrorMsg("database_error\n"), string(out))

	require.NoError(t, mock.ExpectationsWereMet())
}

func testSecretListCmdMissingProject(t *testing.T, mock sqlmock.Sqlmock, db *sql.DB) {
	cmdIn := bytes.NewReader([]byte{})
	cmdOut := bytes.NewBufferString("")
	errOut := bytes.NewBufferString("")

	err := cmd.Execute(
		[]*cobra.Command{secret.SecretCommand()},
		[]string{
			"secret",
			"list",
			"-e",
			"staging",
		},
		cmdIn,
		cmdOut,
		errOut,
		db,
	)

	require.Error(t, err)

	out, err := io.ReadAll(errOut)
	if err != nil {
		t.Error("failed to read from err out")
	}

	require.Equal(t, test.ErrorMsg(test.RequiredFlagsErrorMsg("project")), string(out))
}

func testSecretListCmdMissingEnvironment(t *testing.T, mock sqlmock.Sqlmock, db *sql.DB) {
	cmdIn := bytes.NewReader([]byte{})
	cmdOut := bytes.NewBufferString("")
	errOut := bytes.NewBufferString("")

	err := cmd.Execute(
		[]*cobra.Command{secret.SecretCommand()},
		[]string{
			"secret",
			"list",
			"-p",
			"my_cool_project",
		},
		cmdIn,
		cmdOut,
		errOut,
		db,
	)

	require.Error(t, err)

	out, err := io.ReadAll(errOut)
	if err != nil {
		t.Error("failed to read from err out")
	}

	require.Equal(t, test.ErrorMsg(test.RequiredFlagsErrorMsg("environment")), string(out))
}

func testSecretListCmdValidationError(t *testing.T, mock sqlmock.Sqlmock, db *sql.DB) {
	cmdIn := bytes.NewReader([]byte{})
	cmdOut := bytes.NewBufferString("")
	errOut := bytes.NewBufferString("")

	err := cmd.Execute(
		[]*cobra.Command{secret.SecretCommand()},
		[]string{
			"secret",
			"list",
			"-p",
			"my_cool_project_my_cool_project_my_cool_project_my_cool_project_my_cool_project_my_cool_project_my_cool_project_my_cool_project_my_cool_project_my_cool_project_my_cool_project_my_cool_project_my_cool_project_my_cool_project_my_cool_project_my_cool_project_my_cool_project_",
			"-e",
			"staging",
		},
		cmdIn,
		cmdOut,
		errOut,
		db,
	)

	require.Error(t, err)

	out, err := io.ReadAll(errOut)
	if err != nil {
		t.Error("failed to read from err out")
	}

	require.Equal(

		t,
		test.ErrorMsg(test.MaxLengthValidationErrorMsg("project name", 256)),
		string(out),
	)

	err = cmd.Execute(
		[]*cobra.Command{secret.SecretCommand()},
		[]string{
			"secret",
			"list",
			"-p",
			"my_cool_project",
			"-e",
			"staging_staging_staging_staging_staging_staging_staging_staging_staging_staging_staging_staging_staging_staging_staging_staging_staging_staging_staging_staging_staging_staging_staging_staging_staging_staging_staging_staging_staging_staging_staging_staging_staging",
		},
		cmdIn,
		cmdOut,
		errOut,
		db,
	)

	require.Error(t, err)

	out, err = io.ReadAll(errOut)
	if err != nil {
		t.Error("failed to read from err out")
	}

	require.Equal(
		t,
		test.ErrorMsg(test.MaxLengthValidationErrorMsg("environment name", 256)),
		string(out),
	)
}

func testSecretRemoveCmdDatabaseError(t *testing.T, mock sqlmock.Sqlmock, db *sql.DB) {
	cmdIn := bytes.NewReader([]byte{})
	cmdOut := bytes.NewBufferString("")
	errOut := bytes.NewBufferString("")

	query := `
		delete from secrets_ 
		where id_ in (
			select s.id_ from secrets_ s
			inner join
			environments_ e
			on s.environment_id_ = e.id_
			inner join
			projects_ p
			on e.project_id_ = p.id_
			where p.name_ = $projectName
			and e.name_ = $environmentName
			and s.key_ = $key
		)
	`

	mock.ExpectExec(regexp.QuoteMeta(query)).WithArgs(
		"my_cool_project", "staging", "key_1",
	).WillReturnError(errors.New("database_error"))

	err := cmd.Execute(
		[]*cobra.Command{secret.SecretCommand()},
		[]string{
			"secret",
			"remove",
			"-p",
			"my_cool_project",
			"-e",
			"staging",
			"key_1",
		},
		cmdIn,
		cmdOut,
		errOut,
		db,
	)

	require.Error(t, err)

	out, err := io.ReadAll(errOut)
	if err != nil {
		t.Error("failed to read from err out")
	}

	require.Equal(t, test.ErrorMsg("database_error\n"), string(out))

	require.NoError(t, mock.ExpectationsWereMet())
}

func testSecretRemoveCmdMissingProject(t *testing.T, mock sqlmock.Sqlmock, db *sql.DB) {
	cmdIn := bytes.NewReader([]byte{})
	cmdOut := bytes.NewBufferString("")
	errOut := bytes.NewBufferString("")

	err := cmd.Execute(
		[]*cobra.Command{secret.SecretCommand()},
		[]string{
			"secret",
			"remove",
			"-e",
			"staging",
			"key_1",
		},
		cmdIn,
		cmdOut,
		errOut,
		db,
	)

	require.Error(t, err)

	out, err := io.ReadAll(errOut)
	if err != nil {
		t.Error("failed to read from err out")
	}

	require.Equal(t, test.ErrorMsg(test.RequiredFlagsErrorMsg("project")), string(out))
}

func testSecretRemoveCmdMissingEnvironment(t *testing.T, mock sqlmock.Sqlmock, db *sql.DB) {
	cmdIn := bytes.NewReader([]byte{})
	cmdOut := bytes.NewBufferString("")
	errOut := bytes.NewBufferString("")

	err := cmd.Execute(
		[]*cobra.Command{secret.SecretCommand()},
		[]string{
			"secret",
			"remove",
			"-p",
			"my_cool_project",
			"key_1",
		},
		cmdIn,
		cmdOut,
		errOut,
		db,
	)

	require.Error(t, err)

	out, err := io.ReadAll(errOut)
	if err != nil {
		t.Error("failed to read from err out")
	}

	require.Equal(t, test.ErrorMsg(test.RequiredFlagsErrorMsg("environment")), string(out))
}

func testSecretRemoveCmdMissingKey(t *testing.T, mock sqlmock.Sqlmock, db *sql.DB) {
	cmdIn := bytes.NewReader([]byte{})
	cmdOut := bytes.NewBufferString("")
	errOut := bytes.NewBufferString("")

	err := cmd.Execute(
		[]*cobra.Command{secret.SecretCommand()},
		[]string{
			"secret",
			"remove",
			"-p",
			"my_cool_project",
			"-e",
			"staging",
		},
		cmdIn,
		cmdOut,
		errOut,
		db,
	)

	require.Error(t, err)

	out, err := io.ReadAll(errOut)
	if err != nil {
		t.Error("failed to read from err out")
	}

	require.Equal(t, test.ErrorMsg(test.IncorrectNumberOfArgsErrorMsg(1, 0)), string(out))
}

func testSecretRemoveCmdValidationError(t *testing.T, mock sqlmock.Sqlmock, db *sql.DB) {
	cmdIn := bytes.NewReader([]byte{})
	cmdOut := bytes.NewBufferString("")
	errOut := bytes.NewBufferString("")

	err := cmd.Execute(
		[]*cobra.Command{secret.SecretCommand()},
		[]string{
			"secret",
			"remove",
			"-p",
			"my_cool_project_my_cool_project_my_cool_project_my_cool_project_my_cool_project_my_cool_project_my_cool_project_my_cool_project_my_cool_project_my_cool_project_my_cool_project_my_cool_project_my_cool_project_my_cool_project_my_cool_project_my_cool_project_my_cool_project_my_cool_project_",
			"-e",
			"staging",
			"key_1",
		},
		cmdIn,
		cmdOut,
		errOut,
		db,
	)

	require.Error(t, err)

	out, err := io.ReadAll(errOut)
	if err != nil {
		t.Error("failed to read from err out")
	}

	require.Equal(t, test.ErrorMsg(test.MaxLengthValidationErrorMsg("project name", 256)), string(out))

	err = cmd.Execute(
		[]*cobra.Command{secret.SecretCommand()},
		[]string{
			"secret",
			"remove",
			"-p",
			"my_cool_project",
			"-e",
			"staging_staging_staging_staging_staging_staging_staging_staging_staging_staging_staging_staging_staging_staging_staging_staging_staging_staging_staging_staging_staging_staging_staging_staging_staging_staging_staging_staging_staging_staging_staging_staging_staging_staging_staging_",
			"key_1",
		},
		cmdIn,
		cmdOut,
		errOut,
		db,
	)

	require.Error(t, err)

	out, err = io.ReadAll(errOut)
	if err != nil {
		t.Error("failed to read from err out")
	}

	require.Equal(t, test.ErrorMsg(test.MaxLengthValidationErrorMsg("environment name", 256)), string(out))
}
