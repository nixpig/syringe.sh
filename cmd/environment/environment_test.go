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
	"github.com/nixpig/syringe.sh/server/test"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

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

		"test environment list command happy path":           testEnvironmentListCmdHappyPath,
		"test environment list command database error":       testEnvironmentListCmdDatabaseError,
		"test environment list command validation errors":    testEnvironmentListCmdValidationError,
		"test environment list command missing project flag": testEnvironmentListCmdMissingProjectFlag,
		"test environment list command with too many args":   testEnvironmentListCmdWithTooManyArgs,
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
		test.EnvironmentAddedSuccessMsg("staging", "my_cool_project"),
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
		test.EnvironmentRemovedSuccessMsg("staging", "my_cool_project"),
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
		test.RequiredFlagsErrorMsg("project"),
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
		test.EnvironmentRenamedSuccessMsg("staging", "prod", "my_cool_project"),
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
		test.ErrorMsg("database_error"),
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
		test.MaxLengthValidationErrorMsg("environment name", 256),
		string(out),
	)
}

func testEnvironmentRenameCmdMissingProjectFlag(t *testing.T, mock sqlmock.Sqlmock, db *sql.DB) {
	cmdIn := bytes.NewReader([]byte{})
	cmdOut := bytes.NewBufferString("")
	errOut := bytes.NewBufferString("")

	err := cmd.Execute(
		[]*cobra.Command{environment.EnvironmentCommand()},
		[]string{
			"environment",
			"rename",
			"current_name",
			"new_name",
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
		test.RequiredFlagsErrorMsg("project"),
		string(out),
	)
}

func testEnvironmentRenameCmdWithNoArgs(t *testing.T, mock sqlmock.Sqlmock, db *sql.DB) {
	cmdIn := bytes.NewReader([]byte{})
	cmdOut := bytes.NewBufferString("")
	errOut := bytes.NewBufferString("")

	err := cmd.Execute(
		[]*cobra.Command{environment.EnvironmentCommand()},
		[]string{
			"environment",
			"rename",
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
		t.Errorf("failed to read from out")
	}

	require.Equal(
		t,
		test.IncorrectNumberOfArgsErrorMsg(2, 0),
		string(out),
	)
}

func testEnvironmentRenameCmdWithTooManyArgs(t *testing.T, mock sqlmock.Sqlmock, db *sql.DB) {
	cmdIn := bytes.NewReader([]byte{})
	cmdOut := bytes.NewBufferString("")
	errOut := bytes.NewBufferString("")

	err := cmd.Execute(
		[]*cobra.Command{environment.EnvironmentCommand()},
		[]string{
			"environment",
			"rename",
			"-p",
			"my_cool_project",
			"foo",
			"bar",
			"baz",
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
		test.IncorrectNumberOfArgsErrorMsg(2, 3),
		string(out),
	)
}

func testEnvironmentListCmdHappyPath(t *testing.T, mock sqlmock.Sqlmock, db *sql.DB) {
	cmdIn := bytes.NewReader([]byte{})
	cmdOut := bytes.NewBufferString("")

	query := `
		select e.name_ from environments_ e
		inner join projects_ p
		on e.project_id_ = p.id_ 
		where p.name_ = $projectName
	`

	mock.
		ExpectQuery(regexp.QuoteMeta(query)).
		WithArgs("my_cool_project").
		WillReturnRows(
			sqlmock.
				NewRows([]string{"name_"}).
				AddRow("dev").
				AddRow("staging").
				AddRow("prod"),
		)

	err := cmd.Execute(
		[]*cobra.Command{environment.EnvironmentCommand()},
		[]string{
			"environment",
			"list",
			"-p",
			"my_cool_project",
		},
		cmdIn,
		cmdOut,
		os.Stderr,
		db,
	)

	require.NoError(t, err)

	require.NoError(t, mock.ExpectationsWereMet())
}

func testEnvironmentListCmdDatabaseError(t *testing.T, mock sqlmock.Sqlmock, db *sql.DB) {
	cmdIn := bytes.NewReader([]byte{})
	cmdOut := bytes.NewBufferString("")
	errOut := bytes.NewBufferString("")

	query := `
		select e.name_ from environments_ e
		inner join projects_ p
		on e.project_id_ = p.id_ 
		where p.name_ = $projectName
	`

	mock.
		ExpectQuery(regexp.QuoteMeta(query)).
		WithArgs("my_cool_project").
		WillReturnError(
			errors.New("database_error"),
		)

	err := cmd.Execute(
		[]*cobra.Command{environment.EnvironmentCommand()},
		[]string{
			"environment",
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

	require.Equal(t, test.ErrorMsg("database_error"), string(out))

	require.NoError(t, mock.ExpectationsWereMet())
}

func testEnvironmentListCmdValidationError(t *testing.T, mock sqlmock.Sqlmock, db *sql.DB) {
	cmdIn := bytes.NewReader([]byte{})
	cmdOut := bytes.NewBufferString("")
	errOut := bytes.NewBufferString("")

	err := cmd.Execute(
		[]*cobra.Command{environment.EnvironmentCommand()},
		[]string{
			"environment",
			"list",
			"-p",
			"my_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_project",
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

	require.Equal(t, test.MaxLengthValidationErrorMsg("project name", 256), string(out))
}

func testEnvironmentListCmdMissingProjectFlag(t *testing.T, mock sqlmock.Sqlmock, db *sql.DB) {
	cmdIn := bytes.NewReader([]byte{})
	cmdOut := bytes.NewBufferString("")
	errOut := bytes.NewBufferString("")

	err := cmd.Execute(
		[]*cobra.Command{environment.EnvironmentCommand()},
		[]string{
			"environment",
			"list",
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

	require.Equal(t, test.RequiredFlagsErrorMsg("project"), string(out))
}

func testEnvironmentListCmdWithTooManyArgs(t *testing.T, mock sqlmock.Sqlmock, db *sql.DB) {
	cmdIn := bytes.NewReader([]byte{})
	cmdOut := bytes.NewBufferString("")
	errOut := bytes.NewBufferString("")

	err := cmd.Execute(
		[]*cobra.Command{environment.EnvironmentCommand()},
		[]string{
			"environment",
			"list",
			"-p",
			"my_cool_project",
			"foobarbaz",
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

	require.Equal(t, test.IncorrectNumberOfArgsErrorMsg(0, 1), string(out))
}
