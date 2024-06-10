package main

import (
	"net/http"
	"os"

	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
	"github.com/nixpig/syringe.sh/server/internal/database"
	"github.com/nixpig/syringe.sh/server/internal/server"
	"github.com/nixpig/syringe.sh/server/internal/services"
	"github.com/nixpig/syringe.sh/server/internal/stores"
	"github.com/rs/zerolog"
)

const (
	host = "localhost"
	port = "23234"
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
	appDb, err := database.Connection(
		os.Getenv("DATABASE_URL"),
		os.Getenv("DATABASE_TOKEN"),
	)
	if err != nil {
		log.Error().Err(err).Msg("failed to connect to database")
		os.Exit(1)
	}

	defer appDb.Close()

	log.Info().Msg("building app components")
	validate := validator.New(validator.WithRequiredStructEnabled())
	appStore := stores.NewSqliteAppStore(appDb)
	appService := services.NewAppServiceImpl(appStore, validate, http.Client{}, services.TursoApiSettings{
		Url:   os.Getenv("API_BASE_URL"),
		Token: os.Getenv("API_TOKEN"),
	})

	sshServer := server.NewSyringeSshServer(appService, &log)

	if err := sshServer.Start(host, port); err != nil {
		log.Error().Err(err).Msg("failed to start ssh server")
		os.Exit(1)
	}
}
