package main

import (
	"fmt"
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
	"github.com/rs/zerolog"
)

func main() {
	log := zerolog.
		New(os.Stdout).
		Output(zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: zerolog.TimeFormatUnix,
		})

	log.Info().Msg("loading environment")
	if err := godotenv.Load(".env"); err != nil {
		log.Error().Err(err).Msg("failed to load '.env' file:")
		os.Exit(1)
	}

	log.Info().Msg("connecting to database")
	db, err := database.Connection(
		os.Getenv("DATABASE_URL"),
		os.Getenv("DATABASE_TOKEN"),
	)
	if err != nil {
		log.Error().Err(err).Msg("failed to connect to database")
		os.Exit(1)
	}

	defer db.Close()

	if slices.Index(os.Args, "--migrate") != -1 {
		log.Info().Msg("running database migration")
		if err := database.MigrateAppDb(db); err != nil {
			log.Error().Err(err).Msg("failed to run database migration")
			os.Exit(1)
		}
	}

	log.Info().Msg("building app components")
	validate := validator.New(validator.WithRequiredStructEnabled())
	appStore := stores.NewSqliteAppStore(db)
	appService := services.NewAppServiceImpl(appStore, validate, http.Client{}, services.TursoApiSettings{
		Url:   os.Getenv("API_BASE_URL"),
		Token: os.Getenv("API_TOKEN"),
	})

	httpHandlers := handlers.NewHttpHandlers(appService, log)

	mux := http.NewServeMux()

	mux.HandleFunc("/users", httpHandlers.RegisterUser)
	mux.HandleFunc("/keys", httpHandlers.AddPublicKey)

	server := &http.Server{
		Addr:         fmt.Sprintf(":%v", "3000"),
		Handler:      (mux),
		IdleTimeout:  time.Minute,
		ReadTimeout:  time.Second * 10,
		WriteTimeout: time.Second * 10,
	}

	log.Info().Msg("starting http server")
	if err := server.ListenAndServe(); err != nil {
		log.Error().Err(err).Msg("failed to start server")
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
