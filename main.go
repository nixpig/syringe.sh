package main

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/nixpig/syringe.sh/server/internal/database"
)

var DB *sql.DB

func main() {
	databaseUrl := os.Getenv("TURSO_DATABASE_URL")
	databaseToken := os.Getenv("TURSO_AUTH_TOKEN")

	databaseConnectionString := databaseUrl + "?authToken=" + databaseToken

	DB, err := database.Connection(databaseConnectionString)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to connect to database: %s\n%s", databaseConnectionString, err)
		os.Exit(1)
	}

	defer DB.Close()

	fmt.Println(DB.Stats())
}
