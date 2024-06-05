package main

import (
	"database/sql"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/joho/godotenv"
	"github.com/nixpig/syringe.sh/server/internal/database"
	"github.com/nixpig/syringe.sh/server/internal/screens"
)

var DB *sql.DB

func main() {
	if err := godotenv.Load(".env"); err != nil {
		fmt.Fprintf(os.Stdout, "unable to load env:\n%s", err)
	}

	databaseUrl := os.Getenv("TURSO_DATABASE_URL")
	databaseToken := os.Getenv("TURSO_AUTH_TOKEN")

	fmt.Println("databaseUrl: ", databaseUrl)
	fmt.Println("databaseToken: ", databaseToken)

	databaseConnectionString := databaseUrl + "?authToken=" + databaseToken

	DB, err := database.Connection(databaseConnectionString)

	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to connect to database: %s\n%s", databaseConnectionString, err)
		os.Exit(1)
	}

	// if err := database.MigrateUserDb(DB); err != nil {
	// 	fmt.Fprintf(os.Stderr, "failed to create tables: \n%s", err)
	// }

	defer DB.Close()

	fmt.Println(DB.Stats())

	registerScreen := screens.RegisterScreen{}

	p := tea.NewProgram(registerScreen.InitialModel(DB))
	if _, err := p.Run(); err != nil {
		fmt.Printf("error:\n%s", err)
		os.Exit(1)
	}
}

// func main() {
//
// 	mux := http.NewServeMux()
//
// 	userStore := user.NewSqliteUserStore(DB)
// 	userService := user.NewJsonUserService(userStore, validator.New(validator.WithRequiredStructEnabled()))
// 	httpHandlers := handlers.NewHttpHandlers(userService)
//
// 	mux.HandleFunc("POST /users/create", httpHandlers.CreateUser)
//
// 	fmt.Println("starting server...")
// 	server := &http.Server{
// 		Handler: mux,
// 		Addr:    ":3000",
// 	}
// 	if err := server.ListenAndServe(); err != nil {
// 		fmt.Println("failed to start server", err)
// 	}
// }
