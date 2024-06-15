package project_test

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
	"github.com/nixpig/syringe.sh/server/cmd/project"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

func TestProjectCmd(t *testing.T) {
	scenarios := map[string]func(t *testing.T, mock sqlmock.Sqlmock, db *sql.DB){
		"test project add command happy path":         testProjectAddCommandHappyPath,
		"test project add command with no args":       testProjectAddCmdWithNoArgs,
		"test project add command with too many args": testProjectAddCmdWithTooManyArgs,
		"test project add command database error":     testProjectAddDatabaseError,

		"test project remove command happy path":   testProjectRemoveCmdHappyPath,
		"test project remove command with no args": testProjectRemoveCmdWithNoArgs,
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

	err := cmd.Execute(
		[]*cobra.Command{project.ProjectCommand()},
		[]string{"project", "add"},
		cmdIn,
		cmdOut,
		os.Stderr,
		db,
	)

	require.Error(t, err)
}

func testProjectAddCmdWithTooManyArgs(
	t *testing.T,
	mock sqlmock.Sqlmock,
	db *sql.DB,
) {
	cmdIn := bytes.NewReader([]byte{})
	cmdOut := bytes.NewBufferString("")

	err := cmd.Execute(
		[]*cobra.Command{project.ProjectCommand()},
		[]string{"project", "add", "foo", "bar"},
		cmdIn,
		cmdOut,
		os.Stderr,
		db,
	)

	require.Error(t, err)
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
		"",
		string(out),
		"should not output anything",
	)

	require.NoError(t, mock.ExpectationsWereMet())
}

func testProjectAddDatabaseError(t *testing.T, mock sqlmock.Sqlmock, db *sql.DB) {
	cmdIn := bytes.NewReader([]byte{})
	cmdOut := bytes.NewBufferString("")

	mock.ExpectExec(regexp.QuoteMeta(`
		insert into projects_ (name_) values ($name)
	`)).WithArgs("my_cool_project").
		WillReturnError(errors.New("database_error"))

	err := cmd.Execute(
		[]*cobra.Command{project.ProjectCommand()},
		[]string{"project", "add", "my_cool_project"},
		cmdIn,
		cmdOut,
		os.Stderr,
		db,
	)

	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func testProjectRemoveCmdHappyPath(t *testing.T, mock sqlmock.Sqlmock, db *sql.DB) {
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
	require.NoError(t, mock.ExpectationsWereMet())
}

func testProjectRemoveCmdWithNoArgs(t *testing.T, mock sqlmock.Sqlmock, db *sql.DB) {

}
