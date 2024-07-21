package environment_test

import (
	"bytes"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/nixpig/syringe.sh/internal/environment"
	"github.com/nixpig/syringe.sh/pkg/validation"
	"github.com/nixpig/syringe.sh/test"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

func TestEnvironmentCmd(t *testing.T) {
	scenarios := map[string]func(
		t *testing.T,
		cmd *cobra.Command,
		service environment.EnvironmentService,
		mock sqlmock.Sqlmock,
	){
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
		"test environment remove command zero affected rows":   testEnvironmentRemoveCmdZeroAffectedRows,
		"test environment remove command validation error":     testEnvironmentRemoveCmdValidationError,

		"test environment rename command happy path":           testEnvironmentRenameCmdHappyPath,
		"test environment rename command database error":       testEnvironmentRenameCmdDatabaseError,
		"test environment rename command validation errors":    testEnvironmentRenameCmdValidationError,
		"test environment rename command missing project flag": testEnvironmentRenameCmdMissingProjectFlag,
		"test environment rename command with no args":         testEnvironmentRenameCmdWithNoArgs,
		"test environment rename command with too many args":   testEnvironmentRenameCmdWithTooManyArgs,

		"test environment list command happy path":           testEnvironmentListCmdHappyPath,
		"test environment list command zero results":         testEnvironmentListCmdZeroResults,
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

			cmd := environment.NewCmdEnvironment()

			service := environment.NewEnvironmentServiceImpl(
				environment.NewSqliteEnvironmentStore(db),
				validation.New(),
			)

			fn(t, cmd, service, mock)
		})
	}
}

func testEnvironmentAddCmdHappyPath(
	t *testing.T,
	cmd *cobra.Command,
	service environment.EnvironmentService,
	mock sqlmock.Sqlmock,
) {
	cmdIn := bytes.NewReader([]byte{})
	cmdOut := bytes.NewBufferString("")
	errOut := bytes.NewBufferString("")

	cmdAdd := environment.NewCmdEnvironmentAdd(
		environment.NewHandlerEnvironmentAdd(service),
	)

	cmd.AddCommand(cmdAdd)
	cmd.SetArgs([]string{
		"add",
		"-p",
		"my_cool_project",
		"staging",
	})
	cmd.SetIn(cmdIn)
	cmd.SetOut(cmdOut)
	cmd.SetErr(errOut)

	mock.ExpectExec(regexp.QuoteMeta(`
		insert into environments_ (name_, project_id_) values (
			$name,
			(select id_ from projects_ where name_ = $projectName)
		)
	`)).
		WithArgs("staging", "my_cool_project").
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := cmd.Execute()

	require.NoError(t, err)
	require.Empty(t, errOut.String())

	require.Equal(
		t,
		test.EnvironmentAddedSuccessMsg("staging", "my_cool_project"),
		cmdOut.String(),
	)

	require.NoError(t, mock.ExpectationsWereMet())
}

func testEnvironmentAddCmdMissingProjectFlag(
	t *testing.T,
	cmd *cobra.Command,
	service environment.EnvironmentService,
	mock sqlmock.Sqlmock,
) {
	cmdIn := bytes.NewReader([]byte{})
	cmdOut := bytes.NewBufferString("")
	errOut := bytes.NewBufferString("")

	cmdAdd := environment.NewCmdEnvironmentAdd(
		environment.NewHandlerEnvironmentAdd(service),
	)

	cmd.AddCommand(cmdAdd)
	cmd.SetArgs([]string{
		"add",
		"staging",
	})
	cmd.SetIn(cmdIn)
	cmd.SetOut(cmdOut)
	cmd.SetErr(errOut)

	err := cmd.Execute()

	require.Error(t, err)
	require.Equal(
		t,
		test.RequiredFlagsErrorMsg("project"),
		errOut.String(),
	)

	require.Equal(
		t,
		fmt.Sprintf("%s\n", cmdAdd.UsageString()),
		cmdOut.String(),
	)
}

