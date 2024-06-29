package environment

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/nixpig/syringe.sh/pkg"
	"github.com/nixpig/syringe.sh/pkg/ctxkeys"
	"github.com/nixpig/syringe.sh/pkg/validation"
	"github.com/spf13/cobra"
)

func NewCmdEnvironment(init pkg.CobraHandler) *cobra.Command {
	environmentCmd := &cobra.Command{
		Use:               "environment",
		Aliases:           []string{"e"},
		Short:             "Manage environments",
		Long:              "Manage your environments.",
		PersistentPreRunE: init,
	}

	return environmentCmd
}

func NewCmdEnvironmentAdd(handler pkg.CobraHandler) *cobra.Command {
	addCmd := &cobra.Command{
		Use:     "add [flags] ENVIRONMENT_NAME",
		Aliases: []string{"a"},
		Short:   "Add an environment",
		Example: "syringe environment add -p my_cool_project local",
		Args:    cobra.MatchAll(cobra.ExactArgs(1)),
		RunE:    handler,
	}

	addCmd.Flags().StringP("project", "p", "", "Project name")
	addCmd.MarkFlagRequired("project")

	addCmd.MarkFlagRequired("project")

	return addCmd
}

func NewCmdEnvironmentRemove(handler pkg.CobraHandler) *cobra.Command {
	removeCmd := &cobra.Command{
		Use:     "remove [flags] ENVIRONMENT_NAME",
		Aliases: []string{"r"},
		Short:   "Remove an environment",
		Example: "syringe environment remove -p my_cool_project staging",
		Args:    cobra.MatchAll(cobra.ExactArgs(1)),
		RunE:    handler,
	}

	removeCmd.Flags().StringP("project", "p", "", "Project name")
	removeCmd.MarkFlagRequired("project")

	return removeCmd
}

func NewCmdEnvironmentRename(handler pkg.CobraHandler) *cobra.Command {
	renameCmd := &cobra.Command{
		Use:     "rename [flags] CURRENT_ENVIRONMENT_NAME NEW_ENVIRONMENT_NAME",
		Aliases: []string{"u"},
		Short:   "Rename an environment",
		Example: "syringe environment rename -p my_cool_project staging prod",
		Args:    cobra.MatchAll(cobra.ExactArgs(2)),
		RunE:    handler,
	}

	renameCmd.Flags().StringP("project", "p", "", "Project name")
	renameCmd.MarkFlagRequired("project")

	return renameCmd
}

func NewCmdEnvironmentList(handler pkg.CobraHandler) *cobra.Command {
	listCmd := &cobra.Command{
		Use:     "list [flags]",
		Aliases: []string{"l"},
		Short:   "List environments",
		Example: "syringe environment list -p my_cool_project",
		Args:    cobra.MatchAll(cobra.ExactArgs(0)),
		RunE:    handler,
	}

	listCmd.Flags().StringP("project", "p", "", "Project name")
	listCmd.MarkFlagRequired("project")

	return listCmd
}

func InitContext(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	db, ok := ctx.Value(ctxkeys.USER_DB).(*sql.DB)
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
