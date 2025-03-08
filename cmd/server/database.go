package main

import (
	"database/sql"
	"embed"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	"github.com/golang-migrate/migrate/v4/source"
	_ "github.com/mattn/go-sqlite3"
)

//go:embed sql/*.sql
var Migrations embed.FS

func NewConnection(filename string) (*sql.DB, error) {
	connectionString := fmt.Sprintf("file:%s", filename)

	db, err := sql.Open("sqlite3", connectionString)
	if err != nil {
		return nil, fmt.Errorf("open database connection (%s): %w", connectionString, err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping database: %w", err)
	}

	return db, nil
}

type Migrator interface {
	Up() error
	Down() error
}

type Migration struct {
	migrate *migrate.Migrate
}

func (m *Migration) Up() error {
	return m.migrate.Up()
}

func (m *Migration) Down() error {
	return m.migrate.Down()
}

func NewMigration(db *sql.DB, migrations source.Driver) (*Migration, error) {
	instance, err := sqlite3.WithInstance(db, &sqlite3.Config{})
	if err != nil {
		return nil, fmt.Errorf("sqlite3 instance: %w", err)
	}

	m, err := migrate.NewWithInstance(
		"file",
		migrations,
		"sqlite3",
		instance,
	)
	if err != nil {
		return nil, fmt.Errorf("create migration: %w", err)
	}

	return &Migration{migrate: m}, nil
}
