package project_test

import (
	"bytes"
	"database/sql"
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/nixpig/syringe.sh/internal/project"
	"github.com/nixpig/syringe.sh/pkg/validation"
	"github.com/nixpig/syringe.sh/test"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

func TestProjectCmd(t *testing.T) {
	scenarios := map[string]func(
		t *testing.T,
		cmd *cobra.Command,
		service project.ProjectService,
		mock sqlmock.Sqlmock,
	){
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
		"test project list command scan error":         testProjectListCmdScanError,
		"test project list command with too many args": testProjectListCmdWithTooManyArgs,
	}

	for scenario, fn := range scenarios {
		t.Run(scenario, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("unable to create mock database:\n%s", err)
			}

			cmd := project.NewCmdProject()

			service := project.NewProjectServiceImpl(
				project.NewSqliteProjectStore(db),
				validation.New(),
			)

			fn(t, cmd, service, mock)
		})
	}
}

func testProjectAddCmdWithNoArgs(
	t *testing.T,
	cmd *cobra.Command,
	service project.ProjectService,
	mock sqlmock.Sqlmock,
) {
	cmdIn := bytes.NewReader([]byte{})
	cmdOut := bytes.NewBufferString("")
	errOut := bytes.NewBufferString("")

	cmdAdd := project.NewCmdProjectAdd(project.NewHandlerProjectAdd(service))

	cmd.AddCommand(cmdAdd)
	cmd.SetArgs([]string{"add"})
	cmd.SetIn(cmdIn)
	cmd.SetOut(cmdOut)
	cmd.SetErr(errOut)

	err := cmd.Execute()

	require.Error(t, err)

	require.Equal(
		t,
		fmt.Sprintf("%s\n", cmdAdd.UsageString()),
		cmdOut.String(),
	)

	require.Equal(
		t,
		test.IncorrectNumberOfArgsErrorMsg(1, 0),
		string(errOut.String()),
	)
}