func testEnvironmentAddCmdWithNoArgs(
	t *testing.T,
	cmd *cobra.Command,
	service environment.EnvironmentService,
	mock sqlmock.Sqlmock,
) {
	cmdIn := bytes.NewReader([]byte{})
	cmdOut := bytes.NewBufferString("")
	errOut := bytes.NewBufferString("")

	cmdAdd := environment.NewCmdEnvironmentAdd(
		environment.NewHandlerEnvironmentAdd(service),
	)

	cmd.AddCommand(cmdAdd)
	cmd.SetArgs([]string{
		"add",
		"-p",
		"my_cool_project",
	})
	cmd.SetIn(cmdIn)
	cmd.SetOut(cmdOut)
	cmd.SetErr(errOut)

	err := cmd.Execute()

	require.Error(t, err)
	require.Equal(
		t,
		test.IncorrectNumberOfArgsErrorMsg(1, 0),
		errOut.String(),
	)

	require.Equal(
		t,
		fmt.Sprintf("%s\n", cmdAdd.UsageString()),
		cmdOut.String(),
	)
}

func testEnvironmentAddCmdWithTooManyArgs(
	t *testing.T,
	cmd *cobra.Command,
	service environment.EnvironmentService,
	mock sqlmock.Sqlmock,
) {
	cmdIn := bytes.NewReader([]byte{})
	cmdOut := bytes.NewBufferString("")
	errOut := bytes.NewBufferString("")

	cmdAdd := environment.NewCmdEnvironmentAdd(
		environment.NewHandlerEnvironmentAdd(service),
	)

	cmd.AddCommand(cmdAdd)
	cmd.SetArgs([]string{
		"add",
		"-p",
		"my_cool_project",
		"foo",
		"bar",
	})
	cmd.SetIn(cmdIn)
	cmd.SetOut(cmdOut)
	cmd.SetErr(errOut)

	err := cmd.Execute()

	require.Error(t, err)
	require.Equal(
		t,
		test.IncorrectNumberOfArgsErrorMsg(1, 2),
		errOut.String(),
	)

	require.Equal(
		t,
		fmt.Sprintf("%s\n", cmdAdd.UsageString()),
		cmdOut.String(),
	)
}

func testEnvironmentAddCmdDatabaseError(
	t *testing.T,
	cmd *cobra.Command,
	service environment.EnvironmentService,
	mock sqlmock.Sqlmock,
) {
	cmdIn := bytes.NewReader([]byte{})
	cmdOut := bytes.NewBufferString("")
	errOut := bytes.NewBufferString("")

	cmdAdd := environment.NewCmdEnvironmentAdd(
		environment.NewHandlerEnvironmentAdd(service),
	)

	cmd.AddCommand(cmdAdd)
	cmd.SetArgs([]string{
		"add",
		"-p",
		"my_cool_project",
		"staging",
	})
	cmd.SetIn(cmdIn)
	cmd.SetOut(cmdOut)
	cmd.SetErr(errOut)

	mock.ExpectExec(regexp.QuoteMeta(`
		insert into environments_ (name_, project_id_) values (
			$name,
			(select id_ from projects_ where name_ = $projectName)
		)
	`)).
		WithArgs("staging", "my_cool_project").
		WillReturnError(fmt.Errorf("database_error"))

	err := cmd.Execute()

	require.Error(t, err)
	require.Equal(
		t,
		test.ErrorMsg("environment add database error: database_error\n"),
		errOut.String(),
	)

	require.Equal(
		t,
		fmt.Sprintf("%s\n", cmdAdd.UsageString()),
		cmdOut.String(),
	)

	require.NoError(t, mock.ExpectationsWereMet())
}

