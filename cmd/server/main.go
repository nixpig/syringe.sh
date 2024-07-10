package main

import (
	"os"
	"time"

	"github.com/charmbracelet/wish"
	"github.com/joho/godotenv"
	"github.com/nixpig/syringe.sh/internal/auth"
	"github.com/nixpig/syringe.sh/internal/database"
	"github.com/nixpig/syringe.sh/internal/middleware"
	"github.com/nixpig/syringe.sh/pkg/validation"
	"github.com/rs/zerolog"
)

func main() {
	// -- LOGGING
	log := zerolog.
		New(os.Stdout).
		Output(zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: "2006-01-02T15:04:05.999Z07:00",
		}).With().Timestamp().Logger()

	// -- ENV
	log.Info().Msg("loading environment")
	if err := godotenv.Load(".env"); err != nil {
		log.Error().Err(err).Msg("failed to load '.env' file")
		os.Exit(1)
	}

	// -- DATABASE
	log.Info().Msg("connecting to database")
	appDB, err := database.Connection(
		os.Getenv("DATABASE_URL"),
		os.Getenv("DATABASE_TOKEN"),
	)
	if err != nil {
		log.Error().Err(err).Msg("failed to connect to database")
		os.Exit(1)
	}

	defer appDB.Close()

	// -- DEPENDENCY CONSTRUCTION
	log.Info().Msg("building app components")
	validate := validation.New()
	authStore := auth.NewSqliteAuthStore(appDB)
	authService := auth.NewAuthService(authStore, validate)

	// -- CMD

	// -- SERVER
	sshServer := newServer(
		&log,
		[]wish.Middleware{
			middleware.NewMiddlewareCommand(&log, appDB, validate),
			middleware.NewMiddlewareAuth(&log, authService),
			middleware.NewMiddlewareLogging(&log),
		},
		time.Duration(time.Second*30),
		os.Getenv("HOST_KEY_PATH"),
	)

	if err := sshServer.Start(
		os.Getenv("APP_HOST"),
		os.Getenv("APP_PORT"),
	); err != nil {
		log.Error().Err(err).Msg("failed to start ssh server")
		os.Exit(1)
	}
}
