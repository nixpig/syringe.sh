package cmd

import (
	"fmt"
	"os"
	"strings"

	_ "github.com/mattn/go-sqlite3"
	"github.com/nixpig/syringe.sh/internal/database"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:     "init",
	Short:   "Initialise syringe",
	Long:    `Initialise syringe the first time it's used on a particular system.`,
	Example: `  syringe init`,
	Run: func(cmd *cobra.Command, args []string) {
		if _, err := os.Open(CONFIG.ConfigFilePath); err == nil {
			fmt.Fprintf(os.Stdout, "Config file already exists. Overwriting this will break any project and environment links you currently have.\nDo you want to overwrite it? (y/N) ")

			var confirmation string

			fmt.Scan(&confirmation)

			confirmation = strings.ToUpper(strings.TrimSpace(confirmation))

			if confirmation != "Y" && confirmation != "YES" {
				fmt.Fprint(os.Stdout, "Exiting.")
				os.Exit(0)
			}
		}

		// 1. Generate UUID and save to config file - use a hash of UUID and path to generate link id??
		if err := os.WriteFile(
			CONFIG.ConfigFilePath,
			[]byte(""),
			os.ModePerm,
		); err != nil {
			fmt.Fprintf(os.Stderr, "unable to write config file: %s", err)
			os.Exit(1)
		}

		// 2. Create database and tables
		if err := database.CreateTables(DB); err != nil {
			fmt.Fprintf(os.Stderr, "unable to create database tables: %s", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