func testEnvironmentAddCmdValidationError(
	t *testing.T,
	cmd *cobra.Command,
	service environment.EnvironmentService,
	mock sqlmock.Sqlmock,
) {
	var err error
	var cmdAdd *cobra.Command
	var cmdIn *bytes.Reader
	var cmdOut *bytes.Buffer
	var errOut *bytes.Buffer

	cmdIn = bytes.NewReader([]byte{})
	cmdOut = bytes.NewBufferString("")
	errOut = bytes.NewBufferString("")

	cmdAdd = environment.NewCmdEnvironmentAdd(
		environment.NewHandlerEnvironmentAdd(service),
	)

	cmd.AddCommand(cmdAdd)
	cmd.SetArgs([]string{
		"add",
		"-p",
		"my_cool_project",
		"stagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstaging",
	})
	cmd.SetIn(cmdIn)
	cmd.SetOut(cmdOut)
	cmd.SetErr(errOut)

	err = cmd.Execute()

	require.Error(t, err)
	require.Equal(
		t,
		test.MaxLengthValidationErrorMsg("environment name", 256),
		errOut.String(),
	)
	require.Equal(
		t,
		fmt.Sprintf("%s\n", cmdAdd.UsageString()),
		cmdOut.String(),
	)

	cmd.RemoveCommand(cmdAdd)

	cmdIn = bytes.NewReader([]byte{})
	cmdOut = bytes.NewBufferString("")
	errOut = bytes.NewBufferString("")

	cmdAdd = environment.NewCmdEnvironmentAdd(
		environment.NewHandlerEnvironmentAdd(service),
	)
	cmd.AddCommand(cmdAdd)
	cmd.SetIn(cmdIn)
	cmd.SetOut(cmdOut)
	cmd.SetErr(errOut)
	cmd.SetArgs([]string{
		"add",
		"-p",
		"mmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projecty_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_project",
		"staging",
	})

	err = cmd.Execute()

	require.Error(t, err)
	require.Equal(
		t,
		test.MaxLengthValidationErrorMsg("project name", 256),
		errOut.String(),
	)
	require.Equal(
		t,
		fmt.Sprintf("%s\n", cmdAdd.UsageString()),
		cmdOut.String(),
	)
}

func testEnvironmentRemoveCmdHappyPath(
	t *testing.T,
	cmd *cobra.Command,
	service environment.EnvironmentService,
	mock sqlmock.Sqlmock,
) {
	cmdIn := bytes.NewReader([]byte{})
	cmdOut := bytes.NewBufferString("")
	errOut := bytes.NewBufferString("")

	cmdRemove := environment.NewCmdEnvironmentRemove(
		environment.NewHandlerEnvironmentRemove(service),
	)

	cmd.AddCommand(cmdRemove)
	cmd.SetArgs([]string{
		"remove",
		"-p",
		"my_cool_project",
		"staging",
	})
	cmd.SetIn(cmdIn)
	cmd.SetOut(cmdOut)
	cmd.SetErr(errOut)

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

	err := cmd.Execute()

	require.NoError(t, err)
	require.Empty(t, errOut.String())

	require.Equal(
		t,
		test.EnvironmentRemovedSuccessMsg("staging", "my_cool_project"),
		cmdOut.String(),
	)

	require.NoError(t, mock.ExpectationsWereMet())
}

func testEnvironmentRemoveCmdDatabaseError(
	t *testing.T,
	cmd *cobra.Command,
	service environment.EnvironmentService,
	mock sqlmock.Sqlmock,
) {
	cmdIn := bytes.NewReader([]byte{})
	cmdOut := bytes.NewBufferString("")
	errOut := bytes.NewBufferString("")

	cmdRemove := environment.NewCmdEnvironmentRemove(
		environment.NewHandlerEnvironmentRemove(service),
	)

	cmd.AddCommand(cmdRemove)

	cmd.SetArgs([]string{
		"remove",
		"-p",
		"my_cool_project",
		"staging",
	})
	cmd.SetIn(cmdIn)
	cmd.SetOut(cmdOut)
	cmd.SetErr(errOut)

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

	err := cmd.Execute()

	require.Error(t, err)

	require.Equal(
		t,
		test.ErrorMsg("environment remove database error: database_error\n"),
		errOut.String(),
	)

	require.Equal(
		t,
		fmt.Sprintf("%s\n", cmdRemove.UsageString()),
		cmdOut.String(),
	)

	require.NoError(t, mock.ExpectationsWereMet())
}

