package environment_test

import (
	"bytes"
	"database/sql"
	"errors"
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

func renamedSuccessMsg(name, newName, project string) string {
	return fmt.Sprintf("Environment '%s' renamed to '%s' in project '%s'\n", name, newName, project)
}

func maxLengthValidationErrorMsg(field string, length int) string {
	return fmt.Sprintf("Error: \"%s\" exceeds max length of %d characters\n", field, length)
}

func errorMsg(e string) string {
	return fmt.Sprintf("Error: %s\n", e)
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

		"test environment rename command happy path":           testEnvironmentRenameCmdHappyPath,
		"test environment rename command database error":       testEnvironmentRenameCmdDatabaseError,
		"test environment rename command validation errors":    testEnvironmentRenameCmdValidationError,
		"test environment rename command missing project flag": testEnvironmentRenameCmdMissingProjectFlag,
		"test environment rename command with no args":         testEnvironmentRenameCmdWithNoArgs,
		"test environment rename command with too many args":   testEnvironmentRenameCmdWithTooManyArgs,
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
			"stagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstaging",
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

func testEnvironmentRenameCmdHappyPath(t *testing.T, mock sqlmock.Sqlmock, db *sql.DB) {
	cmdIn := bytes.NewReader([]byte{})
	cmdOut := bytes.NewBufferString("")

	mock.ExpectExec(regexp.QuoteMeta(`
		update environments_ set name_ = $newName
		where name_ = $originalName 
		and id_ in (
			select e.id_ from environments_ e
			inner join
			projects_ p
			on e.project_id_ = p.id_
			where p.name_ = $projectName
			and e.name_ = $originalName
		)
	`)).WithArgs(
		"staging",
		"prod",
		"my_cool_project",
	).WillReturnResult(sqlmock.NewResult(23, 1))

	err := cmd.Execute(
		[]*cobra.Command{environment.EnvironmentCommand()},
		[]string{
			"environment",
			"rename",
			"-p",
			"my_cool_project",
			"staging",
			"prod",
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
		renamedSuccessMsg("staging", "prod", "my_cool_project"),
		string(out),
	)

	require.NoError(t, mock.ExpectationsWereMet())
}

func testEnvironmentRenameCmdDatabaseError(t *testing.T, mock sqlmock.Sqlmock, db *sql.DB) {
	cmdIn := bytes.NewReader([]byte{})
	cmdOut := bytes.NewBufferString("")
	errOut := bytes.NewBufferString("")

	mock.ExpectExec(regexp.QuoteMeta(`
		update environments_ set name_ = $newName
		where name_ = $originalName 
		and id_ in (
			select e.id_ from environments_ e
			inner join
			projects_ p
			on e.project_id_ = p.id_
			where p.name_ = $projectName
			and e.name_ = $originalName
		)
	`)).WithArgs(
		"staging",
		"prod",
		"my_cool_project",
	).WillReturnError(errors.New("database_error"))

	err := cmd.Execute(
		[]*cobra.Command{environment.EnvironmentCommand()},
		[]string{
			"environment",
			"rename",
			"-p",
			"my_cool_project",
			"staging",
			"prod",
		},
		cmdIn,
		cmdOut,
		errOut,
		db,
	)

	require.EqualError(t, err, "database_error")

	out, err := io.ReadAll(errOut)
	if err != nil {
		t.Errorf("failed to read from out")
	}

	require.Equal(
		t,
		errorMsg("database_error"),
		string(out),
	)

	require.NoError(t, mock.ExpectationsWereMet())
}

func testEnvironmentRenameCmdValidationError(t *testing.T, mock sqlmock.Sqlmock, db *sql.DB) {
	cmdIn := bytes.NewReader([]byte{})
	cmdOut := bytes.NewBufferString("")
	errOut := bytes.NewBufferString("")

	err := cmd.Execute(
		[]*cobra.Command{environment.EnvironmentCommand()},
		[]string{
			"environment",
			"rename",
			"-p",
			"my_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_project",
			"current_name_current_name_current_name_current_name_current_name_current_name_current_name_current_name_current_name_current_name_current_name_current_name_current_name_current_name_current_name_current_name_current_name_current_name_current_name_current_name_current_name_current_name_current_name_current_name_current_name_",
			"new_name_new_name_new_name_new_name_new_name_new_name_new_name_new_name_new_name_new_name_new_name_new_name_new_name_new_name_new_name_new_name_new_name_new_name_new_name_new_name_new_name_new_name_new_name_new_name_new_name_new_name_new_name_new_name_new_name_new_name_new_name_new_name_new_name_new_name_new_name_new_name_",
		},
		cmdIn,
		cmdOut,
		errOut,
		db,
	)

	require.Error(t, err)

	out, err := io.ReadAll(errOut)
	if err != nil {
		t.Errorf("failed to read from out")
	}

	require.Equal(
		t,
		maxLengthValidationErrorMsg("environment name", 256),
		string(out),
	)
}

func testEnvironmentRenameCmdMissingProjectFlag(t *testing.T, mock sqlmock.Sqlmock, db *sql.DB) {}

func testEnvironmentRenameCmdWithNoArgs(t *testing.T, mock sqlmock.Sqlmock, db *sql.DB) {}

func testEnvironmentRenameCmdWithTooManyArgs(t *testing.T, mock sqlmock.Sqlmock, db *sql.DB) {}
