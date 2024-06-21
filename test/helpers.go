package test

import "fmt"

func IncorrectNumberOfArgsErrorMsg(accepts, received int) string {
	return fmt.Sprintf("Error: accepts %d arg(s), received %d\n", accepts, received)
}

func MaxLengthValidationErrorMsg(field string, length int) string {
	return fmt.Sprintf("Error: \"%s\" exceeds max length of %d characters\n", field, length)
}

func RequiredFlagsErrorMsg(flag string) string {
	return fmt.Sprintf("Error: required flag(s) \"%s\" not set\n", flag)
}

func ErrorMsg(e string) string {
	return fmt.Sprintf("Error: %s\n", e)
}

func ProjectAddedSuccessMsg(name string) string {
	return fmt.Sprintf("Project '%s' added\n", name)
}

func ProjectRemovedSuccessMsg(name string) string {
	return fmt.Sprintf("Project '%s' removed\n", name)
}

func ProjectRenamedSuccessMsg(name, newName string) string {
	return fmt.Sprintf("Project '%s' renamed to '%s'\n", name, newName)
}

func EnvironmentAddedSuccessMsg(environment, project string) string {
	return fmt.Sprintf("Environment '%s' added to project '%s'\n", environment, project)
}

func EnvironmentRemovedSuccessMsg(environment, project string) string {
	return fmt.Sprintf("Environment '%s' removed from project '%s'\n", environment, project)
}

func EnvironmentRenamedSuccessMsg(name, newName, project string) string {
	return fmt.Sprintf("Environment '%s' renamed to '%s' in project '%s'\n", name, newName, project)
}
