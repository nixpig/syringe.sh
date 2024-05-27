package cmd

import (
	"database/sql"
	"errors"
	"fmt"
	"os"

	"github.com/nixpig/syringe.sh/internal/config"
	"github.com/nixpig/syringe.sh/internal/database"
	"github.com/spf13/cobra"
)

var DB *sql.DB
var CONFIG *config.Config

var rootCmd = &cobra.Command{
	Use:   "syringe",
	Short: "A terminal-based utility to securely manage environment variables across projects and environments.",
	Long:  "A terminal-based utility to securely manage environment variables across projects and environments.",
}

func Execute() error {
	userConfigDir, err := os.UserConfigDir()
	if err != nil {
		return errors.New(fmt.Sprintf("unable to determine user config directory: %s", err))
	}

	cfg, err := config.GetConfig(userConfigDir)
	if err != nil {
		return errors.New(fmt.Sprintf("unable to load config: %s", err))
	}

	CONFIG = &cfg

	DB, err = database.Connection(CONFIG.DatabaseFilePath)
	if err != nil {
		return errors.New(fmt.Sprintf("unable to open database connection: %s", err))
	}

	defer DB.Close()

	err = rootCmd.Execute()
	if err != nil {
		return errors.New(fmt.Sprintf("unable to execute root command: %s", err))
	}

	return nil
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.syringe.sh.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	// rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