func testProjectAddCmdWithTooManyArgs(
	t *testing.T,
	cmd *cobra.Command,
	service project.ProjectService,
	mock sqlmock.Sqlmock,
) {
	cmdIn := bytes.NewReader([]byte{})
	cmdOut := bytes.NewBufferString("")
	errOut := bytes.NewBufferString("")

	cmdAdd := project.NewCmdProjectAdd(project.NewHandlerProjectAdd(service))

	cmd.AddCommand(cmdAdd)

	cmd.SetArgs([]string{"add", "foo", "bar"})
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

func testProjectAddCommandHappyPath(
	t *testing.T,
	cmd *cobra.Command,
	service project.ProjectService,
	mock sqlmock.Sqlmock,
) {
	cmdIn := bytes.NewReader([]byte{})
	cmdOut := bytes.NewBufferString("")
	errOut := bytes.NewBufferString("")

	cmd.AddCommand(
		project.NewCmdProjectAdd(project.NewHandlerProjectAdd(service)),
	)

	cmd.SetArgs([]string{"add", "my_cool_project"})
	cmd.SetIn(cmdIn)
	cmd.SetOut(cmdOut)
	cmd.SetErr(errOut)

	mock.ExpectExec(regexp.QuoteMeta(`
		insert into projects_ (name_) values ($name)
	`)).WithArgs("my_cool_project").WillReturnResult(sqlmock.NewResult(1, 1))

	err := cmd.Execute()

	require.NoError(t, err)
	require.Empty(t, errOut.String())

	require.Equal(
		t,
		test.ProjectAddedSuccessMsg("my_cool_project"),
		cmdOut.String(),
	)

	require.NoError(t, mock.ExpectationsWereMet())
}

func testProjectAddCmdDatabaseError(
	t *testing.T,
	cmd *cobra.Command,
	service project.ProjectService,
	mock sqlmock.Sqlmock,
) {
	cmdIn := bytes.NewReader([]byte{})
	cmdOut := bytes.NewBufferString("")
	errOut := bytes.NewBufferString("")

	cmdAdd := project.NewCmdProjectAdd(project.NewHandlerProjectAdd(service))

	cmd.AddCommand(cmdAdd)

	cmd.SetArgs([]string{"add", "my_cool_project"})
	cmd.SetIn(cmdIn)
	cmd.SetOut(cmdOut)
	cmd.SetErr(errOut)

	mock.ExpectExec(regexp.QuoteMeta(`
		insert into projects_ (name_) values ($name)
	`)).WithArgs("my_cool_project").
		WillReturnError(fmt.Errorf("database_error"))

	err := cmd.Execute()

	require.Error(t, err)

	require.Equal(
		t,
		test.ErrorMsg("project add database error: database_error\n"),
		errOut.String(),
	)

	require.Equal(
		t,
		fmt.Sprintf("%s\n", cmdAdd.UsageString()),
		cmdOut.String(),
	)

	require.NoError(t, mock.ExpectationsWereMet())
}

func testProjectRemoveCmdHappyPath(
	t *testing.T,
	cmd *cobra.Command,
	service project.ProjectService,
	mock sqlmock.Sqlmock,
) {
	cmdIn := bytes.NewReader([]byte{})
	cmdOut := bytes.NewBufferString("")
	errOut := bytes.NewBufferString("")

	cmd.AddCommand(
		project.NewCmdProjectRemove(project.NewHandlerProjectRemove(service)),
	)

	cmd.SetArgs([]string{"remove", "my_cool_project"})
	cmd.SetIn(cmdIn)
	cmd.SetOut(cmdOut)
	cmd.SetErr(errOut)

	mock.ExpectExec(regexp.QuoteMeta(`
		delete from projects_ where name_ = $name
	`)).
		WithArgs("my_cool_project").
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := cmd.Execute()

	require.NoError(t, err)
	require.Empty(t, errOut.String())

	require.Equal(
		t,
		test.ProjectRemovedSuccessMsg("my_cool_project"),
		cmdOut.String(),
	)

	require.NoError(t, mock.ExpectationsWereMet())
}

func testProjectRemoveCmdWithNoArgs(
	t *testing.T,
	cmd *cobra.Command,
	service project.ProjectService,
	mock sqlmock.Sqlmock,
) {
	cmdIn := bytes.NewReader([]byte{})
	cmdOut := bytes.NewBufferString("")
	errOut := bytes.NewBufferString("")

	cmdRemove := project.NewCmdProjectRemove(
		project.NewHandlerProjectRemove(service),
	)

	cmd.AddCommand(cmdRemove)
	cmd.SetArgs([]string{"remove"})
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
}

func testProjectRemoveCmdWithTooManyArgs(
	t *testing.T,
	cmd *cobra.Command,
	service project.ProjectService,
	mock sqlmock.Sqlmock,
) {
	cmdIn := bytes.NewReader([]byte{})
	cmdOut := bytes.NewBufferString("")
	errOut := bytes.NewBufferString("")

	cmdRemove := project.NewCmdProjectRemove(
		project.NewHandlerProjectRemove(service),
	)

	cmd.AddCommand(cmdRemove)
	cmd.SetArgs([]string{"remove", "foo", "bar"})
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
}

func testProjectRemoveCmdDatabaseError(
	t *testing.T,
	cmd *cobra.Command,
	service project.ProjectService,
	mock sqlmock.Sqlmock,
) {
	cmdIn := bytes.NewReader([]byte{})
	cmdOut := bytes.NewBufferString("")
	errOut := bytes.NewBufferString("")

	cmdRemove := project.NewCmdProjectRemove(
		project.NewHandlerProjectRemove(service),
	)

	cmd.AddCommand(cmdRemove)
	cmd.SetArgs([]string{"remove", "my_cool_project"})
	cmd.SetIn(cmdIn)
	cmd.SetOut(cmdOut)
	cmd.SetErr(errOut)

	mock.ExpectExec(regexp.QuoteMeta(`
		delete from projects_ where name_ = $name
	`)).WillReturnError(fmt.Errorf("database_error"))

	err := cmd.Execute()

	require.Error(t, err)
	require.Equal(
		t,
		test.ErrorMsg("project remove database error: database_error\n"),
		errOut.String(),
	)

	require.Equal(
		t,
		fmt.Sprintf("%s\n", cmdRemove.UsageString()),
		cmdOut.String(),
	)

	require.NoError(t, mock.ExpectationsWereMet())
}

func testProjectRemoveCmdRowError(
	t *testing.T,
	cmd *cobra.Command,
	service project.ProjectService,
	mock sqlmock.Sqlmock,
) {
	cmdIn := bytes.NewReader([]byte{})
	cmdOut := bytes.NewBufferString("")
	errOut := bytes.NewBufferString("")

	cmdRemove := project.NewCmdProjectRemove(
		project.NewHandlerProjectRemove(service),
	)

	cmd.AddCommand(cmdRemove)
	cmd.SetArgs([]string{"remove", "my_cool_project"})
	cmd.SetIn(cmdIn)
	cmd.SetOut(cmdOut)
	cmd.SetErr(errOut)

	mock.ExpectExec(regexp.QuoteMeta(`
		delete from projects_ where name_ = $name
	`)).WillReturnResult(sqlmock.NewErrorResult(fmt.Errorf("rows_error")))

	err := cmd.Execute()

	require.Error(t, err)
	require.Equal(
		t,
		test.ErrorMsg("rows_error\n"),
		errOut.String(),
	)

	require.Equal(
		t,
		fmt.Sprintf("%s\n", cmdRemove.UsageString()),
		cmdOut.String(),
	)

	require.NoError(t, mock.ExpectationsWereMet())
}

func testProjectRemoveCmdZeroAffectedRows(
	t *testing.T,
	cmd *cobra.Command,
	service project.ProjectService,
	mock sqlmock.Sqlmock,
) {
	cmdIn := bytes.NewReader([]byte{})
	cmdOut := bytes.NewBufferString("")
	errOut := bytes.NewBufferString("")

	cmdRemove := project.NewCmdProjectRemove(
		project.NewHandlerProjectRemove(service),
	)

	cmd.AddCommand(cmdRemove)
	cmd.SetArgs([]string{"remove", "my_cool_project"})
	cmd.SetIn(cmdIn)
	cmd.SetOut(cmdOut)
	cmd.SetErr(errOut)

	mock.ExpectExec(regexp.QuoteMeta(`
		delete from projects_ where name_ = $name
	`)).WillReturnResult(sqlmock.NewResult(0, 0))

	err := cmd.Execute()

	require.Error(t, err)
	require.Equal(
		t,
		test.ErrorMsg("project 'my_cool_project' not found\n"),
		errOut.String(),
	)

	require.Equal(
		t,
		fmt.Sprintf("%s\n", cmdRemove.UsageString()),
		cmdOut.String(),
	)

	require.NoError(t, mock.ExpectationsWereMet())
}

func testProjectRenameCmdHappyPath(
	t *testing.T,
	cmd *cobra.Command,
	service project.ProjectService,
	mock sqlmock.Sqlmock,
) {
	cmdIn := bytes.NewReader([]byte{})
	cmdOut := bytes.NewBufferString("")
	errOut := bytes.NewBufferString("")

	cmdRename := project.NewCmdProjectRename(project.NewHandlerProjectRename(service))
	cmd.AddCommand(cmdRename)

	cmd.SetArgs([]string{"rename", "my_cool_project", "my_awesome_project"})
	cmd.SetIn(cmdIn)
	cmd.SetOut(cmdOut)
	cmd.SetErr(errOut)

	mock.
		ExpectExec(regexp.QuoteMeta(`
			update projects_ set name_ = $newName where name_ = $originalName
		`)).
		WithArgs("my_cool_project", "my_awesome_project").
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := cmd.Execute()

	require.NoError(t, err)
	require.Empty(t, errOut.String())

	require.Equal(
		t,
		test.ProjectRenamedSuccessMsg("my_cool_project", "my_awesome_project"),
		cmdOut.String(),
	)

	require.NoError(t, mock.ExpectationsWereMet())
}

func testProjectRenameCmdWithNoArgs(
	t *testing.T,
	cmd *cobra.Command,
	service project.ProjectService,
	mock sqlmock.Sqlmock,
) {
	cmdIn := bytes.NewReader([]byte{})
	cmdOut := bytes.NewBufferString("")
	errOut := bytes.NewBufferString("")

	cmdRename := project.NewCmdProjectRename(project.NewHandlerProjectRename(service))
	cmd.AddCommand(cmdRename)

	cmd.SetArgs([]string{"rename"})
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

func testProjectRenameCmdWithTooFewArgs(
	t *testing.T,
	cmd *cobra.Command,
	service project.ProjectService,
	mock sqlmock.Sqlmock,
) {
	cmdIn := bytes.NewReader([]byte{})
	cmdOut := bytes.NewBufferString("")
	errOut := bytes.NewBufferString("")

	cmdRename := project.NewCmdProjectRename(project.NewHandlerProjectRename(service))
	cmd.AddCommand(cmdRename)

	cmd.SetArgs([]string{"rename", "my_cool_project"})
	cmd.SetIn(cmdIn)
	cmd.SetOut(cmdOut)
	cmd.SetErr(errOut)

	err := cmd.Execute()

	require.Error(t, err)

	require.Equal(
		t,
		test.IncorrectNumberOfArgsErrorMsg(2, 1),
		errOut.String(),
	)

	require.Equal(
		t,
		fmt.Sprintf("%s\n", cmdRename.UsageString()),
		cmdOut.String(),
	)
}

func testProjectRenameCmdWithTooManyArgs(
	t *testing.T,
	cmd *cobra.Command,
	service project.ProjectService,
	mock sqlmock.Sqlmock,
) {
	cmdIn := bytes.NewReader([]byte{})
	cmdOut := bytes.NewBufferString("")
	errOut := bytes.NewBufferString("")

	cmdRename := project.NewCmdProjectRename(project.NewHandlerProjectRename(service))
	cmd.AddCommand(cmdRename)

	cmd.SetArgs([]string{"rename", "my_cool_project", "my_awesome_project", "my_super_project"})
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

func testProjectRenameCmdDatabaseError(
	t *testing.T,
	cmd *cobra.Command,
	service project.ProjectService,
	mock sqlmock.Sqlmock,
) {
	cmdIn := bytes.NewReader([]byte{})
	cmdOut := bytes.NewBufferString("")
	errOut := bytes.NewBufferString("")

	cmdRename := project.NewCmdProjectRename(project.NewHandlerProjectRename(service))
	cmd.AddCommand(cmdRename)

	cmd.SetArgs([]string{"rename", "my_cool_project", "my_awesome_project"})
	cmd.SetIn(cmdIn)
	cmd.SetOut(cmdOut)
	cmd.SetErr(errOut)

	mock.ExpectExec(regexp.QuoteMeta(`
			update projects_ set name_ = $newName where name_ = $originalName
	`)).WithArgs("my_cool_project", "my_awesome_project").
		WillReturnError(fmt.Errorf("database_error"))

	err := cmd.Execute()

	require.Error(t, err)

	require.Equal(
		t,
		test.ErrorMsg("project rename database error: database_error\n"),
		errOut.String(),
	)

	require.Equal(
		t,
		fmt.Sprintf("%s\n", cmdRename.UsageString()),
		cmdOut.String(),
	)

	require.NoError(t, mock.ExpectationsWereMet())
}

func testProjectListCmdHappyPath(
	t *testing.T,
	cmd *cobra.Command,
	service project.ProjectService,
	mock sqlmock.Sqlmock,
) {
	cmdIn := bytes.NewReader([]byte{})
	cmdOut := bytes.NewBufferString("")
	errOut := bytes.NewBufferString("")

	cmdList := project.NewCmdProjectList(project.NewHandlerProjectList(service))

	cmd.AddCommand(cmdList)

	cmd.SetArgs([]string{"list"})
	cmd.SetIn(cmdIn)
	cmd.SetOut(cmdOut)
	cmd.SetErr(errOut)

	mock.
		ExpectQuery(regexp.QuoteMeta(`select id_, name_ from projects_`)).
		WillReturnRows(
			sqlmock.NewRows([]string{"id_", "name_"}).
				AddRow(1, "my_cool_project").
				AddRow(2, "my_awesome_project").
				AddRow(3, "my_super_project"),
		)

	err := cmd.Execute()

	require.NoError(t, err)
	require.Empty(t, errOut.String())

	require.Equal(
		t,
		cmdOut.String(),
		"my_cool_project\nmy_awesome_project\nmy_super_project\n",
	)

	require.NoError(t, mock.ExpectationsWereMet())
}

func testProjectListCmdZeroResults(
	t *testing.T,
	cmd *cobra.Command,
	service project.ProjectService,
	mock sqlmock.Sqlmock,
) {
	cmdIn := bytes.NewReader([]byte{})
	cmdOut := bytes.NewBufferString("")
	errOut := bytes.NewBufferString("")

	cmd.AddCommand(
		project.NewCmdProjectList(project.NewHandlerProjectList(service)),
	)

	cmd.SetArgs([]string{"list"})
	cmd.SetIn(cmdIn)
	cmd.SetOut(cmdOut)
	cmd.SetErr(errOut)

	mock.
		ExpectQuery(regexp.QuoteMeta(`select id_, name_ from projects_`)).
		WillReturnError(sql.ErrNoRows)

	err := cmd.Execute()

	require.Error(t, err)
	require.Equal(
		t,
		test.ErrorMsg("no projects found\n"),
		errOut.String(),
	)

	require.NoError(t, mock.ExpectationsWereMet())
}

func testProjectListCmdDatabaseError(
	t *testing.T,
	cmd *cobra.Command,
	service project.ProjectService,
	mock sqlmock.Sqlmock,
) {
	cmdIn := bytes.NewReader([]byte{})
	cmdOut := bytes.NewBufferString("")
	errOut := bytes.NewBufferString("")

	cmdList := project.NewCmdProjectList(project.NewHandlerProjectList(service))

	cmd.AddCommand(cmdList)

	cmd.SetArgs([]string{"list"})
	cmd.SetIn(cmdIn)
	cmd.SetOut(cmdOut)
	cmd.SetErr(errOut)

	mock.
		ExpectQuery(regexp.QuoteMeta(`select id_, name_ from projects_`)).
		WillReturnError(fmt.Errorf("database_error"))

	err := cmd.Execute()

	require.Error(t, err)

	require.Equal(
		t,
		test.ErrorMsg("project list database error: database_error\n"),
		errOut.String(),
	)

	require.Equal(
		t,
		fmt.Sprintf("%s\n", cmdList.UsageString()),
		cmdOut.String(),
	)

	require.NoError(t, mock.ExpectationsWereMet())
}

func testProjectListCmdScanError(
	t *testing.T,
	cmd *cobra.Command,
	service project.ProjectService,
	mock sqlmock.Sqlmock,
) {
	cmdIn := bytes.NewReader([]byte{})
	cmdOut := bytes.NewBufferString("")
	errOut := bytes.NewBufferString("")

	cmdList := project.NewCmdProjectList(
		project.NewHandlerProjectList(service),
	)

	cmd.AddCommand(cmdList)

	cmd.SetArgs([]string{"list"})
	cmd.SetIn(cmdIn)
	cmd.SetOut(cmdOut)
	cmd.SetErr(errOut)

	mock.
		ExpectQuery(regexp.QuoteMeta(`select id_, name_ from projects_`)).
		WillReturnRows(
			sqlmock.
				NewRows([]string{"id_", "name_"}).
				AddRow("invalid id", nil),
		)

	err := cmd.Execute()

	require.Error(t, err)
	require.True(
		t,
		strings.HasPrefix(errOut.String(), "Error: sql: Scan error"),
	)

	require.Equal(
		t,
		fmt.Sprintf("%s\n", cmdList.UsageString()),
		cmdOut.String(),
	)

	require.NoError(t, mock.ExpectationsWereMet())
}

func testProjectListCmdWithTooManyArgs(
	t *testing.T,
	cmd *cobra.Command,
	service project.ProjectService,
	mock sqlmock.Sqlmock,
) {
	cmdIn := bytes.NewReader([]byte{})
	cmdOut := bytes.NewBufferString("")
	errOut := bytes.NewBufferString("")

	cmdList := project.NewCmdProjectList(project.NewHandlerProjectList(service))

	cmd.AddCommand(cmdList)

	cmd.SetArgs([]string{"list", "foo"})
	cmd.SetIn(cmdIn)
	cmd.SetOut(cmdOut)
	cmd.SetErr(errOut)

	err := cmd.Execute()

	require.Error(t, err)

	require.Equal(
		t,
		test.UnknownCommandErrorMsg("foo", "project list"),
		errOut.String(),
	)

	require.Equal(
		t,
		fmt.Sprintf("%s\n", cmdList.UsageString()),
		cmdOut.String(),
	)
}

func testProjectAddCmdValidationError(
	t *testing.T,
	cmd *cobra.Command,
	service project.ProjectService,
	mock sqlmock.Sqlmock,
) {
	cmdIn := bytes.NewReader([]byte{})
	cmdOut := bytes.NewBufferString("")
	errOut := bytes.NewBufferString("")

	cmdAdd := project.NewCmdProjectAdd(project.NewHandlerProjectAdd(service))

	cmd.AddCommand(cmdAdd)
	cmd.SetArgs([]string{
		"add",
		"mmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projecty_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_project",
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
		fmt.Sprintf("%s\n", cmdAdd.UsageString()),
		cmdOut.String(),
	)
}

func testProjectRemoveCmdValidationError(
	t *testing.T,
	cmd *cobra.Command,
	service project.ProjectService,
	mock sqlmock.Sqlmock,
) {
	cmdIn := bytes.NewReader([]byte{})
	cmdOut := bytes.NewBufferString("")
	errOut := bytes.NewBufferString("")

	cmdRemove := project.NewCmdProjectRemove(
		project.NewHandlerProjectRemove(service),
	)

	cmd.AddCommand(cmdRemove)

	cmd.SetArgs([]string{
		"remove",
		"my_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_project",
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
		fmt.Sprintf("%s\n", cmdRemove.UsageString()),
		cmdOut.String(),
	)
}

func testProjectRenameCmdValidationError(
	t *testing.T,
	cmd *cobra.Command,
	service project.ProjectService,
	mock sqlmock.Sqlmock,
) {
	cmdIn := bytes.NewReader([]byte{})
	cmdOut := bytes.NewBufferString("")
	errOut := bytes.NewBufferString("")

	cmdRename := project.NewCmdProjectRename(project.NewHandlerProjectRename(service))

	cmd.AddCommand(cmdRename)

	cmd.SetArgs([]string{
		"rename",
		"my_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_project",
		"my_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_project",
	})
	cmd.SetIn(cmdIn)
	cmd.SetOut(cmdOut)
	cmd.SetErr(errOut)

	err := cmd.Execute()

	require.Error(t, err)

	require.Equal(
		t,
		fmt.Sprintf("%s\n", cmdRename.UsageString()),
		cmdOut.String(),
	)
}
