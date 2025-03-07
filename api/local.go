package api

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/nixpig/syringe.sh/internal/database"
	"github.com/nixpig/syringe.sh/internal/items"
	"github.com/nixpig/syringe.sh/internal/migrations"
	"github.com/nixpig/syringe.sh/stores"
)

/*
	Queries local database directly
*/

type fileAPI struct {
	// calls s directly
	s  stores.Store
	db *sql.DB
}

func newFileAPI(path string) (*fileAPI, error) {
	var err error

	dbDir := filepath.Dir(path)
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		return nil, fmt.Errorf("create store directory: %w", err)
	}

	db, err := database.NewConnection(path)
	if err != nil {
		return nil, fmt.Errorf("new database connection: %w", err)
	}

	driver, err := iofs.New(migrations.Migrations, "sql")
	if err != nil {
		return nil, fmt.Errorf("new driver: %w", err)
	}

	migrator, err := database.NewMigration(db, driver)
	if err != nil {
		return nil, fmt.Errorf("create new migration: %w", err)
	}

	if err := migrator.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {

		} else {
			return nil, fmt.Errorf("run migration: %w", err)
		}
	}

	store := stores.NewSqliteStore(db)

	return &fileAPI{s: store, db: db}, nil
}

func (l *fileAPI) Set(item *items.Item) error {
	return l.s.Set(item)
}

func (l *fileAPI) Get(key string) (*items.Item, error) {
	return l.s.Get(key)
}

func (l *fileAPI) List() ([]items.Item, error) {
	return l.s.List()
}

func (l *fileAPI) Remove(key string) error {
	return l.s.Remove(key)
}

func (l *fileAPI) Close() error {
	return l.db.Close()
}
