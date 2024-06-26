package environment

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/nixpig/syringe.sh/pkg/ctxkeys"
	"github.com/nixpig/syringe.sh/pkg/validation"
	"github.com/spf13/cobra"
)

func EnvironmentCommand() *cobra.Command {
	environmentCmd := &cobra.Command{
		Use:               "environment",
		Aliases:           []string{"e"},
		Short:             "Manage environments",
		Long:              "Manage your environments.",
		PersistentPreRunE: initEnvironmentContext,
	}

	environmentCmd.AddCommand(environmentAddCommand())
	environmentCmd.AddCommand(environmentRemoveCommand())
	environmentCmd.AddCommand(environmentRenameCommand())
	environmentCmd.AddCommand(environmentListCommand())

	return environmentCmd
}

func environmentAddCommand() *cobra.Command {
	environmentAddCmd := &cobra.Command{
		Use:     "add [flags] ENVIRONMENT_NAME",
		Aliases: []string{"a"},
		Short:   "Add an environment",
		Example: "syringe environment add -p my_cool_project local",
		Args:    cobra.MatchAll(cobra.ExactArgs(1)),
		RunE:    environmentAddRunE,
	}

	environmentAddCmd.Flags().StringP("project", "p", "", "Project name")
	environmentAddCmd.MarkFlagRequired("project")

	environmentAddCmd.MarkFlagRequired("project")

	return environmentAddCmd
}

func environmentRemoveCommand() *cobra.Command {
	environmentRemoveCmd := &cobra.Command{
		Use:     "remove [flags] ENVIRONMENT_NAME",
		Aliases: []string{"r"},
		Short:   "Remove an environment",
		Example: "syringe environment remove -p my_cool_project staging",
		Args:    cobra.MatchAll(cobra.ExactArgs(1)),
		RunE:    environmentRemoveRunE,
	}

	environmentRemoveCmd.Flags().StringP("project", "p", "", "Project name")
	environmentRemoveCmd.MarkFlagRequired("project")

	return environmentRemoveCmd
}

func environmentAddRunE(cmd *cobra.Command, args []string) error {
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

func environmentRemoveRunE(cmd *cobra.Command, args []string) error {
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

func environmentRenameCommand() *cobra.Command {
	environmentRenameCmd := &cobra.Command{
		Use:     "rename [flags] CURRENT_ENVIRONMENT_NAME NEW_ENVIRONMENT_NAME",
		Aliases: []string{"u"},
		Short:   "Rename an environment",
		Example: "syringe environment rename -p my_cool_project staging prod",
		Args:    cobra.MatchAll(cobra.ExactArgs(2)),
		RunE:    environmentRenameE,
	}

	environmentRenameCmd.Flags().StringP("project", "p", "", "Project name")
	environmentRenameCmd.MarkFlagRequired("project")

	return environmentRenameCmd
}

func environmentRenameE(cmd *cobra.Command, args []string) error {
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

func environmentListCommand() *cobra.Command {
	environmentListCmd := &cobra.Command{
		Use:     "list [flags]",
		Aliases: []string{"l"},
		Short:   "List environments",
		Example: "syringe environment list -p my_cool_project",
		Args:    cobra.MatchAll(cobra.ExactArgs(0)),
		RunE:    environmentListE,
	}

	environmentListCmd.Flags().StringP("project", "p", "", "Project name")
	environmentListCmd.MarkFlagRequired("project")

	return environmentListCmd
}

func environmentListE(cmd *cobra.Command, args []string) error {
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

func initEnvironmentContext(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	db, ok := ctx.Value(ctxkeys.DB).(*sql.DB)
	if !ok {
		return fmt.Errorf("unable to get database from context")
	}

	environmentService := NewEnvironmentServiceImpl(
		NewSqliteEnvironmentStore(db),
		validation.NewValidator(),
	)

	ctx = context.WithValue(ctx, ctxkeys.EnvironmentService, environmentService)

	cmd.SetContext(ctx)

	return nil
}
