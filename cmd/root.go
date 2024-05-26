package cmd

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/nixpig/syringe.sh/internal/config"
	"github.com/nixpig/syringe.sh/internal/database"
	"github.com/spf13/cobra"
)

var DB *sql.DB

var rootCmd = &cobra.Command{
	Use:   "syringe",
	Short: "A terminal-based utility to securely manage environment variables across projects and environments.",
	Long:  "A terminal-based utility to securely manage environment variables across projects and environments.",
}

func Execute() {
	cfg, err := config.GetConfig()
	if err != nil {
		fmt.Println(fmt.Errorf("unable to load config: %s", err))
	}

	DB, err = database.Connection(cfg.DatabaseFilePath)
	if err != nil {
		fmt.Println(fmt.Errorf("unable to open database connection: %s", err))
		os.Exit(1)
	}

	defer DB.Close()

	err = rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.syringe.sh.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
