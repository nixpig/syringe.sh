package cmd

import (
	"context"
	"crypto/sha1"
	"fmt"
	"net/http"
	"os"

	"github.com/charmbracelet/ssh"
	"github.com/go-playground/validator/v10"
	"github.com/nixpig/syringe.sh/server/internal/database"
	"github.com/nixpig/syringe.sh/server/internal/services"
	"github.com/nixpig/syringe.sh/server/internal/stores"
	"github.com/nixpig/syringe.sh/server/pkg/turso"
	"github.com/spf13/cobra"
	gossh "golang.org/x/crypto/ssh"
)

func projectCommand(sess ssh.Session) *cobra.Command {
	projectCmd := &cobra.Command{
		Use:                "project",
		Aliases:            []string{"p"},
		Short:              "Project",
		Long:               "Project",
		Example:            "syringe project",
		PersistentPreRunE:  initProjectContext(sess),
		PersistentPostRunE: closeProjectContext(sess),
	}

	projectCmd.AddCommand(projectAddCommand())

	return projectCmd
}

func projectAddCommand() *cobra.Command {
	projectAddCmd := &cobra.Command{
		Use:     "add",
		Aliases: []string{"a"},
		Short:   "add",
		Long:    "add",
		Example: "syringe project add []",
		Args:    cobra.MatchAll(cobra.ExactArgs(1)),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			projectService := cmd.Context().Value("PROJECT_SERVICE").(services.ProjectService)

			if err := projectService.AddProject(services.AddProjectRequest{
				Name: name,
			}); err != nil {
				return err
			}

			return nil
		},
	}

	return projectAddCmd
}

func initProjectContext(sess ssh.Session) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {

		ctx := cmd.Context()
		// add ssh sess to cobra cmd context
		// REVIEW: is this even used??
		ctx = context.WithValue(ctx, "sess", sess)

		// add user database to cobra cmd context
		api := turso.New(
			os.Getenv("DATABASE_ORG"),
			os.Getenv("API_TOKEN"),
			http.Client{},
		)

		marshalledKey := gossh.MarshalAuthorizedKey(sess.PublicKey())

		hashedKey := fmt.Sprintf("%x", sha1.Sum(marshalledKey))

		token, err := api.CreateToken(hashedKey, "30s")
		if err != nil {
			fmt.Println("failed to create token:", err)
		}

		db, err := database.Connection(
			"libsql://"+hashedKey+"-"+os.Getenv("DATABASE_ORG")+".turso.io",
			string(token.Jwt),
		)
		if err != nil {
			fmt.Println("error creating database connection:\n", err)
			return nil
		}

		projectStore := stores.NewSqliteProjectStore(db)
		projectService := services.NewProjectServiceImpl(projectStore, validator.New(validator.WithRequiredStructEnabled()))

		ctx = context.WithValue(ctx, "PROJECT_SERVICE", projectService)

		cmd.SetContext(ctx)

		return nil
	}
}

func closeProjectContext(sess ssh.Session) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {

		return nil
	}
}
