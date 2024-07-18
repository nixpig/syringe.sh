package database

import (
	"database/sql"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	"github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/mattn/go-sqlite3"
)

type Migrator interface {
	Up() error
	Down() error
}

type Migration struct {
	migrate *migrate.Migrate
}

func NewMigration(db *sql.DB, migrationsDir string) (*Migration, error) {
	instance, err := sqlite3.WithInstance(db, &sqlite3.Config{})
	if err != nil {
		return nil, err
	}

	src, err := (&file.File{}).Open(migrationsDir)
	if err != nil {
		return nil, err
	}

	m, err := migrate.NewWithInstance(
		"file",
		src,
		"sqlite3",
		instance,
	)
	if err != nil {
		return nil, err
	}

	return &Migration{migrate: m}, nil
}

func (m Migration) Up() error {
	return m.migrate.Up()
}

func (m Migration) Down() error {
	return m.migrate.Down()
}
