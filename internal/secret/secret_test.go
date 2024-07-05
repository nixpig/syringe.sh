package secret_test

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/charmbracelet/ssh"
	"github.com/nixpig/syringe.sh/internal/secret"
	"github.com/nixpig/syringe.sh/pkg/ctxkeys"
	"github.com/nixpig/syringe.sh/pkg/validation"
	"github.com/nixpig/syringe.sh/test"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	gossh "golang.org/x/crypto/ssh"
)

var mockCrypt = new(MockCrypt)

func TestSecretCmd(t *testing.T) {
	scenarios := map[string]func(
		t *testing.T,
		cmd *cobra.Command,
		service secret.SecretService,
		mock sqlmock.Sqlmock,
	){
		"test secret set command happy path":              testSecretSetCmdHappyPath,
		"test secret set command missing project":         testSecretSetCmdMissingProject,
		"test secret set command missing environment":     testSecretSetCmdMissingEnvironment,
		"test secret set command too few args":            testSecretSetCmdTooFewArgs,
		"test secret set command too many args":           testSecretSetCmdTooManyArgs,
		"test secret set command database error":          testSecretSetCmdDatabaseError,
		"test secret set command validation error":        testSecretSetCmdValidationError,
		"test secret set public key not in context error": testSecretSetCmdPublicKeyNotInContextError,

		"test secret get command happy path":          testSecretGetCmdHappyPath,
		"test secret get command missing project":     testSecretGetCmdMissingProject,
		"test secret get command missing environment": testSecretGetCmdMissingEnvironment,
		"test secret get command missing key":         testSecretGetCmdMissingKey,
		"test secret get command database error":      testSecretGetCmdDatabaseError,
		"test secret get command validation error":    testSecretGetCmdValidationError,

		"test secret list command happy path":          testSecretListCmdHappyPath,
		"test secret list command zero results":        testSecretListCmdZeroResults,
		"test secret list command database error":      testSecretListCmdDatabaseError,
		"test secret list command missing project":     testSecretListCmdMissingProject,
		"test secret list command missing environment": testSecretListCmdMissingEnvironment,
		"test secret list command validation error":    testSecretListCmdValidationError,

		"test secret remove command happy path":          testSecretRemoveCmdHappyPath,
		"test secret remove command zero results":        testSecretRemoveCmdZeroResults,
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

			service := secret.NewSecretServiceImpl(
				secret.NewSqliteSecretStore(db),
				validation.New(),
				mockCrypt,
			)

			cmd := secret.NewCmdSecret()

			fn(t, cmd, service, mock)
		})
	}
}

func testSecretSetCmdHappyPath(
	t *testing.T,
	cmd *cobra.Command,
	service secret.SecretService,
	mock sqlmock.Sqlmock,
) {
	var err error

	cmdIn := bytes.NewReader([]byte{})
	cmdOut := bytes.NewBufferString("")
	errOut := bytes.NewBufferString("")

	cmdSet := secret.NewCmdSecretSet(
		secret.NewHandlerSecretSet(service),
	)

	publicKey, err := generatePublicKey()
	if err != nil {
		t.Error("unable to generate public key")
	}

	ctx := context.Background()
	ctx = context.WithValue(ctx, ctxkeys.PublicKey, publicKey)
	cmdSet.SetContext(ctx)

	cmd.AddCommand(cmdSet)
	cmd.SetArgs([]string{
		"set",
		"-p",
		"my_cool_project",
		"-e",
		"staging",
		"secret_key",
		"secret_value",
	})
	cmd.SetIn(cmdIn)
	cmd.SetOut(cmdOut)
	cmd.SetErr(errOut)

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
			"mock_encrypted_value",
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mockCrypt.On("Encrypt", "secret_value", publicKey).Return("mock_encrypted_value", nil)

	err = cmd.Execute()

	require.NoError(t, err)
	require.Empty(t, errOut.String())

	require.Equal(
		t,
		"", // maybe we want to print out like a 'secret added' message in future?
		cmdOut.String(),
	)

	require.NoError(t, mock.ExpectationsWereMet())
	mockCrypt.AssertExpectations(t)
}

