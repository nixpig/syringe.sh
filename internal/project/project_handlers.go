package project

import (
	"fmt"
	"strings"

	"github.com/nixpig/syringe.sh/pkg/ctxkeys"
	"github.com/spf13/cobra"
)

func ListCmdHandler(cmd *cobra.Command, args []string) error {
	projectService, ok := cmd.Context().Value(ctxkeys.ProjectService).(ProjectService)
	if !ok {
		return fmt.Errorf("unable to get project service from context")
	}

	projects, err := projectService.List()
	if err != nil {
		return err
	}

	projectNames := make([]string, len(projects.Projects))
	for i, p := range projects.Projects {
		projectNames[i] = p.Name
	}

	cmd.Print(strings.Join(projectNames, "\n"))

	return nil
}

func AddCmdHandler(cmd *cobra.Command, args []string) error {
	projectName := args[0]

	projectService, ok := cmd.Context().Value(ctxkeys.ProjectService).(ProjectService)
	if !ok {
		return fmt.Errorf("unable to get project service from context")
	}

	if err := projectService.Add(AddProjectRequest{
		Name: projectName,
	}); err != nil {
		return err
	}

	cmd.Println(fmt.Sprintf("Project '%s' added", projectName))

	return nil
}

func RemoveCmdHandler(cmd *cobra.Command, args []string) error {
	projectName := args[0]

	projectService, ok := cmd.Context().Value(ctxkeys.ProjectService).(ProjectService)
	if !ok {
		return fmt.Errorf("unable to get project service from context")
	}

	if err := projectService.Remove(RemoveProjectRequest{
		Name: projectName,
	}); err != nil {
		return err
	}

	cmd.Println(fmt.Sprintf("Project '%s' removed", projectName))

	return nil
}

func RenameCmdHandler(cmd *cobra.Command, args []string) error {
	name := args[0]
	newName := args[1]

	projectService, ok := cmd.Context().Value(ctxkeys.ProjectService).(ProjectService)
	if !ok {
		return fmt.Errorf("unable to get project service from context")
	}

	if err := projectService.Rename(RenameProjectRequest{
		Name:    name,
		NewName: newName,
	}); err != nil {
		return err
	}

	cmd.Println(fmt.Sprintf("Project '%s' renamed to '%s'", name, newName))

	return nil
}
