package environment

import (
	"fmt"
	"strings"

	"github.com/nixpig/syringe.sh/pkg"
	"github.com/spf13/cobra"
)

func NewHandlerEnvironmentAdd(environmentService EnvironmentService) pkg.CobraHandler {
	return func(cmd *cobra.Command, args []string) error {
		environmentName := args[0]

		project, _ := cmd.Flags().GetString("project")

		if err := environmentService.Add(AddEnvironmentRequest{
			Name:    environmentName,
			Project: project,
		}); err != nil {
			return err
		}

		cmd.Println(fmt.Sprintf("Environment '%s' added to project '%s'.", environmentName, project))

		return nil
	}
}

func NewHandlerEnvironmentRemove(environmentService EnvironmentService) pkg.CobraHandler {
	return func(cmd *cobra.Command, args []string) error {
		environmentName := args[0]

		project, _ := cmd.Flags().GetString("project")

		if err := environmentService.Remove(RemoveEnvironmentRequest{
			Name:    environmentName,
			Project: project,
		}); err != nil {
			return err
		}

		cmd.Println(fmt.Sprintf("Environment '%s' removed from project '%s'.", environmentName, project))

		return nil
	}
}

func NewHandlerEnvironmentRename(environmentService EnvironmentService) pkg.CobraHandler {
	return func(cmd *cobra.Command, args []string) error {
		name := args[0]
		newName := args[1]

		project, _ := cmd.Flags().GetString("project")

		if err := environmentService.Rename(RenameEnvironmentRequest{
			Name:    name,
			NewName: newName,
			Project: project,
		}); err != nil {
			return err
		}

		cmd.Println(fmt.Sprintf("Environment '%s' renamed to '%s' in project '%s'.", name, newName, project))

		return nil
	}
}

func NewHandlerEnvironmentList(environmentService EnvironmentService) pkg.CobraHandler {
	return func(cmd *cobra.Command, args []string) error {
		project, _ := cmd.Flags().GetString("project")

		environments, err := environmentService.List(ListEnvironmentRequest{
			Project: project,
		})
		if err != nil {
			return err
		}

		environmentNames := make([]string, len(environments.Environments))
		for i, e := range environments.Environments {
			environmentNames[i] = e.Name
		}

		cmd.Println(strings.Join(environmentNames, "\n"))

		return nil
	}
}