func testSecretSetCmdMissingProject(
	t *testing.T,
	cmd *cobra.Command,
	service secret.SecretService,
	mock sqlmock.Sqlmock,
) {
	cmdIn := bytes.NewReader([]byte{})
	cmdOut := bytes.NewBufferString("")
	errOut := bytes.NewBufferString("")

	cmdSet := secret.NewCmdSecretSet(
		secret.NewHandlerSecretSet(service),
	)

	cmd.AddCommand(cmdSet)
	cmd.SetArgs([]string{
		"set",
		"-e",
		"staging",
		"secret_key",
		"secret_value",
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
		fmt.Sprintf("%s\n", cmdSet.UsageString()),
		cmdOut.String(),
	)
}

func testSecretSetCmdMissingEnvironment(
	t *testing.T,
	cmd *cobra.Command,
	service secret.SecretService,
	mock sqlmock.Sqlmock,
) {
	cmdIn := bytes.NewReader([]byte{})
	cmdOut := bytes.NewBufferString("")
	errOut := bytes.NewBufferString("")

	cmdSet := secret.NewCmdSecretSet(
		secret.NewHandlerSecretSet(service),
	)

	cmd.AddCommand(cmdSet)
	cmd.SetArgs([]string{
		"set",
		"-p",
		"my_cool_project",
		"secret_key",
		"secret_value",
	})
	cmd.SetIn(cmdIn)
	cmd.SetOut(cmdOut)
	cmd.SetErr(errOut)

	err := cmd.Execute()

	require.Error(t, err)

	require.Equal(
		t,
		test.RequiredFlagsErrorMsg("environment"),
		errOut.String(),
	)

	require.Equal(
		t,
		fmt.Sprintf("%s\n", cmdSet.UsageString()),
		cmdOut.String(),
	)
}

func testSecretSetCmdTooFewArgs(
	t *testing.T,
	cmd *cobra.Command,
	service secret.SecretService,
	mock sqlmock.Sqlmock,
) {
	cmdIn := bytes.NewReader([]byte{})
	cmdOut := bytes.NewBufferString("")
	errOut := bytes.NewBufferString("")

	cmdSet := secret.NewCmdSecretSet(
		secret.NewHandlerSecretSet(service),
	)

	cmd.AddCommand(cmdSet)
	cmd.SetArgs([]string{
		"set",
		"-p",
		"my_cool_project",
		"-e",
		"staging",
		"secret_key",
	})
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
		fmt.Sprintf("%s\n", cmdSet.UsageString()),
		cmdOut.String(),
	)

}

func testSecretSetCmdTooManyArgs(
	t *testing.T,
	cmd *cobra.Command,
	service secret.SecretService,
	mock sqlmock.Sqlmock,
) {
	cmdIn := bytes.NewReader([]byte{})
	cmdOut := bytes.NewBufferString("")
	errOut := bytes.NewBufferString("")

	cmdSet := secret.NewCmdSecretSet(
		secret.NewHandlerSecretSet(service),
	)

	cmd.AddCommand(cmdSet)
	cmd.SetArgs([]string{
		"set",
		"-p",
		"my_cool_project",
		"-e",
		"staging",
		"secret_key",
		"secret_value",
		"to_many",
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
		fmt.Sprintf("%s\n", cmdSet.UsageString()),
		cmdOut.String(),
	)
}

func testSecretSetCmdDatabaseError(
	t *testing.T,
	cmd *cobra.Command,
	service secret.SecretService,
	mock sqlmock.Sqlmock,
) {
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
			"mock_encrypted_value",
		).
		WillReturnError(fmt.Errorf("database_error"))

	cmdSet := secret.NewCmdSecretSet(
		secret.NewHandlerSecretSet(service),
	)

	cmd.AddCommand(cmdSet)
	cmd.SetArgs([]string{
		"set",
		"-p",
		"my_cool_project",
		"-e",
		"staging",
		"secret_key",
		"secret_value",
	})
	cmd.SetIn(cmdIn)
	cmd.SetOut(cmdOut)
	cmd.SetErr(errOut)

	publicKey, err := generatePublicKey()
	if err != nil {
		t.Error("unable to generate public key")
	}

	mockCrypt.On("Encrypt", "secret_value", publicKey).Return("mock_encrypted_value", nil)

	ctx := context.Background()
	ctx = context.WithValue(ctx, ctxkeys.PublicKey, publicKey)
	cmdSet.SetContext(ctx)

	err = cmd.Execute()

	require.Error(t, err)

	require.Equal(
		t,
		test.ErrorMsg("database exec error\n"),
		errOut.String(),
	)

	require.Equal(
		t,
		fmt.Sprintf("%s\n", cmdSet.UsageString()),
		cmdOut.String(),
	)

	require.NoError(t, mock.ExpectationsWereMet())

	mockCrypt.AssertExpectations(t)
}

