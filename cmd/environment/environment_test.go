package environment_test

import (
	"bytes"
	"database/sql"
	"fmt"
	"io"
	"os"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/nixpig/syringe.sh/server/cmd"
	"github.com/nixpig/syringe.sh/server/cmd/environment"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

func addedSuccessMsg(environment, project string) string {
	return fmt.Sprintf("Environment '%s' added to project '%s'\n", environment, project)
}

func removedSuccessMsg(environment, project string) string {
	return fmt.Sprintf("Environment '%s' removed from project '%s'\n", environment, project)
}

func TestEnvironmentCmd(t *testing.T) {
	scenarios := map[string]func(t *testing.T, mock sqlmock.Sqlmock, db *sql.DB){
		"test environment add command happy path":           testEnvironmentAddCmdHappyPath,
		"test environment add command missing project flag": testEnvironmentAddCmdMissingProjectFlag,
		"test environment add command with no args":         testEnvironmentAddCmdWithNoArgs,
		"test environment add command with too many args":   testEnvironmentAddCmdWithTooManyArgs,
		"test environment add command database error":       testEnvironmentAddCmdDatabaseError,
		"test environment add command validation error":     testEnvironmentAddCmdValidationError,

		"test environment remove command happy path":           testEnvironmentRemoveCmdHappyPath,
		"test environment remove command missing project flag": testEnvironmentRemoveCmdMissingProjectFlag,
		"test environment remove command with no args":         testEnvironmentRemoveCmdWithNoArgs,
		"test environment remove command with too many args":   testEnvironmentRemoveCmdWithTooManyArgs,
		"test environment remove command database error":       testEnvironmentRemoveCmdDatabaseError,
		"test environment remove command validation error":     testEnvironmentRemoveCmdValidationError,
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

func testEnvironmentAddCmdHappyPath(t *testing.T, mock sqlmock.Sqlmock, db *sql.DB) {
	cmdIn := bytes.NewReader([]byte{})
	cmdOut := bytes.NewBufferString("")

	mock.ExpectExec(regexp.QuoteMeta(`
		insert into environments_ (name_, project_id_) values (
			$name,
			(select id_ from projects_ where name_ = $projectName)
		)
	`)).
		WithArgs("staging", "my_cool_project").
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := cmd.Execute(
		[]*cobra.Command{environment.EnvironmentCommand()},
		[]string{
			"environment",
			"add",
			"-p",
			"my_cool_project",
			"staging",
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
		addedSuccessMsg("staging", "my_cool_project"),
		string(out),
	)

	require.NoError(t, mock.ExpectationsWereMet())
}

func testEnvironmentAddCmdMissingProjectFlag(t *testing.T, mock sqlmock.Sqlmock, db *sql.DB) {
	cmdIn := bytes.NewReader([]byte{})
	cmdOut := bytes.NewBufferString("")

	err := cmd.Execute(
		[]*cobra.Command{environment.EnvironmentCommand()},
		[]string{
			"environment",
			"add",
			"staging",
		},
		cmdIn,
		cmdOut,
		os.Stderr,
		db,
	)

	require.Error(t, err)
}

func testEnvironmentAddCmdWithNoArgs(t *testing.T, mock sqlmock.Sqlmock, db *sql.DB) {
	cmdIn := bytes.NewReader([]byte{})
	cmdOut := bytes.NewBufferString("")

	err := cmd.Execute(
		[]*cobra.Command{environment.EnvironmentCommand()},
		[]string{
			"environment",
			"add",
			"-p",
			"my_cool_project",
		},
		cmdIn,
		cmdOut,
		os.Stderr,
		db,
	)

	require.Error(t, err)
}

func testEnvironmentAddCmdWithTooManyArgs(t *testing.T, mock sqlmock.Sqlmock, db *sql.DB) {
	cmdIn := bytes.NewReader([]byte{})
	cmdOut := bytes.NewBufferString("")

	err := cmd.Execute(
		[]*cobra.Command{environment.EnvironmentCommand()},
		[]string{
			"environment",
			"add",
			"-p",
			"my_cool_project",
			"foo",
			"bar",
		},
		cmdIn,
		cmdOut,
		os.Stderr,
		db,
	)

	require.Error(t, err)
}

func testEnvironmentAddCmdDatabaseError(t *testing.T, mock sqlmock.Sqlmock, db *sql.DB) {
	cmdIn := bytes.NewReader([]byte{})
	cmdOut := bytes.NewBufferString("")

	mock.ExpectExec(regexp.QuoteMeta(`
		insert into environments_ (name_, project_id_) values (
			$name,
			(select id_ from projects_ where name_ = $projectName)
		)
	`)).
		WithArgs("staging", "my_cool_project").
		WillReturnError(fmt.Errorf("database_error"))

	err := cmd.Execute(
		[]*cobra.Command{environment.EnvironmentCommand()},
		[]string{
			"environment",
			"add",
			"-p",
			"my_cool_project",
			"staging",
		},
		cmdIn,
		cmdOut,
		os.Stderr,
		db,
	)

	require.Error(t, err)

	require.NoError(t, mock.ExpectationsWereMet())
}

func testEnvironmentAddCmdValidationError(t *testing.T, mock sqlmock.Sqlmock, db *sql.DB) {
	cmdIn := bytes.NewReader([]byte{})
	cmdOut := bytes.NewBufferString("")

	var err error

	err = cmd.Execute(
		[]*cobra.Command{environment.EnvironmentCommand()},
		[]string{
			"environment",
			"add",
			"-p",
			"my_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_project",
			"stagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstaging",
		},
		cmdIn,
		cmdOut,
		os.Stderr,
		db,
	)

	require.Error(t, err)

	err = cmd.Execute(
		[]*cobra.Command{environment.EnvironmentCommand()},
		[]string{
			"environment",
			"add",
			"-p",
			"my_cool_project",
			"stagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstaging",
		},
		cmdIn,
		cmdOut,
		os.Stderr,
		db,
	)

	require.Error(t, err)

	err = cmd.Execute(
		[]*cobra.Command{environment.EnvironmentCommand()},
		[]string{
			"environment",
			"add",
			"-p",
			"my_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_project",
			"staging",
		},
		cmdIn,
		cmdOut,
		os.Stderr,
		db,
	)

	require.Error(t, err)

	err = cmd.Execute(
		[]*cobra.Command{environment.EnvironmentCommand()},
		[]string{
			"environment",
			"add",
			"-p",
			"",
			"staging",
		},
		cmdIn,
		cmdOut,
		os.Stderr,
		db,
	)

	require.Error(t, err)

	err = cmd.Execute(
		[]*cobra.Command{environment.EnvironmentCommand()},
		[]string{
			"environment",
			"add",
			"-p",
			"my_cool_project",
			"",
		},
		cmdIn,
		cmdOut,
		os.Stderr,
		db,
	)

	require.Error(t, err)
}

func testEnvironmentRemoveCmdHappyPath(t *testing.T, mock sqlmock.Sqlmock, db *sql.DB) {
	cmdIn := bytes.NewReader([]byte{})
	cmdOut := bytes.NewBufferString("")

	mock.ExpectExec(regexp.QuoteMeta(`
		delete from environments_ 
		where id_ in (
			select e.id_ from environments_ e
			inner join
			projects_ p
			on e.project_id_ = p.id_
			where p.name_ = $projectName
			and e.name_ = $name
		)
	`)).
		WithArgs("staging", "my_cool_project").
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := cmd.Execute(
		[]*cobra.Command{environment.EnvironmentCommand()},
		[]string{
			"environment",
			"remove",
			"-p",
			"my_cool_project",
			"staging",
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
		removedSuccessMsg("staging", "my_cool_project"),
		string(out),
	)

	require.NoError(t, mock.ExpectationsWereMet())
}

func testEnvironmentRemoveCmdDatabaseError(t *testing.T, mock sqlmock.Sqlmock, db *sql.DB) {
	cmdIn := bytes.NewReader([]byte{})
	cmdOut := bytes.NewBufferString("")

	mock.ExpectExec(regexp.QuoteMeta(`
		delete from environments_ 
		where id_ in (
			select e.id_ from environments_ e
			inner join
			projects_ p
			on e.project_id_ = p.id_
			where p.name_ = $projectName
			and e.name_ = $name
		)
	`)).
		WithArgs("staging", "my_cool_project").
		WillReturnError(fmt.Errorf("database_error"))

	err := cmd.Execute(
		[]*cobra.Command{environment.EnvironmentCommand()},
		[]string{
			"environment",
			"remove",
			"-p",
			"my_cool_project",
			"staging",
		},
		cmdIn,
		cmdOut,
		os.Stderr,
		db,
	)

	require.Error(t, err)

	require.NoError(t, mock.ExpectationsWereMet())
}

func testEnvironmentRemoveCmdMissingProjectFlag(t *testing.T, mock sqlmock.Sqlmock, db *sql.DB) {
	cmdIn := bytes.NewReader([]byte{})
	cmdOut := bytes.NewBufferString("")
	errOut := bytes.NewBufferString("")

	err := cmd.Execute(
		[]*cobra.Command{environment.EnvironmentCommand()},
		[]string{
			"environment",
			"remove",
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
		t.Errorf("failed to read from err out")
	}

	require.Equal(
		t,
		"Error: required flag(s) \"project\" not set\n",
		string(out),
	)

	require.NoError(t, mock.ExpectationsWereMet())
}

func testEnvironmentRemoveCmdWithNoArgs(t *testing.T, mock sqlmock.Sqlmock, db *sql.DB) {
	cmdIn := bytes.NewReader([]byte{})
	cmdOut := bytes.NewBufferString("")

	err := cmd.Execute(
		[]*cobra.Command{environment.EnvironmentCommand()},
		[]string{
			"environment",
			"remove",
			"-p",
			"my_cool_project",
		},
		cmdIn,
		cmdOut,
		os.Stderr,
		db,
	)

	require.Error(t, err)

	require.NoError(t, mock.ExpectationsWereMet())
}

func testEnvironmentRemoveCmdWithTooManyArgs(t *testing.T, mock sqlmock.Sqlmock, db *sql.DB) {
	cmdIn := bytes.NewReader([]byte{})
	cmdOut := bytes.NewBufferString("")

	err := cmd.Execute(
		[]*cobra.Command{environment.EnvironmentCommand()},
		[]string{
			"environment",
			"remove",
			"-p",
			"my_cool_project",
			"foo",
			"bar",
		},
		cmdIn,
		cmdOut,
		os.Stderr,
		db,
	)

	require.Error(t, err)

	require.NoError(t, mock.ExpectationsWereMet())
}

func testEnvironmentRemoveCmdValidationError(t *testing.T, mock sqlmock.Sqlmock, db *sql.DB) {
	cmdIn := bytes.NewReader([]byte{})
	cmdOut := bytes.NewBufferString("")

	err := cmd.Execute(
		[]*cobra.Command{environment.EnvironmentCommand()},
		[]string{
			"environment",
			"remove",
			"-p",
			"my_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_project",
			"stagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstaging",
		},
		cmdIn,
		cmdOut,
		os.Stderr,
		db,
	)

	require.Error(t, err)

	require.NoError(t, mock.ExpectationsWereMet())
}
