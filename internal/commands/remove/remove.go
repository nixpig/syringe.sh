package remove

import (
	"database/sql"
	"fmt"
	"os"
	"path"

	"github.com/spf13/cobra"
)

func Run(cmd *cobra.Command, args []string) {
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

	_, err = db.Exec(`
		delete from variables_ where key_ = ?
	`, args[0])
	if err != nil {
		fmt.Println(fmt.Errorf("unable to remove: %s", err))
		os.Exit(1)
	}
}