func testSecretSetCmdValidationError(
	t *testing.T,
	cmd *cobra.Command,
	service secret.SecretService,
	mock sqlmock.Sqlmock,
) {
	cmdIn := bytes.NewReader([]byte{})
	cmdOut := bytes.NewBufferString("")
	errOut := bytes.NewBufferString("")

	cmdSet := secret.NewCmdSecretSet(
		secret.NewHandlerSecretSet(service),
	)

	cmd.AddCommand(cmdSet)
	cmd.SetArgs([]string{
		"set",
		"-p",
		"my_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_project",
		"-e",
		"sstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingtagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstaging",
		"ssecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keyecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_key",
		"ssecret_valuesecret_valuesecret_valuesecret_valuesecret_valuesecret_valuesecret_valuesecret_valuesecret_valuesecret_valuesecret_valuesecret_valuesecret_valuesecret_valuesecret_valuesecret_valuesecret_valuesecret_valuesecret_valuesecret_valuesecret_valuesecret_valuesecret_valuesecret_valuesecret_valuesecret_valuesecret_valuesecret_valuesecret_valuesecret_valuesecret_valuesecret_valuesecret_valuesecret_valuesecret_valuesecret_valuesecret_valuesecret_valuesecret_valuesecret_valuesecret_valuesecret_valuesecret_valuesecret_valuesecret_valuesecret_valuesecret_valuesecret_valuesecret_valuesecret_valuesecret_valuesecret_valuesecret_valuesecret_valuesecret_valuesecret_valuesecret_valuesecret_valuesecret_valuesecret_valueecret_valuesecret_valuesecret_valuesecret_valuesecret_valuesecret_valuesecret_valuesecret_valuesecret_valuesecret_valuesecret_valuesecret_valuesecret_valuesecret_valuesecret_valuesecret_valuesecret_valuesecret_valuesecret_valuesecret_valuesecret_valuesecret_valuesecret_valuesecret_valuesecret_valuesecret_valuesecret_valuesecret_valuesecret_valuesecret_value",
	})
	cmd.SetIn(cmdIn)
	cmd.SetOut(cmdOut)
	cmd.SetErr(errOut)

	publicKey, err := generatePublicKey()
	if err != nil {
		t.Error("unable to generate public key")
	}

	ctx := context.Background()
	ctx = context.WithValue(ctx, ctxkeys.PublicKey, publicKey)
	cmd.SetContext(ctx)

	err = cmd.Execute()

	require.Error(t, err)

	require.Equal(
		t,
		fmt.Sprintf("%s\n", cmdSet.UsageString()),
		cmdOut.String(),
	)
}

func testSecretGetCmdHappyPath(
	t *testing.T,
	cmd *cobra.Command,
	service secret.SecretService,
	mock sqlmock.Sqlmock,
) {
	cmdIn := bytes.NewReader([]byte{})
	cmdOut := bytes.NewBufferString("")
	errOut := bytes.NewBufferString("")

	cmdGet := secret.NewCmdSecretGet(
		secret.NewHandlerSecretGet(service),
	)

	cmd.AddCommand(cmdGet)
	cmd.SetArgs([]string{
		"get",
		"-p",
		"my_cool_project",
		"-e",
		"staging",
		"secret_key",
	})
	cmd.SetIn(cmdIn)
	cmd.SetOut(cmdOut)
	cmd.SetErr(errOut)

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

	err := cmd.Execute()

	require.NoError(t, err)
	require.Empty(t, errOut.String())

	require.Equal(
		t,
		"secret_value",
		cmdOut.String(),
	)

	require.NoError(t, mock.ExpectationsWereMet())
}