func testEnvironmentRemoveCmdMissingProjectFlag(
	t *testing.T,
	cmd *cobra.Command,
	service environment.EnvironmentService,
	mock sqlmock.Sqlmock,
) {
	cmdIn := bytes.NewReader([]byte{})
	cmdOut := bytes.NewBufferString("")
	errOut := bytes.NewBufferString("")

	cmdRemove := environment.NewCmdEnvironmentRemove(
		environment.NewHandlerEnvironmentRemove(service),
	)

	cmd.AddCommand(cmdRemove)

	cmd.SetArgs([]string{
		"remove",
		"staging",
	})
	cmd.SetIn(cmdIn)
	cmd.SetOut(cmdOut)
	cmd.SetErr(errOut)

	err := cmd.Execute()

	require.Error(t, err)

	require.Equal(
		t,
		test.RequiredFlagsErrorMsg("project"),
		errOut.String(),
	)

	require.Equal(
		t,
		fmt.Sprintf("%s\n", cmdRemove.UsageString()),
		cmdOut.String(),
	)

	require.NoError(t, mock.ExpectationsWereMet())
}

func testEnvironmentRemoveCmdWithNoArgs(
	t *testing.T,
	cmd *cobra.Command,
	service environment.EnvironmentService,
	mock sqlmock.Sqlmock,
) {
	cmdIn := bytes.NewReader([]byte{})
	cmdOut := bytes.NewBufferString("")
	errOut := bytes.NewBufferString("")

	cmdRemove := environment.NewCmdEnvironmentRemove(
		environment.NewHandlerEnvironmentRemove(service),
	)

	cmd.AddCommand(cmdRemove)

	cmd.SetArgs([]string{
		"remove",
	})
	cmd.SetIn(cmdIn)
	cmd.SetOut(cmdOut)
	cmd.SetErr(errOut)

	err := cmd.Execute()

	require.Error(t, err)

	require.Equal(
		t,
		test.IncorrectNumberOfArgsErrorMsg(1, 0),
		errOut.String(),
	)

	require.Equal(
		t,
		fmt.Sprintf("%s\n", cmdRemove.UsageString()),
		cmdOut.String(),
	)

	require.NoError(t, mock.ExpectationsWereMet())
}

func testEnvironmentRemoveCmdWithTooManyArgs(
	t *testing.T,
	cmd *cobra.Command,
	service environment.EnvironmentService,
	mock sqlmock.Sqlmock,
) {
	cmdIn := bytes.NewReader([]byte{})
	cmdOut := bytes.NewBufferString("")
	errOut := bytes.NewBufferString("")

	cmdRemove := environment.NewCmdEnvironmentRemove(
		environment.NewHandlerEnvironmentRemove(service),
	)

	cmd.AddCommand(cmdRemove)

	cmd.SetArgs([]string{
		"remove",
		"foo",
		"bar",
	})
	cmd.SetIn(cmdIn)
	cmd.SetOut(cmdOut)
	cmd.SetErr(errOut)

	err := cmd.Execute()

	require.Error(t, err)

	require.Equal(
		t,
		test.IncorrectNumberOfArgsErrorMsg(1, 2),
		errOut.String(),
	)

	require.Equal(
		t,
		fmt.Sprintf("%s\n", cmdRemove.UsageString()),
		cmdOut.String(),
	)

	require.NoError(t, mock.ExpectationsWereMet())
}

func testEnvironmentRemoveCmdValidationError(
	t *testing.T,
	cmd *cobra.Command,
	service environment.EnvironmentService,
	mock sqlmock.Sqlmock,
) {

	cmdIn := bytes.NewReader([]byte{})
	cmdOut := bytes.NewBufferString("")
	errOut := bytes.NewBufferString("")

	cmdRemove := environment.NewCmdEnvironmentRemove(
		environment.NewHandlerEnvironmentRemove(service),
	)

	cmd.AddCommand(cmdRemove)

	cmd.SetArgs([]string{
		"remove",
		"-p",
		"my_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_project",
		"stagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstaging",
	})
	cmd.SetIn(cmdIn)
	cmd.SetOut(cmdOut)
	cmd.SetErr(errOut)

	err := cmd.Execute()

	require.Error(t, err)
	require.Equal(
		t,
		test.MaxLengthValidationErrorMsg("environment name", 256),
		errOut.String(),
	)

	require.Equal(
		t,
		fmt.Sprintf("%s\n", cmdRemove.UsageString()),
		cmdOut.String(),
	)

	require.NoError(t, mock.ExpectationsWereMet())
}

