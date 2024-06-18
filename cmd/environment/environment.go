package environment

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/nixpig/syringe.sh/server/internal/services"
	"github.com/nixpig/syringe.sh/server/internal/stores"
	"github.com/nixpig/syringe.sh/server/pkg"
	"github.com/spf13/cobra"
)

const (
	dbCtxKey          = pkg.DBCtxKey
	environmentCtxKey = pkg.EnvironmentCtxKey
)

func EnvironmentCommand() *cobra.Command {
	environmentCmd := &cobra.Command{
		Use:               "environment",
		Aliases:           []string{"e"},
		Short:             "Manage environments",
		PersistentPreRunE: initEnvironmentContext,
	}

	environmentCmd.AddCommand(environmentAddCommand())
	environmentCmd.AddCommand(environmentRemoveCommand())
	environmentCmd.AddCommand(environmentRenameCommand())

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

	environmentService, ok := cmd.Context().Value(environmentCtxKey).(services.EnvironmentService)
	if !ok {
		return fmt.Errorf("unable to get environment service")
	}

	if err := environmentService.Add(services.AddEnvironmentRequest{
		Name:        environmentName,
		ProjectName: project,
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

	environmentService, ok := cmd.Context().Value(environmentCtxKey).(services.EnvironmentService)
	if !ok {
		return fmt.Errorf("unable to get environment service")
	}

	if err := environmentService.Remove(services.RemoveEnvironmentRequest{
		Name:        environmentName,
		ProjectName: project,
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

	environmentService, ok := cmd.Context().Value(environmentCtxKey).(services.EnvironmentService)
	if !ok {
		return fmt.Errorf("unable to get environment service")
	}

	if err := environmentService.Rename(services.RenameEnvironmentRequest{
		Name:        name,
		NewName:     newName,
		ProjectName: project,
	}); err != nil {
		return err
	}

	cmd.Println(fmt.Sprintf("Environment '%s' renamed to '%s' in project '%s'", name, newName, project))

	return nil
}

func initEnvironmentContext(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	db, ok := ctx.Value(dbCtxKey).(*sql.DB)
	if !ok {
		return fmt.Errorf("unable to get database from context")
	}

	environmentService := services.NewEnvironmentServiceImpl(
		stores.NewSqliteEnvironmentStore(db),
		validator.New(validator.WithRequiredStructEnabled()))

	ctx = context.WithValue(ctx, environmentCtxKey, environmentService)

	cmd.SetContext(ctx)

	return nil
}
