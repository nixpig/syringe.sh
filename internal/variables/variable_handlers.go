package internal

import (
	"fmt"
	"strings"
)

type Validator interface {
	Struct(s interface{}) error
}

type VariableHandler interface {
	Set(
		projectName,
		environmentName,
		variableKey,
		variableValue string,
		secret *bool,
	) error

	Get(projectName, environmentName, variableKey string) (string, error)
	Delete(projectName, environmentName, variableKey string) error
	GetAll(projectName, environmentName string) ([]string, error)
}

type VariableCliHandler struct {
	store    VariableStore
	validate Validator
}

func NewVariableCliHandler(store VariableStore, validate Validator) VariableCliHandler {
	return VariableCliHandler{
		store:    store,
		validate: validate,
	}
}

func (v VariableCliHandler) Set(
	projectName,
	environmentName,
	variableKey,
	variableValue string,
	secret *bool,
) error {
	variable := Variable{
		ProjectName:     projectName,
		EnvironmentName: environmentName,
		Key:             variableKey,
		Value:           variableValue,
		Secret:          secret,
	}

	if err := v.validate.Struct(variable); err != nil {
		return err
	}

	if err := v.store.Set(variable); err != nil {
		return err
	}

	return nil
}

func (v VariableCliHandler) Get(
	projectName,
	environmentName,
	variableKey string,
) (string, error) {
	variable, err := v.store.Get(projectName, environmentName, variableKey)
	if err != nil {
		return "", err
	}

	return variable, nil
}

func (v VariableCliHandler) Delete(
	projectName,
	environmentName,
	variableKey string,
) error {
	err := v.store.Delete(projectName, environmentName, variableKey)
	if err != nil {
		return err
	}

	return nil
}

func (v VariableCliHandler) GetAll(projectName, environmentName string) ([]string, error) {
	fmt.Println("handling: ", projectName, environmentName)
	vars, err := v.store.GetAll(projectName, environmentName)
	if err != nil {
		return nil, err
	}

	variables := make([]string, len(vars))

	for i, variable := range vars {
		variables[i] = strings.Join([]string{variable.Key, variable.Value}, "=")
	}

	return variables, nil
}
