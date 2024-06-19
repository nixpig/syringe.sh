package project_test

import (
	"bytes"
	"database/sql"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/nixpig/syringe.sh/server/cmd"
	"github.com/nixpig/syringe.sh/server/cmd/project"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

func incorrectNumberOfArgsErrorMsg(accepts, received int) string {
	return fmt.Sprintf("Error: accepts %d arg(s), received %d\n", accepts, received)
}

func addedSuccessMsg(name string) string {
	return fmt.Sprintf("Project '%s' added\n", name)
}

func removedSuccessMsg(name string) string {
	return fmt.Sprintf("Project '%s' removed\n", name)
}

func renamedSuccessMsg(name, newName string) string {
	return fmt.Sprintf("Project '%s' renamed to '%s'\n", name, newName)
}

func maxLengthValidationErrorMsg(field string, length int) string {
	return fmt.Sprintf("Error: \"%s\" exceeds max length of %d characters\n", field, length)
}

func errorMsg(e string) string {
	return fmt.Sprintf("Error: %s\n", e)
}

func TestProjectCmd(t *testing.T) {
	scenarios := map[string]func(t *testing.T, mock sqlmock.Sqlmock, db *sql.DB){
		"test project add command happy path":         testProjectAddCommandHappyPath,
		"test project add command with no args":       testProjectAddCmdWithNoArgs,
		"test project add command with too many args": testProjectAddCmdWithTooManyArgs,
		"test project add command database error":     testProjectAddCmdDatabaseError,
		"test project add command validation error":   testProjectAddCmdValidationError,

		"test project remove command happy path":         testProjectRemoveCmdHappyPath,
		"test project remove command with no args":       testProjectRemoveCmdWithNoArgs,
		"test project remove command with too many args": testProjectRemoveCmdWithTooManyArgs,
		"test project remove command database error":     testProjectRemoveCmdDatabaseError,
		"test project remove command row error":          testProjectRemoveCmdRowError,
		"test project remove command zero affected rows": testProjectRemoveCmdZeroAffectedRows,
		"test project remove command validation error":   testProjectRemoveCmdValidationError,

		"test project rename command happy path":         testProjectRenameCmdHappyPath,
		"test project rename command with no args":       testProjectRenameCmdWithNoArgs,
		"test project rename command with too few args":  testProjectRenameCmdWithTooFewArgs,
		"test project rename command with too many args": testProjectRenameCmdWithTooManyArgs,
		"test project rename command database error":     testProjectRenameCmdDatabaseError,
		"test project rename command validation error":   testProjectRenameCmdValidationError,

		"test project list command happy path":         testProjectListCmdHappyPath,
		"test project list command zero results":       testProjectListCmdZeroResults,
		"test project list command database error":     testProjectListCmdDatabaseError,
		"test project list command with too many args": testProjectListCmdWithTooManyArgs,
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

func testProjectAddCmdWithNoArgs(
	t *testing.T,
	mock sqlmock.Sqlmock,
	db *sql.DB,
) {
	cmdIn := bytes.NewReader([]byte{})
	cmdOut := bytes.NewBufferString("")
	errOut := bytes.NewBufferString("")

	err := cmd.Execute(
		[]*cobra.Command{project.ProjectCommand()},
		[]string{"project", "add"},
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

	require.Equal(t, incorrectNumberOfArgsErrorMsg(1, 0), string(out))
}

func testProjectAddCmdWithTooManyArgs(
	t *testing.T,
	mock sqlmock.Sqlmock,
	db *sql.DB,
) {
	cmdIn := bytes.NewReader([]byte{})
	cmdOut := bytes.NewBufferString("")
	errOut := bytes.NewBufferString("")

	err := cmd.Execute(
		[]*cobra.Command{project.ProjectCommand()},
		[]string{"project", "add", "foo", "bar"},
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

	require.Equal(t, incorrectNumberOfArgsErrorMsg(1, 2), string(out))
}

func testProjectAddCommandHappyPath(
	t *testing.T,
	mock sqlmock.Sqlmock,
	db *sql.DB,
) {
	cmdIn := bytes.NewReader([]byte{})
	cmdOut := bytes.NewBufferString("")

	mock.ExpectExec(regexp.QuoteMeta(`
		insert into projects_ (name_) values ($name)
	`)).WithArgs("my_cool_project").WillReturnResult(sqlmock.NewResult(1, 1))

	err := cmd.Execute(
		[]*cobra.Command{project.ProjectCommand()},
		[]string{"project", "add", "my_cool_project"},
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
		addedSuccessMsg("my_cool_project"),
		string(out),
	)

	require.NoError(t, mock.ExpectationsWereMet())
}

func testProjectAddCmdDatabaseError(
	t *testing.T,
	mock sqlmock.Sqlmock,
	db *sql.DB,
) {
	cmdIn := bytes.NewReader([]byte{})
	cmdOut := bytes.NewBufferString("")
	errOut := bytes.NewBufferString("")

	mock.ExpectExec(regexp.QuoteMeta(`
		insert into projects_ (name_) values ($name)
	`)).WithArgs("my_cool_project").
		WillReturnError(fmt.Errorf("database_error"))

	err := cmd.Execute(
		[]*cobra.Command{project.ProjectCommand()},
		[]string{"project", "add", "my_cool_project"},
		cmdIn,
		cmdOut,
		errOut,
		db,
	)

	require.Error(t, err)

	out, err := io.ReadAll(errOut)
	if err != nil {
		t.Error("unable to read from err out")
	}

	require.Equal(
		t,
		errorMsg("database_error"),
		string(out),
	)

	require.NoError(t, mock.ExpectationsWereMet())
}

func testProjectRemoveCmdHappyPath(
	t *testing.T,
	mock sqlmock.Sqlmock,
	db *sql.DB,
) {
	cmdIn := bytes.NewReader([]byte{})
	cmdOut := bytes.NewBufferString("")

	mock.ExpectExec(regexp.QuoteMeta(`
		delete from projects_ where name_ = $name
	`)).WillReturnResult(sqlmock.NewResult(1, 1))

	err := cmd.Execute(
		[]*cobra.Command{project.ProjectCommand()},
		[]string{"project", "remove", "my_cool_project"},
		cmdIn,
		cmdOut,
		os.Stderr,
		db,
	)

	require.NoError(t, err)

	out, err := io.ReadAll(cmdOut)
	if err != nil {
		t.Error("failed to read from cmd out")
	}

	require.Equal(t, removedSuccessMsg("my_cool_project"), string(out))

	require.NoError(t, mock.ExpectationsWereMet())
}

func testProjectRemoveCmdWithNoArgs(
	t *testing.T,
	mock sqlmock.Sqlmock,
	db *sql.DB,
) {
	cmdIn := bytes.NewReader([]byte{})
	cmdOut := bytes.NewBufferString("")
	errOut := bytes.NewBufferString("")

	err := cmd.Execute(
		[]*cobra.Command{project.ProjectCommand()},
		[]string{"project", "remove"},
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

	require.Equal(t, incorrectNumberOfArgsErrorMsg(1, 0), string(out))
}

func testProjectRemoveCmdWithTooManyArgs(
	t *testing.T,
	mock sqlmock.Sqlmock,
	db *sql.DB,
) {
	cmdIn := bytes.NewReader([]byte{})
	cmdOut := bytes.NewBufferString("")
	errOut := bytes.NewBufferString("")

	err := cmd.Execute(
		[]*cobra.Command{project.ProjectCommand()},
		[]string{"project", "remove", "foo", "bar"},
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

	require.Equal(t, incorrectNumberOfArgsErrorMsg(1, 2), string(out))
}

func testProjectRemoveCmdDatabaseError(
	t *testing.T,
	mock sqlmock.Sqlmock,
	db *sql.DB,
) {
	cmdIn := bytes.NewReader([]byte{})
	cmdOut := bytes.NewBufferString("")
	errOut := bytes.NewBufferString("")

	mock.ExpectExec(regexp.QuoteMeta(`
		delete from projects_ where name_ = $name
	`)).WillReturnError(fmt.Errorf("database_error"))

	err := cmd.Execute(
		[]*cobra.Command{project.ProjectCommand()},
		[]string{"project", "remove", "my_cool_project"},
		cmdIn,
		cmdOut,
		errOut,
		db,
	)

	require.Error(t, err)

	out, err := io.ReadAll(errOut)
	if err != nil {
		t.Errorf("unable to read from err out")
	}

	require.Equal(t, errorMsg("database_error"), string(out))

	require.NoError(t, mock.ExpectationsWereMet())
}

func testProjectRemoveCmdRowError(
	t *testing.T,
	mock sqlmock.Sqlmock,
	db *sql.DB,
) {
	cmdIn := bytes.NewReader([]byte{})
	cmdOut := bytes.NewBufferString("")

	mock.ExpectExec(regexp.QuoteMeta(`
		delete from projects_ where name_ = $name
	`)).WillReturnResult(sqlmock.NewErrorResult(fmt.Errorf("rows_error")))

	err := cmd.Execute(
		[]*cobra.Command{project.ProjectCommand()},
		[]string{"project", "remove", "my_cool_project"},
		cmdIn,
		cmdOut,
		os.Stderr,
		db,
	)

	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func testProjectRemoveCmdZeroAffectedRows(
	t *testing.T,
	mock sqlmock.Sqlmock,
	db *sql.DB,
) {
	cmdIn := bytes.NewReader([]byte{})
	cmdOut := bytes.NewBufferString("")

	mock.ExpectExec(regexp.QuoteMeta(`
		delete from projects_ where name_ = $name
	`)).WillReturnResult(sqlmock.NewResult(1, 0))

	err := cmd.Execute(
		[]*cobra.Command{project.ProjectCommand()},
		[]string{"project", "remove", "my_cool_project"},
		cmdIn,
		cmdOut,
		os.Stderr,
		db,
	)

	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func testProjectRenameCmdHappyPath(
	t *testing.T,
	mock sqlmock.Sqlmock,
	db *sql.DB,
) {
	cmdIn := bytes.NewReader([]byte{})
	cmdOut := bytes.NewBufferString("")

	mock.
		ExpectExec(regexp.QuoteMeta(`
			update projects_ set name_ = $newName where name_ = $originalName
		`)).
		WithArgs("my_cool_project", "my_awesome_project").
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := cmd.Execute(
		[]*cobra.Command{project.ProjectCommand()},
		[]string{"project", "rename", "my_cool_project", "my_awesome_project"},
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
		renamedSuccessMsg("my_cool_project", "my_awesome_project"),
		string(out),
	)

	require.NoError(t, mock.ExpectationsWereMet())
}

func testProjectRenameCmdWithNoArgs(
	t *testing.T,
	mock sqlmock.Sqlmock,
	db *sql.DB,
) {
	cmdIn := bytes.NewReader([]byte{})
	cmdOut := bytes.NewBufferString("")
	errOut := bytes.NewBufferString("")

	err := cmd.Execute(
		[]*cobra.Command{project.ProjectCommand()},
		[]string{"project", "rename"},
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

	require.Equal(t, incorrectNumberOfArgsErrorMsg(2, 0), string(out))
}

func testProjectRenameCmdWithTooFewArgs(
	t *testing.T,
	mock sqlmock.Sqlmock,
	db *sql.DB,
) {
	cmdIn := bytes.NewReader([]byte{})
	cmdOut := bytes.NewBufferString("")
	errOut := bytes.NewBufferString("")

	err := cmd.Execute(
		[]*cobra.Command{project.ProjectCommand()},
		[]string{"project", "rename", "foo"},
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

	require.Equal(t, incorrectNumberOfArgsErrorMsg(2, 1), string(out))
}

func testProjectRenameCmdWithTooManyArgs(
	t *testing.T,
	mock sqlmock.Sqlmock,
	db *sql.DB,
) {
	cmdIn := bytes.NewReader([]byte{})
	cmdOut := bytes.NewBufferString("")
	errOut := bytes.NewBufferString("")

	err := cmd.Execute(
		[]*cobra.Command{project.ProjectCommand()},
		[]string{"project", "rename", "foo", "bar", "baz"},
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

	require.Equal(t, incorrectNumberOfArgsErrorMsg(2, 3), string(out))
}

func testProjectRenameCmdDatabaseError(
	t *testing.T,
	mock sqlmock.Sqlmock,
	db *sql.DB,
) {
	cmdIn := bytes.NewReader([]byte{})
	cmdOut := bytes.NewBufferString("")
	errOut := bytes.NewBufferString("")

	mock.ExpectExec(regexp.QuoteMeta(`
			update projects_ set name_ = $newName where name_ = $originalName
	`)).WithArgs("my_cool_project", "my_awesome_project").
		WillReturnError(fmt.Errorf("database_error"))

	err := cmd.Execute(
		[]*cobra.Command{project.ProjectCommand()},
		[]string{"project", "rename", "my_cool_project", "my_awesome_project"},
		cmdIn,
		cmdOut,
		errOut,
		db,
	)

	require.Error(t, err)

	require.NoError(t, mock.ExpectationsWereMet())
}

func testProjectListCmdHappyPath(
	t *testing.T,
	mock sqlmock.Sqlmock,
	db *sql.DB,
) {
	cmdIn := bytes.NewReader([]byte{})
	cmdOut := bytes.NewBufferString("")

	mock.
		ExpectQuery(regexp.QuoteMeta(`select name_ from projects_`)).
		WillReturnRows(
			sqlmock.NewRows([]string{"name_"}).
				AddRow("my_cool_project").
				AddRow("my_awesome_project").
				AddRow("my_super_project"),
		)

	err := cmd.Execute(
		[]*cobra.Command{project.ProjectCommand()},
		[]string{"project", "list"},
		cmdIn,
		cmdOut,
		os.Stderr,
		db,
	)
	require.NoError(t, err)

	out, err := io.ReadAll(cmdOut)
	require.NoError(t, err)

	require.Equal(
		t,
		strings.Split(strings.TrimSpace(string(out)), "\n"),
		[]string{"my_cool_project", "my_awesome_project", "my_super_project"},
	)

	require.NoError(t, mock.ExpectationsWereMet())
}

func testProjectListCmdZeroResults(
	t *testing.T,
	mock sqlmock.Sqlmock,
	db *sql.DB,
) {
	cmdIn := bytes.NewReader([]byte{})
	cmdOut := bytes.NewBufferString("")

	mock.
		ExpectQuery(regexp.QuoteMeta(`select name_ from projects_`)).
		WillReturnError(sql.ErrNoRows)

	err := cmd.Execute(
		[]*cobra.Command{project.ProjectCommand()},
		[]string{"project", "list"},
		cmdIn,
		cmdOut,
		os.Stderr,
		db,
	)

	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func testProjectListCmdDatabaseError(
	t *testing.T,
	mock sqlmock.Sqlmock,
	db *sql.DB,
) {
	cmdIn := bytes.NewReader([]byte{})
	cmdOut := bytes.NewBufferString("")

	mock.
		ExpectQuery(regexp.QuoteMeta(`select name_ from projects_`)).
		WillReturnError(fmt.Errorf("database_error"))

	err := cmd.Execute(
		[]*cobra.Command{project.ProjectCommand()},
		[]string{"project", "list"},
		cmdIn,
		cmdOut,
		os.Stderr,
		db,
	)

	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func testProjectListCmdWithTooManyArgs(
	t *testing.T,
	mock sqlmock.Sqlmock,
	db *sql.DB,
) {
	cmdIn := bytes.NewReader([]byte{})
	cmdOut := bytes.NewBufferString("")

	err := cmd.Execute(
		[]*cobra.Command{project.ProjectCommand()},
		[]string{"project", "list", "foo"},
		cmdIn,
		cmdOut,
		os.Stderr,
		db,
	)

	require.Error(t, err)
}

func testProjectAddCmdValidationError(
	t *testing.T,
	mock sqlmock.Sqlmock,
	db *sql.DB,
) {
	cmdIn := bytes.NewReader([]byte{})
	cmdOut := bytes.NewBufferString("")
	errOut := bytes.NewBufferString("")

	var err error

	err = cmd.Execute(
		[]*cobra.Command{project.ProjectCommand()},
		[]string{
			"project",
			"add",
			"mmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projecty_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_project",
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
		maxLengthValidationErrorMsg("project name", 256),
		string(out),
	)
}

func testProjectRemoveCmdValidationError(
	t *testing.T,
	mock sqlmock.Sqlmock,
	db *sql.DB,
) {
	cmdIn := bytes.NewReader([]byte{})
	cmdOut := bytes.NewBufferString("")

	var err error

	err = cmd.Execute(
		[]*cobra.Command{project.ProjectCommand()},
		[]string{
			"project",
			"remove",
			"my_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_project",
		},
		cmdIn,
		cmdOut,
		os.Stderr,
		db,
	)

	require.Error(t, err)

	err = cmd.Execute(
		[]*cobra.Command{project.ProjectCommand()},
		[]string{
			"project",
			"remove",
			"",
		},
		cmdIn,
		cmdOut,
		os.Stderr,
		db,
	)

	require.Error(t, err)
}

func testProjectRenameCmdValidationError(
	t *testing.T,
	mock sqlmock.Sqlmock,
	db *sql.DB,
) {
	cmdIn := bytes.NewReader([]byte{})
	cmdOut := bytes.NewBufferString("")

	var err error

	err = cmd.Execute(
		[]*cobra.Command{project.ProjectCommand()},
		[]string{
			"project",
			"rename",
			"my_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_project",
			"",
		},
		cmdIn,
		cmdOut,
		os.Stderr,
		db,
	)

	require.Error(t, err)

	err = cmd.Execute(
		[]*cobra.Command{project.ProjectCommand()},
		[]string{
			"project",
			"rename",
			"",
			"my_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_project",
		},
		cmdIn,
		cmdOut,
		os.Stderr,
		db,
	)

	require.Error(t, err)
}
