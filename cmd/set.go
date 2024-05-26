package cmd

import (
	"database/sql"
	"fmt"
	"os"
	"path"

	"github.com/go-playground/validator/v10"
	_ "github.com/mattn/go-sqlite3"
	"github.com/nixpig/syringe.sh/internal/database"
	internal "github.com/nixpig/syringe.sh/internal/variables"
	"github.com/spf13/cobra"
)

var setCmd = &cobra.Command{
	Use:   "set",
	Short: "Set an environment variable.",
	Long: `Set an environment variable against the current project and environment.

Examples:
  syringe set DB_PASSWORD p4ssw0rd
  syringe set -p dunce -e dev DB_PASSWORD p4ssw0rd
	`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 2 {
			fmt.Println(fmt.Errorf("expected 2 arguments - key and value - but got %d", len(args)))
			os.Exit(1)
		}
		userConfigDir, err := os.UserConfigDir()
		if err != nil {
			fmt.Println(fmt.Errorf("could not find user config directory: %s", err))
			os.Exit(1)
		}

		syringeConfigDir := path.Join(userConfigDir, "syringe")

		syringeDatabaseFile := path.Join(syringeConfigDir, "database.db")

		database.Create(database.DbConfig{Location: syringeDatabaseFile})

		db, err := sql.Open("sqlite3", syringeDatabaseFile)
		if err != nil {
			fmt.Println(fmt.Errorf("could not open database file: %s", err))
			os.Exit(1)
		}

		defer db.Close()

		if err := db.Ping(); err != nil {
			fmt.Println(fmt.Errorf("could not ping database: %s", err))
			os.Exit(1)
		}

		store := internal.NewVariableStoreSqlite(db)

		handler := internal.NewVariableCliHandler(store, validator.New())

		projectName, err := cmd.Flags().GetString("project")
		if err != nil {
			fmt.Println("no project provided")
			os.Exit(1)
		}

		environmentName, err := cmd.Flags().GetString("environment")
		if err != nil {
			fmt.Println("no environment provided")
			os.Exit(1)
		}

		secret, err := cmd.Flags().GetBool("secret")
		if err != nil {
			fmt.Println("unable to get secret value")
			os.Exit(1)
		}

		fmt.Println("secret value: ", secret)

		variableKey := args[0]
		variableValue := args[1]

		fmt.Println("args: ",
			projectName,
			environmentName,
			variableKey,
			variableValue,
			secret,
		)

		err = handler.Set(
			projectName,
			environmentName,
			variableKey,
			variableValue,
			secret,
		)
		if err != nil {
			fmt.Println("error setting variable: ", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(setCmd)

	setCmd.Flags().StringP("project", "p", "", "Project")
	setCmd.Flags().StringP("environment", "e", "", "Environment")
	setCmd.Flags().BoolP("secret", "s", false, "Variable is secret")
}
