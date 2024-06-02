package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"

	"github.com/go-playground/validator/v10"
	"github.com/nixpig/syringe.sh/server/internal/database"
	"github.com/nixpig/syringe.sh/server/internal/handlers"
	"github.com/nixpig/syringe.sh/server/internal/user"
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

	if err := database.CreateTables(DB); err != nil {
		fmt.Fprintf(os.Stderr, "failed to create tables: \n%s", err)
	}

	defer DB.Close()

	fmt.Println(DB.Stats())

	mux := http.NewServeMux()

	userStore := user.NewSqliteUserStore(DB)
	userService := user.NewJsonUserService(userStore, validator.New(validator.WithRequiredStructEnabled()))
	httpHandlers := handlers.NewHttpHandlers(userService)

	mux.HandleFunc("POST /users/create", httpHandlers.PostUsersCreate)

	fmt.Println("starting server...")
	server := &http.Server{
		Handler: mux,
		Addr:    ":3000",
	}
	if err := server.ListenAndServe(); err != nil {
		fmt.Println("failed to start server", err)
	}
}
