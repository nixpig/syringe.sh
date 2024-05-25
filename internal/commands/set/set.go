package set

import (
	"database/sql"
	"fmt"
	"os"
	"path"

	"github.com/spf13/cobra"
)

func Run(cmd *cobra.Command, args []string) {
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
		insert into variables_ (key_, value_) values (?, ?)
	`, args[0], args[1])
	if err != nil {
		fmt.Println(fmt.Errorf("unable to insert: %s", err))
		os.Exit(1)
	}
}
