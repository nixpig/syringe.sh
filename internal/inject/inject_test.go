package inject_test

import (
	"bytes"
	"errors"
	"testing"

	"github.com/nixpig/syringe.sh/internal/inject"
	"github.com/nixpig/syringe.sh/internal/secret"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var mockSecretService = new(MockSecretService)

func TestInjectCmd(t *testing.T) {
	scenarios := map[string]func(
		t *testing.T,
		cmd *cobra.Command,
		service secret.SecretService,
	){
		"test inject cmd happy path":    testInjectCmdHappyPath,
		"test inject cmd service error": testInjectCmdServiceError,
	}

	for scenario, fn := range scenarios {
		t.Run(scenario, func(t *testing.T) {
			cmd := &cobra.Command{
				Use: "test",
			}

			fn(t, cmd, mockSecretService)
		})

	}
}

func testInjectCmdHappyPath(
	t *testing.T,
	cmd *cobra.Command,
	service secret.SecretService,
) {
	cmdOut := bytes.NewBufferString("")
	errOut := bytes.NewBufferString("")
	handler := inject.NewHandlerInject(service)
	cmdInject := inject.NewCmdInject(handler)
	cmdInject.SetOut(cmdOut)
	cmdInject.SetErr(errOut)

	cmd.AddCommand(cmdInject)
	cmd.SetArgs([]string{
		"inject",
		"-p",
		"my_cool_project",
		"-e",
		"staging",
		"--",
		"startserver",
	})

	m := mockSecretService.On("List", secret.ListSecretsRequest{
		Project:     "my_cool_project",
		Environment: "staging",
	}).Return(
		&secret.ListSecretsResponse{
			Project:     "my_cool_project",
			Environment: "staging",
			Secrets: []struct {
				ID    int
				Key   string
				Value string
			}{
				{
					ID:    23,
					Key:   "SECRET_KEY_1",
					Value: "SECRET_VALUE_1",
				},
				{
					ID:    69,
					Key:   "SECRET_KEY_2",
					Value: "SECRET_VALUE_2",
				},
			},
		}, nil,
	)

	err := cmd.Execute()

	require.NoError(t, err)
	require.Empty(t, errOut.String())

	require.Equal(
		t,
		"SECRET_KEY_1=SECRET_VALUE_1 SECRET_KEY_2=SECRET_VALUE_2\n",
		cmdOut.String(),
	)

	if m := mockSecretService.AssertExpectations(t); !m {
		t.Error("did not receive expected calls to service")
	}

	mockSecretService.AssertCalled(t, "List", secret.ListSecretsRequest{
		Project:     "my_cool_project",
		Environment: "staging",
	})

	m.Unset()
}

func testInjectCmdServiceError(
	t *testing.T,
	cmd *cobra.Command,
	service secret.SecretService,
) {
	cmdOut := bytes.NewBufferString("")
	errOut := bytes.NewBufferString("")
	handler := inject.NewHandlerInject(service)
	cmdInject := inject.NewCmdInject(handler)
	cmdInject.SetOut(cmdOut)
	cmdInject.SetErr(errOut)

	cmd.AddCommand(cmdInject)
	cmd.SetArgs([]string{
		"inject",
		"-p",
		"my_cool_project",
		"-e",
		"staging",
		"--",
		"startserver",
	})

	m := mockSecretService.On("List", secret.ListSecretsRequest{
		Project:     "my_cool_project",
		Environment: "staging",
	}).Return(
		&secret.ListSecretsResponse{},
		errors.New("secret_service_error"),
	)

	err := cmd.Execute()

	require.Error(t, err)

	require.Equal(
		t,
		"",
		cmdOut.String(),
	)

	mockSecretService.AssertCalled(t, "List", secret.ListSecretsRequest{
		Project:     "my_cool_project",
		Environment: "staging",
	})

	m.Unset()
}

type MockSecretService struct {
	mock.Mock
}

func (m *MockSecretService) CreateTables() error {
	args := m.Called()

	return args.Error(0)
}

func (m *MockSecretService) Set(secret secret.SetSecretRequest) error {
	args := m.Called(secret)

	return args.Error(0)
}

func (m *MockSecretService) Get(request secret.GetSecretRequest) (*secret.GetSecretResponse, error) {
	args := m.Called(request)

	return args.Get(0).(*secret.GetSecretResponse), args.Error(1)
}

func (m *MockSecretService) List(request secret.ListSecretsRequest) (*secret.ListSecretsResponse, error) {
	args := m.Called(request)

	return args.Get(0).(*secret.ListSecretsResponse), args.Error(1)
}

func (m *MockSecretService) Remove(request secret.RemoveSecretRequest) error {
	args := m.Called(request)

	return args.Error(0)
}
