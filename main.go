package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"slices"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
	"github.com/nixpig/syringe.sh/server/internal/database"
	"github.com/nixpig/syringe.sh/server/internal/handlers"
	"github.com/nixpig/syringe.sh/server/internal/services"
	"github.com/nixpig/syringe.sh/server/internal/stores"
)

func main() {
	slog.Info("loading environment\n")
	if err := godotenv.Load(".env"); err != nil {
		slog.Error("failed to load '.env' file:\n%s", err)
		os.Exit(1)
	}

	slog.Info("connecting to database\n")
	db, err := database.Connection(
		os.Getenv("DATABASE_URL"),
		os.Getenv("DATABASE_TOKEN"),
	)
	if err != nil {
		slog.Error("failed to connect to database:\n%s", err)
		os.Exit(1)
	}

	defer db.Close()

	if slices.Index(os.Args, "--migrate") != -1 {
		slog.Info("running database migration\n")
		if err := database.MigrateAppDb(db); err != nil {
			slog.Error("failed to run database migration:\n%s", err)
			os.Exit(1)
		}
	}

	slog.Info("building app components\n")
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

	slog.Info("starting http server")
	if err := server.ListenAndServe(); err != nil {
		slog.Error("failed to start server:\n%s", err)
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
