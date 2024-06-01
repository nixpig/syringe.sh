package cmd

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/nixpig/syringe.sh/cli/internal/config"
	"github.com/nixpig/syringe.sh/cli/internal/database"
	"github.com/spf13/cobra"
)

var DB *sql.DB
var CONFIG *config.Config

var rootCmd = &cobra.Command{
	Use:   "syringe",
	Short: "A terminal-based utility to securely manage environment variables across projects and environments.",
	Long:  "A terminal-based utility to securely manage environment variables across projects and environments.",
}

func Execute() {
	userConfigDir, err := os.UserConfigDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to determine user config directory.\n%s", err)
	}

	cfg, err := config.GetConfig(userConfigDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to load config.\n%s", err)
	}

	CONFIG = &cfg

	DB, err = database.Connection(CONFIG.DatabaseFilePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to open database connection.\n%s", err)
	}

	defer DB.Close()

	err = rootCmd.Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to execute root command.\n%s", err)
	}
}

func init() {}
