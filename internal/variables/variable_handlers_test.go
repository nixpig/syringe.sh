package internal

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockVariableStore struct {
	mock.Mock
}

func (s *MockVariableStore) Set(variable Variable) error {
	args := s.Called(variable)

	return args.Error(0)
}

func (s *MockVariableStore) Get(projectName, environmentName, key string) (string, error) {
	args := s.Called(projectName, environmentName, key)

	return args.Get(0).(string), args.Error(1)
}

type MockValidator struct {
	mock.Mock
}

func (v *MockValidator) Struct(s interface{}) error {
	args := v.Called(s)

	return args.Error(0)
}

var mockStore = new(MockVariableStore)
var mockValidator = new(MockValidator)

func TestVariableHandlers(t *testing.T) {
	scenarios := map[string]func(t *testing.T, handler VariableHandler){
		"test variable handler set variable (success)":            testVariableHandlerSetVariableSuccess,
		"test variable handler set variable (error - validation)": testVariableHandlerSetVariableValidationError,
		"test variable handler set variable (error - store)":      testVariableHandlerSetVariableStoreError,
	}

	for scenario, fn := range scenarios {
		t.Run(scenario, func(t *testing.T) {

			handler := NewVariableCliHandler(mockStore, mockValidator)

			fn(t, handler)
		})
	}
}

func testVariableHandlerSetVariableSuccess(t *testing.T, handler VariableHandler) {
	variable := Variable{
		ProjectName:     "project_name",
		EnvironmentName: "environment_name",
		Key:             "VAR_KEY",
		Value:           "var_value",
		Secret:          true,
	}

	mockValidatorStruct := mockValidator.On("Struct", variable).Return(nil)
	mockStoreSet := mockStore.On("Set", variable).Return(nil)

	err := handler.Set(
		variable.ProjectName,
		variable.EnvironmentName,
		variable.Key,
		variable.Value,
		variable.Secret,
	)

	require.NoError(t, err, "should not return error")

	if expectations := mockValidator.AssertExpectations(t); !expectations {
		t.Error("validator was not called as expected")
	}

	if expectations := mockStore.AssertExpectations(t); !expectations {
		t.Error("store was not called as expected")
	}

	mockValidatorStruct.Unset()
	mockStoreSet.Unset()
}

func testVariableHandlerSetVariableValidationError(t *testing.T, handler VariableHandler) {
	variable := Variable{
		ProjectName:     "project_name",
		EnvironmentName: "environment_name",
		Key:             "VAR_KEY",
		Value:           "var_value",
		Secret:          true,
	}

	mockValidatorStruct := mockValidator.
		On("Struct", variable).
		Return(errors.New("validation_error"))

	err := handler.Set(
		variable.ProjectName,
		variable.EnvironmentName,
		variable.Key,
		variable.Value,
		variable.Secret,
	)

	require.EqualError(t, err, "validation_error", "should return validation error")

	if expectations := mockValidator.AssertExpectations(t); !expectations {
		t.Error("validator was not called as expected")
	}

	if expectations := mockStore.AssertNotCalled(t, mock.Anything, mock.Anything); !expectations {
		t.Error("unexpected call to store")
	}

	mockValidatorStruct.Unset()
}

func testVariableHandlerSetVariableStoreError(t *testing.T, handler VariableHandler) {
	variable := Variable{
		ProjectName:     "project_name",
		EnvironmentName: "environment_name",
		Key:             "VAR_KEY",
		Value:           "var_value",
		Secret:          true,
	}

	mockValidatorStruct := mockValidator.
		On("Struct", variable).
		Return(nil)

	mockStoreSet := mockStore.On("Set", variable).Return(errors.New("store_error"))

	err := handler.Set(
		variable.ProjectName,
		variable.EnvironmentName,
		variable.Key,
		variable.Value,
		variable.Secret,
	)

	require.EqualError(t, err, "store_error", "should return store error")

	if expectations := mockValidator.AssertExpectations(t); !expectations {
		t.Error("validatoe was not called as expected")
	}

	if expectations := mockStore.AssertExpectations(t); !expectations {
		t.Error("store was not called as expected")
	}

	mockValidatorStruct.Unset()
	mockStoreSet.Unset()
}
