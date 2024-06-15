package secret_test

import (
	"bytes"
	"database/sql"
	"errors"
	"io"
	"os"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/nixpig/syringe.sh/server/cmd"
	"github.com/nixpig/syringe.sh/server/cmd/secret"
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
		os.Stderr,
		db,
	)

	require.Error(t, err)
}

func testSecretSetCmdMissingEnvironment(t *testing.T, mock sqlmock.Sqlmock, db *sql.DB) {
	cmdIn := bytes.NewReader([]byte{})
	cmdOut := bytes.NewBufferString("")

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
		os.Stderr,
		db,
	)

	require.Error(t, err)
}

func testSecretSetCmdTooFewArgs(t *testing.T, mock sqlmock.Sqlmock, db *sql.DB) {
	cmdIn := bytes.NewReader([]byte{})
	cmdOut := bytes.NewBufferString("")

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
		os.Stderr,
		db,
	)

	require.Error(t, err)
}

func testSecretSetCmdTooManyArgs(t *testing.T, mock sqlmock.Sqlmock, db *sql.DB) {
	cmdIn := bytes.NewReader([]byte{})
	cmdOut := bytes.NewBufferString("")

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
		os.Stderr,
		db,
	)

	require.Error(t, err)
}

func testSecretSetCmdDatabaseError(t *testing.T, mock sqlmock.Sqlmock, db *sql.DB) {
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
		WillReturnError(errors.New("database_error"))

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

	require.Error(t, err)

	require.NoError(t, mock.ExpectationsWereMet())
}

func testSecretSetCmdValidationError(t *testing.T, mock sqlmock.Sqlmock, db *sql.DB) {
	cmdIn := bytes.NewReader([]byte{})
	cmdOut := bytes.NewBufferString("")

	err := cmd.Execute(
		[]*cobra.Command{secret.SecretCommand()},
		[]string{
			"secret",
			"set",
			"-p",
			"my_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_project",
			"-e",
			"stagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstaging",
			"secret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_key",
			"secret_valuesecret_valuesecret_valuesecret_valuesecret_valuesecret_valuesecret_valuesecret_valuesecret_valuesecret_valuesecret_valuesecret_valuesecret_valuesecret_valuesecret_valuesecret_valuesecret_valuesecret_valuesecret_valuesecret_valuesecret_valuesecret_valuesecret_valuesecret_valuesecret_valuesecret_valuesecret_valuesecret_valuesecret_valuesecret_value",
		},
		cmdIn,
		cmdOut,
		os.Stderr,
		db,
	)

	require.Error(t, err)
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
		os.Stderr,
		db,
	)

	require.Error(t, err)
}

func testSecretGetCmdMissingEnvironment(t *testing.T, mock sqlmock.Sqlmock, db *sql.DB) {
	cmdIn := bytes.NewReader([]byte{})
	cmdOut := bytes.NewBufferString("")

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
		os.Stderr,
		db,
	)

	require.Error(t, err)
}

func testSecretGetCmdMissingKey(t *testing.T, mock sqlmock.Sqlmock, db *sql.DB) {
	cmdIn := bytes.NewReader([]byte{})
	cmdOut := bytes.NewBufferString("")

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
		os.Stderr,
		db,
	)

	require.Error(t, err)
}

func testSecretGetCmdDatabaseError(t *testing.T, mock sqlmock.Sqlmock, db *sql.DB) {
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
		WillReturnError(errors.New("database_error"))

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

	require.Error(t, err)

	require.NoError(t, mock.ExpectationsWereMet())
}

func testSecretGetCmdValidationError(t *testing.T, mock sqlmock.Sqlmock, db *sql.DB) {
	cmdIn := bytes.NewReader([]byte{})
	cmdOut := bytes.NewBufferString("")

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
		os.Stderr,
		db,
	)

	require.Error(t, err)

	require.NoError(t, mock.ExpectationsWereMet())
}
