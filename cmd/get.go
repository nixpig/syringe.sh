package cmd

import (
	"database/sql"
	"fmt"
	"os"
	"path"

	_ "github.com/mattn/go-sqlite3"
	"github.com/spf13/cobra"
)

// getCmd represents the get command
var getCmd = &cobra.Command{
	Use:   "get",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		userConfigDir, err := os.UserConfigDir()
		if err != nil {
			fmt.Println(fmt.Errorf("could not find user config directory: %s", err))
			os.Exit(1)
		}

		syringeConfigDir := path.Join(userConfigDir, "syringe")

		syringeDatabaseFile := path.Join(syringeConfigDir, "database.db")

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

		var rows *sql.Rows
		var params = []string{}

		if len(args) == 0 {
			rows, err = db.Query(`select value_ from variables_`)
			if err != nil {
				return
			}
		} else {
			params = append(params, args[0])
			rows, err = db.Query(`select value_ from variables_ where key_ = ?`, params[0])
			if err != nil {
				return
			}
		}

		var variables []string

		for rows.Next() {
			var variable string

			if err := rows.Scan(&variable); err != nil {
				fmt.Println(fmt.Errorf("unable to scan variable: %s", err))
				os.Exit(1)
			}

			variables = append(variables, variable)

		}

		fmt.Println(variables)
	},
}

func init() {
	rootCmd.AddCommand(getCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// getCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// getCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
