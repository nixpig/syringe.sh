package cmd

import (
	"database/sql"
	"fmt"
	"os"
	"path"

	_ "github.com/mattn/go-sqlite3"
	"github.com/nixpig/syringe.sh/internal"
	"github.com/nixpig/syringe.sh/internal/database"
	"github.com/spf13/cobra"
)

// setCmd represents the set command
var setCmd = &cobra.Command{
	Use:   "set",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
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

		store := internal.NewVariableSqliteStore(db)

		query := `
		create table if not exists variables_ (
			id_ integer primary key autoincrement, 
			key_ text not null, 
			value_ text not null,
			secret_ boolean,
			project_name_ text,
			environment_name_ text
		)
	`

		_, err = db.Exec(query)
		if err != nil {
			fmt.Println("ERROR: ", err)
		}

		if err := store.Set(internal.Variable{
			Key:   args[0],
			Value: args[1],
		}); err != nil {
			fmt.Println(fmt.Errorf("unable to insert: %s", err))
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(setCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// setCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// setCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
