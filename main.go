package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
	"github.com/nixpig/syringe.sh/server/internal/database"
	"github.com/nixpig/syringe.sh/server/internal/handlers"
	"github.com/nixpig/syringe.sh/server/internal/services"
	"github.com/nixpig/syringe.sh/server/internal/stores"
)

func main() {
	if err := godotenv.Load(".env"); err != nil {
		fmt.Fprintf(os.Stderr, "failed to load '.env' file:\n%s", err)
		os.Exit(1)
	}

	db, err := database.Connection(
		os.Getenv("DATABASE_URL"),
		os.Getenv("DATABASE_TOKEN"),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to connect to database:\n%s", err)
		os.Exit(1)
	}

	defer db.Close()

	validate := validator.New(validator.WithRequiredStructEnabled())
	appStore := stores.NewSqliteAppStore(db)
	appService := services.NewAppServiceImpl(appStore, validate)

	httpHandlers := handlers.NewHttpHandlers(appService)

	mux := http.NewServeMux()

	mux.HandleFunc("POST /users", httpHandlers.RegisterUser)
	mux.HandleFunc("POST /keys", httpHandlers.AddPublicKey)

	server := &http.Server{
		Addr:         fmt.Sprintf(":%v", "3000"),
		Handler:      (mux),
		IdleTimeout:  time.Minute,
		ReadTimeout:  time.Second * 10,
		WriteTimeout: time.Second * 10,
	}

	if err := server.ListenAndServe(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to start server:\n%s", err)
		os.Exit(1)
	}

	// registerScreen := screens.NewRegisterScreenModel(appService)
	//
	// p := tea.NewProgram(registerScreen, tea.WithAltScreen())
	// if _, err := p.Run(); err != nil {
	// 	fmt.Printf("failed to run program:\n%s", err)
	// 	os.Exit(1)
	// }
}