func testEnvironmentRenameCmdHappyPath(
	t *testing.T,
	cmd *cobra.Command,
	service environment.EnvironmentService,
	mock sqlmock.Sqlmock,
) {
	cmdIn := bytes.NewReader([]byte{})
	cmdOut := bytes.NewBufferString("")
	errOut := bytes.NewBufferString("")

	cmdAdd := environment.NewCmdEnvironmentRename(
		environment.NewHandlerEnvironmentRename(service),
	)

	cmd.AddCommand(cmdAdd)
	cmd.SetArgs([]string{
		"rename",
		"-p",
		"my_cool_project",
		"staging",
		"prod",
	})
	cmd.SetIn(cmdIn)
	cmd.SetOut(cmdOut)
	cmd.SetErr(errOut)

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

	err := cmd.Execute()

	require.NoError(t, err)
	require.Empty(t, errOut.String())

	require.Equal(
		t,
		test.EnvironmentRenamedSuccessMsg("staging", "prod", "my_cool_project"),
		cmdOut.String(),
	)

	require.NoError(t, mock.ExpectationsWereMet())
}

func testEnvironmentRenameCmdDatabaseError(
	t *testing.T,
	cmd *cobra.Command,
	service environment.EnvironmentService,
	mock sqlmock.Sqlmock,
) {
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

	cmdRename := environment.NewCmdEnvironmentRename(
		environment.NewHandlerEnvironmentRename(service),
	)

	cmd.AddCommand(cmdRename)

	cmd.SetArgs([]string{
		"rename",
		"-p",
		"my_cool_project",
		"staging",
		"prod",
	})
	cmd.SetIn(cmdIn)
	cmd.SetOut(cmdOut)
	cmd.SetErr(errOut)

	err := cmd.Execute()

	require.Error(t, err)

	out, err := io.ReadAll(errOut)
	if err != nil {
		t.Errorf("failed to read from out")
	}

	require.Equal(
		t,
		test.ErrorMsg("environment rename database error: database_error\n"),
		string(out),
	)

	require.Equal(
		t,
		fmt.Sprintf("%s\n", cmdRename.UsageString()),
		cmdOut.String(),
	)

	require.NoError(t, mock.ExpectationsWereMet())
}

func testEnvironmentRenameCmdValidationError(
	t *testing.T,
	cmd *cobra.Command,
	service environment.EnvironmentService,
	mock sqlmock.Sqlmock,
) {
	cmdIn := bytes.NewReader([]byte{})
	cmdOut := bytes.NewBufferString("")
	errOut := bytes.NewBufferString("")

	cmdRename := environment.NewCmdEnvironmentRename(
		environment.NewHandlerEnvironmentRename(service),
	)

	cmd.AddCommand(cmdRename)

	cmd.SetArgs([]string{
		"rename",
		"-p",
		"my_cool_project",
		"staging",
		"new_name_new_name_new_name_new_name_new_name_new_name_new_name_new_name_new_name_new_name_new_name_new_name_new_name_new_name_new_name_new_name_new_name_new_name_new_name_new_name_new_name_new_name_new_name_new_name_new_name_new_name_new_name_new_name_new_name_new_name_new_name_new_name_new_name_new_name_new_name_new_name_",
	})
	cmd.SetIn(cmdIn)
	cmd.SetOut(cmdOut)
	cmd.SetErr(errOut)

	err := cmd.Execute()

	require.Error(t, err)

	require.Equal(
		t,
		test.MaxLengthValidationErrorMsg("new environment name", 256),
		errOut.String(),
	)

	require.Equal(
		t,
		fmt.Sprintf("%s\n", cmdRename.UsageString()),
		cmdOut.String(),
	)
}

