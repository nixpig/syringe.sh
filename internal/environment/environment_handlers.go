package environment

import (
	"fmt"
	"strings"

	"github.com/nixpig/syringe.sh/pkg/ctxkeys"
	"github.com/spf13/cobra"
)

func AddCmdHandler(cmd *cobra.Command, args []string) error {
	environmentName := args[0]

	project, err := cmd.Flags().GetString("project")
	if err != nil {
		return err
	}

	environmentService, ok := cmd.Context().Value(ctxkeys.EnvironmentService).(EnvironmentService)
	if !ok {
		return fmt.Errorf("unable to get environment service")
	}

	if err := environmentService.Add(AddEnvironmentRequest{
		Name:    environmentName,
		Project: project,
	}); err != nil {
		return err
	}

	cmd.Println(fmt.Sprintf("Environment '%s' added to project '%s'", environmentName, project))

	return nil
}

func RemoveCmdHandler(cmd *cobra.Command, args []string) error {
	environmentName := args[0]

	project, err := cmd.Flags().GetString("project")
	if err != nil {
		return err
	}

	environmentService, ok := cmd.Context().Value(ctxkeys.EnvironmentService).(EnvironmentService)
	if !ok {
		return fmt.Errorf("unable to get environment service")
	}

	if err := environmentService.Remove(RemoveEnvironmentRequest{
		Name:    environmentName,
		Project: project,
	}); err != nil {
		return err
	}

	cmd.Println(fmt.Sprintf("Environment '%s' removed from project '%s'", environmentName, project))

	return nil
}

func RenameCmdHandler(cmd *cobra.Command, args []string) error {
	name := args[0]
	newName := args[1]

	project, err := cmd.Flags().GetString("project")
	if err != nil {
		return err
	}

	environmentService, ok := cmd.Context().Value(ctxkeys.EnvironmentService).(EnvironmentService)
	if !ok {
		return fmt.Errorf("unable to get environment service")
	}

	if err := environmentService.Rename(RenameEnvironmentRequest{
		Name:    name,
		NewName: newName,
		Project: project,
	}); err != nil {
		return err
	}

	cmd.Println(fmt.Sprintf("Environment '%s' renamed to '%s' in project '%s'", name, newName, project))

	return nil
}

func ListCmdHandler(cmd *cobra.Command, args []string) error {
	project, err := cmd.Flags().GetString("project")
	if err != nil {
		return err
	}

	environmentService, ok := cmd.Context().Value(ctxkeys.EnvironmentService).(EnvironmentService)
	if !ok {
		return fmt.Errorf("unable to get environment service")
	}

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

	cmd.Print(strings.Join(environmentNames, "\n"))

	return nil
}
