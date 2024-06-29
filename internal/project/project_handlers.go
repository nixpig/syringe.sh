package project

import (
	"fmt"
	"strings"

	"github.com/nixpig/syringe.sh/pkg"
	"github.com/spf13/cobra"
)

func NewHandlerProjectList(projectService ProjectService) pkg.CobraHandler {
	return func(cmd *cobra.Command, args []string) error {
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
}

func NewHandlerProjectAdd(projectService ProjectService) pkg.CobraHandler {
	return func(cmd *cobra.Command, args []string) error {
		projectName := args[0]

		if err := projectService.Add(AddProjectRequest{
			Name: projectName,
		}); err != nil {
			return err
		}

		cmd.Println(fmt.Sprintf("Project '%s' added", projectName))

		return nil
	}
}

func NewHandlerProjectRemove(projectService ProjectService) pkg.CobraHandler {
	return func(cmd *cobra.Command, args []string) error {
		projectName := args[0]

		if err := projectService.Remove(RemoveProjectRequest{
			Name: projectName,
		}); err != nil {
			return err
		}

		cmd.Println(fmt.Sprintf("Project '%s' removed", projectName))

		return nil
	}
}

func NewHandlerProjectRename(projectService ProjectService) pkg.CobraHandler {
	return func(cmd *cobra.Command, args []string) error {
		name := args[0]
		newName := args[1]

		if err := projectService.Rename(RenameProjectRequest{
			Name:    name,
			NewName: newName,
		}); err != nil {
			return err
		}

		cmd.Println(fmt.Sprintf("Project '%s' renamed to '%s'", name, newName))

		return nil
	}
}
