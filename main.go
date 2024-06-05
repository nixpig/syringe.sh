package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
	"github.com/nixpig/syringe.sh/server/internal/database"
	"github.com/nixpig/syringe.sh/server/internal/screens"
	"github.com/nixpig/syringe.sh/server/internal/services"
	"github.com/nixpig/syringe.sh/server/internal/stores"
)

func main() {
	if err := godotenv.Load(".env"); err != nil {
		fmt.Fprintf(os.Stdout, "unable to load '.env' file:\n%s", err)
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

	registerScreen := screens.NewRegisterScreenModel(appService)

	p := tea.NewProgram(registerScreen, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("error:\n%s", err)
		os.Exit(1)
	}
}
