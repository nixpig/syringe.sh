package main

import (
	"net/http"
	"os"
	"slices"

	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
	"github.com/nixpig/syringe.sh/server/cmd"
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
	appDb, err := database.Connection(
		os.Getenv("DATABASE_URL"),
		os.Getenv("DATABASE_TOKEN"),
	)
	if err != nil {
		log.Error().Err(err).Msg("failed to connect to database")
		os.Exit(1)
	}

	defer appDb.Close()

	if slices.Index(os.Args, "--migrate") != -1 {
		log.Info().Msg("running database migration")
		if err := database.MigrateAppDb(appDb); err != nil {
			log.Error().Err(err).Msg("failed to run database migration")
			os.Exit(1)
		}
	}

	log.Info().Msg("building app components")
	validate := validator.New(validator.WithRequiredStructEnabled())
	appStore := stores.NewSqliteAppStore(appDb)
	appService := services.NewAppServiceImpl(appStore, validate, http.Client{}, services.TursoApiSettings{
		Url:   os.Getenv("API_BASE_URL"),
		Token: os.Getenv("API_TOKEN"),
	})

	sshHandlers := handlers.NewSshHandlers(appService, &log)
	sshServer := cmd.NewSyringeSshServer(sshHandlers, &log)

	if err := sshServer.Start(); err != nil {
		log.Error().Err(err).Msg("failed to start ssh server")
		os.Exit(1)
	}

	// httpHandlers := handlers.NewHttpHandlers(appService, &log)
	// httpServer := cmd.NewSyringeHttpServer(httpHandlers, &log)
	//
	// if err := httpServer.Start(); err != nil {
	// 	log.Error().Err(err).Msg("failed to start http server")
	// 	os.Exit(1)
	// }

	// registerScreen := screens.NewRegisterScreenModel(appService)
	//
	// p := tea.NewProgram(registerScreen, tea.WithAltScreen())
	// if _, err := p.Run(); err != nil {
	// 	fmt.Printf("failed to run program:\n%s", err)
	// 	os.Exit(1)
	// }
}