func testEnvironmentRenameCmdMissingProjectFlag(
	t *testing.T,
	cmd *cobra.Command,
	service environment.EnvironmentService,
	mock sqlmock.Sqlmock,
) {
	cmdIn := bytes.NewReader([]byte{})
	cmdOut := bytes.NewBufferString("")
	errOut := bytes.NewBufferString("")

	cmdRename := environment.NewCmdEnvironmentRename(
		environment.NewHandlerEnvironmentRename(service),
	)

	cmd.AddCommand(cmdRename)

	cmd.SetArgs([]string{
		"rename",
		"current_name",
		"new_name",
	})
	cmd.SetIn(cmdIn)
	cmd.SetOut(cmdOut)
	cmd.SetErr(errOut)

	err := cmd.Execute()

	require.Error(t, err)

	require.Equal(
		t,
		test.RequiredFlagsErrorMsg("project"),
		errOut.String(),
	)

	require.Equal(
		t,
		fmt.Sprintf("%s\n", cmdRename.UsageString()),
		cmdOut.String(),
	)
}

func testEnvironmentRenameCmdWithNoArgs(
	t *testing.T,
	cmd *cobra.Command,
	service environment.EnvironmentService,
	mock sqlmock.Sqlmock,
) {
	cmdIn := bytes.NewReader([]byte{})
	cmdOut := bytes.NewBufferString("")
	errOut := bytes.NewBufferString("")

	cmdRename := environment.NewCmdEnvironmentRename(
		environment.NewHandlerEnvironmentRename(service),
	)

	cmd.AddCommand(cmdRename)

	cmd.SetArgs([]string{
		"rename",
		"-p",
		"my_cool_project",
	})
	cmd.SetIn(cmdIn)
	cmd.SetOut(cmdOut)
	cmd.SetErr(errOut)

	err := cmd.Execute()

	require.Error(t, err)

	require.Equal(
		t,
		test.IncorrectNumberOfArgsErrorMsg(2, 0),
		errOut.String(),
	)

	require.Equal(
		t,
		fmt.Sprintf("%s\n", cmdRename.UsageString()),
		cmdOut.String(),
	)
}

func testEnvironmentRenameCmdWithTooManyArgs(
	t *testing.T,
	cmd *cobra.Command,
	service environment.EnvironmentService,
	mock sqlmock.Sqlmock,
) {
	cmdIn := bytes.NewReader([]byte{})
	cmdOut := bytes.NewBufferString("")
	errOut := bytes.NewBufferString("")

	cmdRename := environment.NewCmdEnvironmentRename(
		environment.NewHandlerEnvironmentRename(service),
	)

	cmd.AddCommand(cmdRename)

	cmd.SetArgs([]string{
		"rename",
		"-p",
		"my_cool_project",
		"foo",
		"bar",
		"baz",
	})
	cmd.SetIn(cmdIn)
	cmd.SetOut(cmdOut)
	cmd.SetErr(errOut)

	err := cmd.Execute()

	require.Error(t, err)

	require.Equal(
		t,
		test.IncorrectNumberOfArgsErrorMsg(2, 3),
		errOut.String(),
	)

	require.Equal(
		t,
		fmt.Sprintf("%s\n", cmdRename.UsageString()),
		cmdOut.String(),
	)
}

