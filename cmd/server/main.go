package main

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/charmbracelet/log"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"
	"github.com/nixpig/syringe.sh/database"
	"github.com/nixpig/syringe.sh/internal/server"
)

const (
	env     = ".env"
	portEnv = "SYRINGE_PORT"
	hostEnv = "SYRINGE_HOST"
	keyEnv  = "SYRINGE_KEY"
)

func main() {
	log.SetLevel(log.DebugLevel)
	// TODO: use viper instead of (or in addition to) env file?
	log.Info("loading environment", "env", env)
	if err := godotenv.Load(env); err != nil {
		log.Warn("failed to load environment file", "env", env, "err", err)
	}

	port := os.Getenv(portEnv)
	if port == "" {
		log.Fatal("no port configured", "portEnv", portEnv)
	}

	host := os.Getenv(hostEnv)
	if host == "" {
		host = "localhost"
	}

	key := os.Getenv(keyEnv)
	if key == "" {
		log.Warn("no host key path configured", "keyEnv", keyEnv)
	}

	// setup system database

	homeDir, _ := os.UserHomeDir()
	dbDir := filepath.Join(homeDir, ".syringe")
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		log.Fatal(
			"failed to create system directory",
			"dbDir", dbDir,
			"err", err,
		)
	}

	dbPath := filepath.Join(dbDir, "system.db")
	// TODO: make this a connection pool
	db, err := database.NewConnection(dbPath)
	if err != nil {
		log.Fatal(
			"failed to create database connection",
			"dbPath", dbPath,
			"err", err,
		)
	}

	driver, err := iofs.New(database.SystemMigrations, "sql")
	if err != nil {
		log.Fatal(
			"failed to create new system database driver",
			"err", err,
		)
	}

	migrator, err := database.NewMigration(db, driver)
	if err != nil {
		log.Fatal(
			"new system migration",
			"err", err,
		)
	}

	if err := migrator.Up(); err != nil {
		if !errors.Is(err, migrate.ErrNoChange) {
			log.Fatal(
				"run system migration",
				"err", err,
			)
		}
	}

	middleware := []wish.Middleware{
		server.CmdMiddleware,
		server.IdentityMiddleware,
		server.LoggingMiddleware,
	}

	s, err := server.New(host, port, key, middleware...)
	if err != nil {
		log.Fatal("failed to create server", "err", err)
	}

	done := make(chan os.Signal, 1)

	signal.Notify(done, os.Interrupt, syscall.SIGKILL, syscall.SIGINT)

	log.Info("starting server", "host", host, "port", port)
	go func() {
		if err := s.ListenAndServe(); err != nil && err != ssh.ErrServerClosed {
			log.Error("server stopped", "err", err)
		}

		done <- nil
	}()

	log.Info("server started")

	<-done

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
	defer cancel()

	log.Info("stopping server")
	if err := s.Shutdown(ctx); err != nil && err != ssh.ErrServerClosed {
		log.Fatal("server failed to stop gracefully", "err", err)
	}

	log.Info("server stopped")
}

func rateLimitingMiddleware(next ssh.Handler) ssh.Handler {
	return func(sess ssh.Session) {
		// TODO: rate limiting
		next(sess)
	}
}
