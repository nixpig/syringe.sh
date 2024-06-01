package internal

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var mockStore = new(MockVariableStore)
var mockValidator = new(MockValidator)

func TestVariableHandlers(t *testing.T) {
	scenarios := map[string]func(t *testing.T, handler VariableHandler){
		"test variable handler set variable (success)":            testVariableHandlerSetVariableSuccess,
		"test variable handler set variable (error - validation)": testVariableHandlerSetVariableValidationError,
		"test variable handler set variable (error - store)":      testVariableHandlerSetVariableStoreError,

		"test variable handler get variable (success)":       testVariableHandlerGetVariableSuccess,
		"test variable handler get variable (error - store)": testVariableHandlerGetVariableStoreError,

		"test variable handler delete variable (success)":       testVariableHandlerDeleteVariableSuccess,
		"test variable handler delete variable (error - store)": testVariableHandlerDeleteVariableStoreError,

		"test variable handler get all for project and environment (success)":       testVariableHandlerGetAllSuccess,
		"test variable handler get all for project and environment (error - store)": testVariableHandlerGetAllStoreError,
	}

	for scenario, fn := range scenarios {
		t.Run(scenario, func(t *testing.T) {

			handler := NewVariableCliHandler(mockStore, mockValidator)

			fn(t, handler)
		})
	}
}

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

func (s *MockVariableStore) Delete(projectName, environmentName, key string) error {
	args := s.Called(projectName, environmentName, key)

	return args.Error(0)
}

func (s *MockVariableStore) GetAll(projectName, environmentName string) ([]Variable, error) {
	args := s.Called(projectName, environmentName)

	return args.Get(0).([]Variable), args.Error(1)
}

type MockValidator struct {
	mock.Mock
}

func (v *MockValidator) Struct(s interface{}) error {
	args := v.Called(s)

	return args.Error(0)
}

func testVariableHandlerSetVariableSuccess(t *testing.T, handler VariableHandler) {
	secret := true

	variable := Variable{
		ProjectName:     "project_name",
		EnvironmentName: "environment_name",
		Key:             "VAR_KEY",
		Value:           "var_value",
		Secret:          &secret,
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
	secret := true

	variable := Variable{
		ProjectName:     "project_name",
		EnvironmentName: "environment_name",
		Key:             "VAR_KEY",
		Value:           "var_value",
		Secret:          &secret,
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
	secret := true

	variable := Variable{
		ProjectName:     "project_name",
		EnvironmentName: "environment_name",
		Key:             "VAR_KEY",
		Value:           "var_value",
		Secret:          &secret,
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
		t.Error("validator was not called as expected")
	}

	if expectations := mockStore.AssertExpectations(t); !expectations {
		t.Error("store was not called as expected")
	}

	mockValidatorStruct.Unset()
	mockStoreSet.Unset()
}

func testVariableHandlerGetVariableSuccess(t *testing.T, handler VariableHandler) {
	mockStoreGet := mockStore.
		On("Get", "project_name", "environment_name", "VAR_KEY").
		Return("var_value", nil)

	variable, err := handler.Get("project_name", "environment_name", "VAR_KEY")

	require.NoError(t, err, "should not return error")
	require.Equal(t, "var_value", variable, "should return variable value")

	if expectations := mockStore.AssertExpectations(t); !expectations {
		t.Error("store was not called as expected")
	}

	mockStoreGet.Unset()
}

func testVariableHandlerGetVariableStoreError(t *testing.T, handler VariableHandler) {
	mockStoreGet := mockStore.
		On("Get", "project_name", "environment_name", "VAR_KEY").
		Return("", errors.New("store_error"))

	variable, err := handler.Get("project_name", "environment_name", "VAR_KEY")

	require.Empty(t, variable, "should return empty variable")
	require.EqualError(t, err, "store_error", "should return store error")

	if expectations := mockStore.AssertExpectations(t); !expectations {
		t.Error("store was not called as expected")
	}

	mockStoreGet.Unset()
}

func testVariableHandlerDeleteVariableSuccess(t *testing.T, handler VariableHandler) {
	mockStoreDelete := mockStore.
		On("Delete", "project_name", "environment_name", "VAR_KEY").
		Return(nil)

	err := handler.Delete("project_name", "environment_name", "VAR_KEY")

	require.NoError(t, err, "should not return error")

	if expectations := mockStore.AssertExpectations(t); !expectations {
		t.Error("store was not called as expected")
	}

	mockStoreDelete.Unset()
}

func testVariableHandlerDeleteVariableStoreError(t *testing.T, handler VariableHandler) {
	mockStoreDelete := mockStore.
		On("Delete", "project_name", "environment_name", "VAR_KEY").
		Return(errors.New("store_error"))

	err := handler.Delete("project_name", "environment_name", "VAR_KEY")

	require.EqualError(t, err, "store_error", "should return store error")

	if expectations := mockStore.AssertExpectations(t); !expectations {
		t.Error("store was not called as expected")
	}

	mockStoreDelete.Unset()
}

func testVariableHandlerGetAllSuccess(t *testing.T, handler VariableHandler) {
	truePtr := true
	falsePtr := false

	mockStoreGetAll := mockStore.
		On("GetAll", "project_name", "environment_name").
		Return([]Variable{
			{
				Key:    "KEY_1",
				Value:  "value_1",
				Secret: &falsePtr,
			},
			{
				Key:    "KEY_2",
				Value:  "value_2",
				Secret: &truePtr,
			},
		}, nil)

	variables, err := handler.GetAll("project_name", "environment_name")

	require.NoError(t, err, "should not return error")
	require.Equal(t, []string{"KEY_1=value_1", "KEY_2=value_2"}, variables, "should return variables to inject in environment")

	if expectations := mockStore.AssertExpectations(t); !expectations {
		t.Error("store was not called as expected")
	}

	mockStoreGetAll.Unset()
}

func testVariableHandlerGetAllStoreError(t *testing.T, handler VariableHandler) {
	mockStoreGetAll := mockStore.
		On("GetAll", "project_name", "environment_name").
		Return([]Variable{}, errors.New("store_error"))

	variables, err := handler.GetAll("project_name", "environment_name")

	require.EqualError(t, err, "store_error", "should return store error")
	require.Empty(t, variables, "should not return any variables")

	if expectations := mockStore.AssertExpectations(t); !expectations {
		t.Error("store was not called as expected")
	}

	mockStoreGetAll.Unset()
}