func testEnvironmentListCmdHappyPath(
	t *testing.T,
	cmd *cobra.Command,
	service environment.EnvironmentService,
	mock sqlmock.Sqlmock,
) {
	cmdIn := bytes.NewReader([]byte{})
	cmdOut := bytes.NewBufferString("")
	errOut := bytes.NewBufferString("")

	cmdList := environment.NewCmdEnvironmentList(
		environment.NewHandlerEnvironmentList(service),
	)

	cmd.AddCommand(cmdList)
	cmd.SetArgs([]string{
		"list",
		"-p",
		"my_cool_project",
	})
	cmd.SetIn(cmdIn)
	cmd.SetOut(cmdOut)
	cmd.SetErr(errOut)

	query := `
		select e.id_, e.name_, p.name_ from environments_ e
		inner join projects_ p
		on e.project_id_ = p.id_
		where p.name_ = $projectName
	`

	mock.
		ExpectQuery(regexp.QuoteMeta(query)).
		WithArgs("my_cool_project").
		WillReturnRows(
			sqlmock.
				NewRows([]string{"id_", "name_", "project_name_"}).
				AddRow(1, "dev", "my_cool_project").
				AddRow(2, "staging", "my_cool_project").
				AddRow(3, "prod", "my_cool_project"),
		)

	err := cmd.Execute()

	require.NoError(t, err)
	require.Empty(t, errOut.String())

	require.Equal(
		t,
		"dev\nstaging\nprod\n",
		cmdOut.String(),
	)

	require.NoError(t, mock.ExpectationsWereMet())
}

