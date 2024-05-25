package init

import (
	"database/sql"
	"fmt"
	"os"
	"path"

	_ "github.com/mattn/go-sqlite3"
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
		create table if not exists variables_ (
			id_ integer primary key not null,
			key_ text not null,
			value_ text not null
		)
	`)
	if err != nil {
		fmt.Println(fmt.Errorf("could not create variables table: %s", err))
		os.Exit(1)
	}
}