func testSecretGetCmdMissingProject(
	t *testing.T,
	cmd *cobra.Command,
	service secret.SecretService,
	mock sqlmock.Sqlmock,
) {
	cmdIn := bytes.NewReader([]byte{})
	cmdOut := bytes.NewBufferString("")
	errOut := bytes.NewBufferString("")

	cmdGet := secret.NewCmdSecretGet(
		secret.NewHandlerSecretGet(service),
	)

	cmd.AddCommand(cmdGet)
	cmd.SetArgs([]string{
		"get",
		"-e",
		"staging",
		"secret_key",
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
		fmt.Sprintf("%s\n", cmdGet.UsageString()),
		cmdOut.String(),
	)
}

func testSecretGetCmdMissingEnvironment(
	t *testing.T,
	cmd *cobra.Command,
	service secret.SecretService,
	mock sqlmock.Sqlmock,
) {
	cmdIn := bytes.NewReader([]byte{})
	cmdOut := bytes.NewBufferString("")
	errOut := bytes.NewBufferString("")

	cmdGet := secret.NewCmdSecretGet(
		secret.NewHandlerSecretGet(service),
	)

	cmd.AddCommand(cmdGet)
	cmd.SetArgs([]string{
		"get",
		"-p",
		"my_cool_project",
		"secret_key",
	})
	cmd.SetIn(cmdIn)
	cmd.SetOut(cmdOut)
	cmd.SetErr(errOut)

	err := cmd.Execute()

	require.Error(t, err)

	require.Equal(
		t,
		test.RequiredFlagsErrorMsg("environment"),
		errOut.String(),
	)

	require.Equal(
		t,
		fmt.Sprintf("%s\n", cmdGet.UsageString()),
		cmdOut.String(),
	)
}

func testSecretGetCmdMissingKey(
	t *testing.T,
	cmd *cobra.Command,
	service secret.SecretService,
	mock sqlmock.Sqlmock,
) {
	cmdIn := bytes.NewReader([]byte{})
	cmdOut := bytes.NewBufferString("")
	errOut := bytes.NewBufferString("")

	cmdGet := secret.NewCmdSecretGet(
		secret.NewHandlerSecretGet(service),
	)

	cmd.AddCommand(cmdGet)
	cmd.SetArgs([]string{
		"get",
		"-p",
		"my_cool_project",
		"-e",
		"staging",
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
		fmt.Sprintf("%s\n", cmdGet.UsageString()),
		cmdOut.String(),
	)
}

func testSecretGetCmdDatabaseError(
	t *testing.T,
	cmd *cobra.Command,
	service secret.SecretService,
	mock sqlmock.Sqlmock,
) {
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

	cmdGet := secret.NewCmdSecretGet(
		secret.NewHandlerSecretGet(service),
	)

	cmd.AddCommand(cmdGet)
	cmd.SetArgs([]string{
		"get",
		"-p",
		"my_cool_project",
		"-e",
		"staging",
		"secret_key",
	})
	cmd.SetIn(cmdIn)
	cmd.SetOut(cmdOut)
	cmd.SetErr(errOut)

	err := cmd.Execute()

	require.Error(t, err)

	require.Equal(
		t,
		test.ErrorMsg("database exec error\n"),
		errOut.String(),
	)

	require.Equal(
		t,
		fmt.Sprintf("%s\n", cmdGet.UsageString()),
		cmdOut.String(),
	)

	require.NoError(t, mock.ExpectationsWereMet())
}

func testSecretGetCmdValidationError(
	t *testing.T,
	cmd *cobra.Command,
	service secret.SecretService,
	mock sqlmock.Sqlmock,
) {
	cmdIn := bytes.NewReader([]byte{})
	cmdOut := bytes.NewBufferString("")
	errOut := bytes.NewBufferString("")

	cmdGet := secret.NewCmdSecretGet(
		secret.NewHandlerSecretGet(service),
	)

	cmd.AddCommand(cmdGet)
	cmd.SetArgs([]string{
		"get",
		"-p",
		"my_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_project",
		"-e",
		"stagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstagingstaging",
		"secret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_keysecret_key",
	})
	cmd.SetIn(cmdIn)
	cmd.SetOut(cmdOut)
	cmd.SetErr(errOut)

	err := cmd.Execute()

	require.Error(t, err)

	require.Equal(
		t,
		fmt.Sprintf("%s\n", cmdGet.UsageString()),
		cmdOut.String(),
	)
}

func testSecretListCmdHappyPath(
	t *testing.T,
	cmd *cobra.Command,
	service secret.SecretService,
	mock sqlmock.Sqlmock,
) {
	cmdIn := bytes.NewReader([]byte{})
	cmdOut := bytes.NewBufferString("")
	errOut := bytes.NewBufferString("")

	cmdList := secret.NewCmdSecretList(
		secret.NewHandlerSecretList(service),
	)

	cmd.AddCommand(cmdList)
	cmd.SetArgs([]string{
		"list",
		"-p",
		"my_cool_project",
		"-e",
		"staging",
	})
	cmd.SetIn(cmdIn)
	cmd.SetOut(cmdOut)
	cmd.SetErr(errOut)

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

	err := cmd.Execute()

	require.NoError(t, err)
	require.Empty(t, errOut.String())

	require.Equal(
		t,
		"key_1=value_1\nkey_2=value_2",
		cmdOut.String(),
	)

	require.NoError(t, mock.ExpectationsWereMet())
}

func testSecretRemoveCmdHappyPath(
	t *testing.T,
	cmd *cobra.Command,
	service secret.SecretService,
	mock sqlmock.Sqlmock,
) {
	cmdIn := bytes.NewReader([]byte{})
	cmdOut := bytes.NewBufferString("")
	errOut := bytes.NewBufferString("")

	cmdRemove := secret.NewCmdSecretRemove(
		secret.NewHandlerSecretRemove(service),
	)

	cmd.AddCommand(cmdRemove)
	cmd.SetArgs([]string{
		"remove",
		"-p",
		"my_cool_project",
		"-e",
		"staging",
		"key_1",
	})
	cmd.SetIn(cmdIn)
	cmd.SetOut(cmdOut)
	cmd.SetErr(errOut)

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

	err := cmd.Execute()

	require.NoError(t, err)
	require.Empty(t, errOut.String())

	require.Equal(
		t,
		"", // maybe we want to execute a 'success' message or something in future
		cmdOut.String(),
	)

	require.NoError(t, mock.ExpectationsWereMet())
}

func testSecretListCmdDatabaseError(
	t *testing.T,
	cmd *cobra.Command,
	service secret.SecretService,
	mock sqlmock.Sqlmock,
) {
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
	`

	mock.ExpectQuery(regexp.QuoteMeta(query)).WithArgs(
		"my_cool_project", "staging",
	).WillReturnError(errors.New("database_error"))

	cmdList := secret.NewCmdSecretList(
		secret.NewHandlerSecretList(service),
	)

	cmd.AddCommand(cmdList)
	cmd.SetArgs([]string{
		"list",
		"-p",
		"my_cool_project",
		"-e",
		"staging",
	})
	cmd.SetIn(cmdIn)
	cmd.SetOut(cmdOut)
	cmd.SetErr(errOut)

	err := cmd.Execute()

	require.Error(t, err)

	require.Equal(
		t,
		test.ErrorMsg("database query error\n"),
		errOut.String(),
	)

	require.Equal(
		t,
		fmt.Sprintf("%s\n", cmdList.UsageString()),
		cmdOut.String(),
	)

	require.NoError(t, mock.ExpectationsWereMet())
}

func testSecretListCmdMissingProject(
	t *testing.T,
	cmd *cobra.Command,
	service secret.SecretService,
	mock sqlmock.Sqlmock,
) {
	cmdIn := bytes.NewReader([]byte{})
	cmdOut := bytes.NewBufferString("")
	errOut := bytes.NewBufferString("")

	cmdList := secret.NewCmdSecretList(
		secret.NewHandlerSecretList(service),
	)

	cmd.AddCommand(cmdList)
	cmd.SetArgs([]string{
		"list",
		"-e",
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
		fmt.Sprintf("%s\n", cmdList.UsageString()),
		cmdOut.String(),
	)
}

func testSecretListCmdMissingEnvironment(
	t *testing.T,
	cmd *cobra.Command,
	service secret.SecretService,
	mock sqlmock.Sqlmock,
) {
	cmdIn := bytes.NewReader([]byte{})
	cmdOut := bytes.NewBufferString("")
	errOut := bytes.NewBufferString("")

	cmdList := secret.NewCmdSecretList(
		secret.NewHandlerSecretList(service),
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
		test.RequiredFlagsErrorMsg("environment"),
		errOut.String(),
	)

	require.Equal(
		t,
		fmt.Sprintf("%s\n", cmdList.UsageString()),
		cmdOut.String(),
	)
}

func testSecretListCmdValidationError(
	t *testing.T,
	cmd *cobra.Command,
	service secret.SecretService,
	mock sqlmock.Sqlmock,
) {
	var cmdIn *bytes.Reader
	var cmdOut *bytes.Buffer
	var errOut *bytes.Buffer
	var cmdList *cobra.Command
	var err error

	cmdIn = bytes.NewReader([]byte{})
	cmdOut = bytes.NewBufferString("")
	errOut = bytes.NewBufferString("")

	cmdList = secret.NewCmdSecretList(
		secret.NewHandlerSecretList(service),
	)

	cmd.AddCommand(cmdList)
	cmd.SetArgs([]string{
		"list",
		"-p",
		"my_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_projectmy_cool_project",
		"-e",
		"staging",
	})
	cmd.SetIn(cmdIn)
	cmd.SetOut(cmdOut)
	cmd.SetErr(errOut)

	err = cmd.Execute()

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

	cmd.RemoveCommand(cmdList)

	cmdIn = bytes.NewReader([]byte{})
	cmdOut = bytes.NewBufferString("")
	errOut = bytes.NewBufferString("")

	cmdList = secret.NewCmdSecretList(
		secret.NewHandlerSecretList(service),
	)

	cmd.AddCommand(cmdList)
	cmd.SetArgs([]string{
		"list",
		"-p",
		"my_cool_project",
		"-e",
		"staging_staging_staging_staging_staging_staging_staging_staging_staging_staging_staging_staging_staging_staging_staging_staging_staging_staging_staging_staging_staging_staging_staging_staging_staging_staging_staging_staging_staging_staging_staging_staging_staging",
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
		fmt.Sprintf("%s\n", cmdList.UsageString()),
		cmdOut.String(),
	)
}

func testSecretRemoveCmdDatabaseError(
	t *testing.T,
	cmd *cobra.Command,
	service secret.SecretService,
	mock sqlmock.Sqlmock,
) {
	cmdIn := bytes.NewReader([]byte{})
	cmdOut := bytes.NewBufferString("")
	errOut := bytes.NewBufferString("")

	cmdRemove := secret.NewCmdSecretRemove(
		secret.NewHandlerSecretRemove(service),
	)

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

	cmd.AddCommand(cmdRemove)
	cmd.SetArgs([]string{
		"remove",
		"-p",
		"my_cool_project",
		"-e",
		"staging",
		"key_1",
	})
	cmd.SetIn(cmdIn)
	cmd.SetOut(cmdOut)
	cmd.SetErr(errOut)

	err := cmd.Execute()

	require.Error(t, err)

	require.Equal(
		t,
		test.ErrorMsg("database exec error\n"),
		errOut.String(),
	)

	require.Equal(
		t,
		fmt.Sprintf("%s\n", cmdRemove.UsageString()),
		cmdOut.String(),
	)

	require.NoError(t, mock.ExpectationsWereMet())
}

func testSecretRemoveCmdMissingProject(
	t *testing.T,
	cmd *cobra.Command,
	service secret.SecretService,
	mock sqlmock.Sqlmock,
) {
	cmdIn := bytes.NewReader([]byte{})
	cmdOut := bytes.NewBufferString("")
	errOut := bytes.NewBufferString("")

	cmdRemove := secret.NewCmdSecretRemove(
		secret.NewHandlerSecretRemove(service),
	)

	cmd.AddCommand(cmdRemove)
	cmd.SetArgs([]string{
		"remove",
		"-e",
		"staging",
		"key_1",
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
}

func testSecretRemoveCmdMissingEnvironment(
	t *testing.T,
	cmd *cobra.Command,
	service secret.SecretService,
	mock sqlmock.Sqlmock,
) {
	cmdIn := bytes.NewReader([]byte{})
	cmdOut := bytes.NewBufferString("")
	errOut := bytes.NewBufferString("")

	cmdRemove := secret.NewCmdSecretRemove(
		secret.NewHandlerSecretRemove(service),
	)

	cmd.AddCommand(cmdRemove)
	cmd.SetArgs([]string{
		"remove",
		"-p",
		"my_cool_project",
		"key_1",
	})
	cmd.SetIn(cmdIn)
	cmd.SetOut(cmdOut)
	cmd.SetErr(errOut)

	err := cmd.Execute()

	require.Error(t, err)

	require.Equal(
		t,
		test.RequiredFlagsErrorMsg("environment"),
		errOut.String(),
	)

	require.Equal(
		t,
		fmt.Sprintf("%s\n", cmdRemove.UsageString()),
		cmdOut.String(),
	)
}

func testSecretRemoveCmdMissingKey(
	t *testing.T,
	cmd *cobra.Command,
	service secret.SecretService,
	mock sqlmock.Sqlmock,
) {
	cmdIn := bytes.NewReader([]byte{})
	cmdOut := bytes.NewBufferString("")
	errOut := bytes.NewBufferString("")

	cmdRemove := secret.NewCmdSecretRemove(
		secret.NewHandlerSecretRemove(service),
	)

	cmd.AddCommand(cmdRemove)
	cmd.SetArgs([]string{
		"remove",
		"-p",
		"my_cool_project",
		"-e",
		"staging",
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
}

func testSecretRemoveCmdValidationError(
	t *testing.T,
	cmd *cobra.Command,
	service secret.SecretService,
	mock sqlmock.Sqlmock,
) {
	var cmdIn *bytes.Reader
	var cmdOut *bytes.Buffer
	var errOut *bytes.Buffer
	var cmdRemove *cobra.Command
	var err error

	cmdIn = bytes.NewReader([]byte{})
	cmdOut = bytes.NewBufferString("")
	errOut = bytes.NewBufferString("")

	cmdRemove = secret.NewCmdSecretRemove(
		secret.NewHandlerSecretRemove(service),
	)

	cmd.AddCommand(cmdRemove)
	cmd.SetArgs([]string{
		"remove",
		"-p",
		"my_cool_project_my_cool_project_my_cool_project_my_cool_project_my_cool_project_my_cool_project_my_cool_project_my_cool_project_my_cool_project_my_cool_project_my_cool_project_my_cool_project_my_cool_project_my_cool_project_my_cool_project_my_cool_project_my_cool_project_my_cool_project_",
		"-e",
		"staging",
		"key_1",
	})
	cmd.SetIn(cmdIn)
	cmd.SetOut(cmdOut)
	cmd.SetErr(errOut)

	err = cmd.Execute()

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

	cmd.RemoveCommand(cmdRemove)

	cmdIn = bytes.NewReader([]byte{})
	cmdOut = bytes.NewBufferString("")
	errOut = bytes.NewBufferString("")

	cmdRemove = secret.NewCmdSecretRemove(
		secret.NewHandlerSecretRemove(service),
	)

	cmd.AddCommand(cmdRemove)
	cmd.SetArgs([]string{
		"remove",
		"-p",
		"my_cool_project",
		"-e",
		"staging_staging_staging_staging_staging_staging_staging_staging_staging_staging_staging_staging_staging_staging_staging_staging_staging_staging_staging_staging_staging_staging_staging_staging_staging_staging_staging_staging_staging_staging_staging_staging_staging_staging_staging_",
		"key_1",
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
		fmt.Sprintf("%s\n", cmdRemove.UsageString()),
		cmdOut.String(),
	)
}

func testSecretListCmdZeroResults(
	t *testing.T,
	cmd *cobra.Command,
	service secret.SecretService,
	mock sqlmock.Sqlmock,
) {
	cmdIn := bytes.NewReader([]byte{})
	cmdOut := bytes.NewBufferString("")
	errOut := bytes.NewBufferString("")

	cmdList := secret.NewCmdSecretList(
		secret.NewHandlerSecretList(service),
	)

	cmd.AddCommand(cmdList)
	cmd.SetArgs([]string{
		"list",
		"-p",
		"my_cool_project",
		"-e",
		"staging",
	})
	cmd.SetIn(cmdIn)
	cmd.SetOut(cmdOut)
	cmd.SetErr(errOut)

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
	`

	mock.
		ExpectQuery(regexp.QuoteMeta(query)).
		WithArgs("my_cool_project", "staging").
		WillReturnError(sql.ErrNoRows)

	err := cmd.Execute()

	require.Error(t, err)

	require.Equal(
		t,
		test.ErrorMsg("no secrets found\n"),
		errOut.String(),
	)

	require.Equal(
		t,
		fmt.Sprintf("%s\n", cmdList.UsageString()),
		cmdOut.String(),
	)

	require.NoError(t, mock.ExpectationsWereMet())
}

func testSecretRemoveCmdZeroResults(
	t *testing.T,
	cmd *cobra.Command,
	service secret.SecretService,
	mock sqlmock.Sqlmock,
) {
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
	).WillReturnResult(sqlmock.NewResult(0, 0))

	cmdRemove := secret.NewCmdSecretRemove(
		secret.NewHandlerSecretRemove(service),
	)

	cmd.AddCommand(cmdRemove)
	cmd.SetArgs([]string{
		"remove",
		"-p",
		"my_cool_project",
		"-e",
		"staging",
		"key_1",
	})
	cmd.SetIn(cmdIn)
	cmd.SetOut(cmdOut)
	cmd.SetErr(errOut)

	err := cmd.Execute()

	require.Error(t, err)

	require.Equal(
		t,
		test.ErrorMsg("secret not found\n"),
		errOut.String(),
	)

	require.Equal(
		t,
		fmt.Sprintf("%s\n", cmdRemove.UsageString()),
		cmdOut.String(),
	)

	require.NoError(t, mock.ExpectationsWereMet())
}

func testSecretSetCmdPublicKeyNotInContextError(
	t *testing.T,
	cmd *cobra.Command,
	service secret.SecretService,
	mock sqlmock.Sqlmock,
) {
	var err error

	cmdIn := bytes.NewReader([]byte{})
	cmdOut := bytes.NewBufferString("")
	errOut := bytes.NewBufferString("")

	cmdSet := secret.NewCmdSecretSet(
		secret.NewHandlerSecretSet(service),
	)

	cmd.AddCommand(cmdSet)
	cmd.SetArgs([]string{
		"set",
		"-p",
		"my_cool_project",
		"-e",
		"staging",
		"secret_key",
		"secret_value",
	})
	cmd.SetIn(cmdIn)
	cmd.SetOut(cmdOut)
	cmd.SetErr(errOut)

	err = cmd.Execute()

	require.Error(t, err)
	require.Equal(
		t,
		test.ErrorMsg("unable to get public key from context\n"),
		errOut.String(),
	)

	require.Equal(
		t,
		fmt.Sprintf("%s\n", cmdSet.UsageString()),
		cmdOut.String(),
	)

	require.NoError(t, mock.ExpectationsWereMet())
	mockCrypt.AssertExpectations(t)
}

func generatePublicKey() (ssh.PublicKey, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return nil, err
	}

	publicKey, err := gossh.NewPublicKey(&privateKey.PublicKey)
	if err != nil {
		return nil, err
	}

	charmPublicKey, ok := publicKey.(ssh.PublicKey)
	if !ok {
		return nil, errors.New("failed to cast public key")
	}

	return charmPublicKey, err
}

type MockCrypt struct {
	mock.Mock
}

func (c *MockCrypt) Encrypt(secret string, publicKey ssh.PublicKey) (string, error) {
	args := c.Called(secret, publicKey)

	return args.String(0), args.Error(1)
}

func (c *MockCrypt) Decrypt(cypherText string, privateKey *rsa.PrivateKey) (string, error) {
	args := c.Called(cypherText, privateKey)

	return args.String(0), args.Error(1)
}
