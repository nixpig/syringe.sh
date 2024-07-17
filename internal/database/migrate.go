package database

import (
	"database/sql"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	"github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/mattn/go-sqlite3"
)

func MigrateUp(db *sql.DB) error {
	instance, err := sqlite3.WithInstance(db, &sqlite3.Config{})
	if err != nil {
		return err
	}

	src, err := (&file.File{}).Open("migrations/")
	if err != nil {
		return err
	}

	m, err := migrate.NewWithInstance("file", src, "sqlite3", instance)
	if err != nil {
		return err
	}
	// m, err := migrate.New("file://migrations", "sqlite3:syringe.db")
	// if err != nil {
	// 	return err
	// }

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}

	return nil
}
