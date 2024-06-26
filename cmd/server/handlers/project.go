package handlers

import (
	"fmt"
	"strings"

	"github.com/nixpig/syringe.sh/internal/project"
	"github.com/nixpig/syringe.sh/pkg/ctxkeys"
	"github.com/spf13/cobra"
)

func ProjectListHandler(cmd *cobra.Command, args []string) error {
	projectService, ok := cmd.Context().Value(ctxkeys.ProjectService).(project.ProjectService)
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

func ProjectAddHandler(cmd *cobra.Command, args []string) error {
	projectName := args[0]

	projectService, ok := cmd.Context().Value(ctxkeys.ProjectService).(project.ProjectService)
	if !ok {
		return fmt.Errorf("unable to get project service from context")
	}

	if err := projectService.Add(project.AddProjectRequest{
		Name: projectName,
	}); err != nil {
		return err
	}

	cmd.Println(fmt.Sprintf("Project '%s' added", projectName))

	return nil
}

func ProjectRemoveHandler(cmd *cobra.Command, args []string) error {
	projectName := args[0]

	projectService, ok := cmd.Context().Value(ctxkeys.ProjectService).(project.ProjectService)
	if !ok {
		return fmt.Errorf("unable to get project service from context")
	}

	if err := projectService.Remove(project.RemoveProjectRequest{
		Name: projectName,
	}); err != nil {
		return err
	}

	cmd.Println(fmt.Sprintf("Project '%s' removed", projectName))

	return nil
}

func ProjectRenameHandler(cmd *cobra.Command, args []string) error {
	name := args[0]
	newName := args[1]

	projectService, ok := cmd.Context().Value(ctxkeys.ProjectService).(project.ProjectService)
	if !ok {
		return fmt.Errorf("unable to get project service from context")
	}

	if err := projectService.Rename(project.RenameProjectRequest{
		Name:    name,
		NewName: newName,
	}); err != nil {
		return err
	}

	cmd.Println(fmt.Sprintf("Project '%s' renamed to '%s'", name, newName))

	return nil
}
