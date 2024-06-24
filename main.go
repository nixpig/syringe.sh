package main

import (
	"os"

	"github.com/charmbracelet/wish"
	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
	"github.com/nixpig/syringe.sh/server/internal/auth"
	"github.com/nixpig/syringe.sh/server/internal/database"
	"github.com/nixpig/syringe.sh/server/internal/server"
	"github.com/rs/zerolog"
)

func main() {
	log := zerolog.
		New(os.Stdout).
		Output(zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: "2006-01-02T15:04:05.999Z07:00",
		}).With().Timestamp().Logger()

	log.Info().Msg("loading environment")
	if err := godotenv.Load(".env"); err != nil {
		log.Error().Err(err).Msg("failed to load '.env' file")
		os.Exit(1)
	}

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

	log.Info().Msg("building app components")
	validate := validator.New(validator.WithRequiredStructEnabled())
	authStore := auth.NewSqliteAuthStore(appDB)
	authService := auth.NewAuthService(authStore, validate, &log)

	commandHandler := server.NewCommandHandler(&log, appDB)
	authHandler := server.NewAuthHandler(&log, authService)
	loggingHandler := server.NewLoggingHandler(&log)

	sshServer := server.NewServer(
		&log,
		[]wish.Middleware{
			commandHandler,
			authHandler,
			loggingHandler,
		},
	)

	if err := sshServer.Start(
		os.Getenv("APP_HOST"),
		os.Getenv("APP_PORT"),
	); err != nil {
		log.Error().Err(err).Msg("failed to start ssh server")
		os.Exit(1)
	}
}