func testEnvironmentListCmdDatabaseError(
	t *testing.T,
	cmd *cobra.Command,
	service environment.EnvironmentService,
	mock sqlmock.Sqlmock,
) {
	cmdIn := bytes.NewReader([]byte{})
	cmdOut := bytes.NewBufferString("")
	errOut := bytes.NewBufferString("")

	query := `
		select e.id_, e.name_, p.name_ from environments_ e
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

	cmdList := environment.NewCmdEnvironmentList(
		environment.NewHandlerEnvironmentList(service),
	)

	cmd.AddCommand(cmdList)
	cmd.SetArgs([]string{
		"list",
		"-p",
		"my_cool_project",
	})
	cmd.SetIn(cmdIn)
	cmd.SetOut(cmdOut)
	cmd.SetErr(errOut)

	err := cmd.Execute()

	require.Error(t, err)

	require.Equal(
		t,
		test.ErrorMsg("environment list database error: database_error\n"),
		errOut.String(),
	)

	require.Equal(
		t,
		fmt.Sprintf("%s\n", cmdList.UsageString()),
		cmdOut.String(),
	)

	require.NoError(t, mock.ExpectationsWereMet())
}

func testEnvironmentListCmdValidationError(
	t *testing.T,
	cmd *cobra.Command,
	service environment.EnvironmentService,
	mock sqlmock.Sqlmock,
) {
	cmdIn := bytes.NewReader([]byte{})
	cmdOut := bytes.NewBufferString("")
	errOut := bytes.NewBufferString("")

	cmdList := environment.NewCmdEnvironmentList(
		environment.NewHandlerEnvironmentList(service),
	)

	cmd.AddCommand(cmdList)
	cmd.SetArgs([]string{
		"list",
		"-p",
		"my_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_project",
	})
	cmd.SetIn(cmdIn)
	cmd.SetOut(cmdOut)
	cmd.SetErr(errOut)

	err := cmd.Execute()

	require.Error(t, err)

	require.Equal(
		t,
		test.MaxLengthValidationErrorMsg("project name", 256),
		errOut.String(),
	)

	require.Equal(
		t,
		fmt.Sprintf("%s\n", cmdList.UsageString()),
		cmdOut.String(),
	)
}

func testEnvironmentListCmdMissingProjectFlag(
	t *testing.T,
	cmd *cobra.Command,
	service environment.EnvironmentService,
	mock sqlmock.Sqlmock,
) {
	cmdIn := bytes.NewReader([]byte{})
	cmdOut := bytes.NewBufferString("")
	errOut := bytes.NewBufferString("")

	cmdList := environment.NewCmdEnvironmentList(
		environment.NewHandlerEnvironmentList(service),
	)

	cmd.AddCommand(cmdList)
	cmd.SetArgs([]string{
		"list",
	})
	cmd.SetIn(cmdIn)
	cmd.SetOut(cmdOut)
	cmd.SetErr(errOut)

	err := cmd.Execute()

	require.Error(t, err)

	require.Equal(
		t,
		test.RequiredFlagsErrorMsg("project"),
		errOut.String(),
	)

	require.Equal(
		t,
		fmt.Sprintf("%s\n", cmdList.UsageString()),
		cmdOut.String(),
	)
}

func testEnvironmentListCmdWithTooManyArgs(
	t *testing.T,
	cmd *cobra.Command,
	service environment.EnvironmentService,
	mock sqlmock.Sqlmock,
) {
	cmdIn := bytes.NewReader([]byte{})
	cmdOut := bytes.NewBufferString("")
	errOut := bytes.NewBufferString("")

	cmdList := environment.NewCmdEnvironmentList(
		environment.NewHandlerEnvironmentList(service),
	)

	cmd.AddCommand(cmdList)
	cmd.SetArgs([]string{
		"list",
		"-p",
		"my_cool_project",
		"foobarbaz",
	})
	cmd.SetIn(cmdIn)
	cmd.SetOut(cmdOut)
	cmd.SetErr(errOut)

	err := cmd.Execute()

	require.Error(t, err)

	require.Equal(
		t,
		test.IncorrectNumberOfArgsErrorMsg(0, 1),
		errOut.String(),
	)

	require.Equal(
		t,
		fmt.Sprintf("%s\n", cmdList.UsageString()),
		cmdOut.String(),
	)
}

func testEnvironmentListCmdZeroResults(
	t *testing.T,
	cmd *cobra.Command,
	service environment.EnvironmentService,
	mock sqlmock.Sqlmock,
) {
	cmdIn := bytes.NewReader([]byte{})
	cmdOut := bytes.NewBufferString("")
	errOut := bytes.NewBufferString("")

	query := `
		select e.id_, e.name_, p.name_ from environments_ e
		inner join projects_ p
		on e.project_id_ = p.id_
		where p.name_ = $projectName
	`

	mock.
		ExpectQuery(regexp.QuoteMeta(query)).
		WithArgs("my_cool_project").
		WillReturnError(sql.ErrNoRows)

	cmdList := environment.NewCmdEnvironmentList(
		environment.NewHandlerEnvironmentList(service),
	)

	cmd.AddCommand(cmdList)

	cmd.SetArgs([]string{
		"list",
		"-p",
		"my_cool_project",
	})
	cmd.SetIn(cmdIn)
	cmd.SetOut(cmdOut)
	cmd.SetErr(errOut)

	err := cmd.Execute()

	require.Error(t, err)

	require.Equal(
		t,
		test.ErrorMsg("no environments found\n"),
		errOut.String(),
	)

	require.Equal(
		t,
		fmt.Sprintf("%s\n", cmdList.UsageString()),
		cmdOut.String(),
	)

	require.NoError(t, mock.ExpectationsWereMet())
}

func testEnvironmentRemoveCmdZeroAffectedRows(
	t *testing.T,
	cmd *cobra.Command,
	service environment.EnvironmentService,
	mock sqlmock.Sqlmock,
) {
	cmdIn := bytes.NewReader([]byte{})
	cmdOut := bytes.NewBufferString("")
	errOut := bytes.NewBufferString("")

	cmdRemove := environment.NewCmdEnvironmentRemove(
		environment.NewHandlerEnvironmentRemove(service),
	)

	cmd.AddCommand(cmdRemove)

	cmd.SetArgs([]string{
		"remove",
		"-p",
		"my_cool_project",
		"staging",
	})
	cmd.SetIn(cmdIn)
	cmd.SetOut(cmdOut)
	cmd.SetErr(errOut)

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
		WillReturnResult(sqlmock.NewResult(0, 0))

	err := cmd.Execute()

	require.Error(t, err)

	require.Equal(
		t,
		test.ErrorMsg("environment 'staging' not found\n"),
		errOut.String(),
	)

	require.Equal(
		t,
		fmt.Sprintf("%s\n", cmdRemove.UsageString()),
		cmdOut.String(),
	)

	require.NoError(t, mock.ExpectationsWereMet())
}
