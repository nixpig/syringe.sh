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
	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"
	"github.com/nixpig/syringe.sh/database"
	"github.com/nixpig/syringe.sh/internal/server"
	"github.com/nixpig/syringe.sh/internal/stores"
)

const (
	env         = ".env"
	portEnv     = "SYRINGE_PORT"
	hostEnv     = "SYRINGE_HOST"
	keyEnv      = "SYRINGE_KEY"
	systemDBEnv = "SYRINGE_DB_SYSTEM_DIR"
	tenantDBEnv = "SYRINGE_DB_TENANT_DIR"
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
		log.Fatal("no port configured")
	}

	host := os.Getenv(hostEnv)
	if host == "" {
		log.Info("no host specified; defaulting to localhost")
		host = "localhost"
	}

	key := os.Getenv(keyEnv)
	if key == "" {
		log.Warn("no host key path configured", "keyEnv", keyEnv)
	}

	dbDir := os.Getenv(systemDBEnv)
	if dbDir == "" {
		log.Fatal("no system database dir specified")
	}

	if err := os.MkdirAll(dbDir, 0755); err != nil {
		log.Fatal(
			"failed to create system database directory",
			"dbDir", dbDir,
			"err", err,
		)
	}

	dbPath := filepath.Join(dbDir, "system.db")
	// TODO: make this a connection pool?
	db, err := database.NewConnection(dbPath)
	if err != nil {
		log.Fatal(
			"failed to create database connection",
			"dbPath", dbPath,
			"err", err,
		)
	}

	migrator, err := database.NewMigration(db, database.SystemMigrations)
	if err != nil {
		log.Fatal("new system migration", "err", err)
	}

	if err := migrator.Up(); err != nil {
		if !errors.Is(err, migrate.ErrNoChange) {
			log.Fatal("run system migration", "err", err)
		}
	}

	tenantDBDir := os.Getenv("SYRINGE_DB_TENANT_DIR")
	if err := os.MkdirAll(tenantDBDir, 0755); err != nil {
		log.Fatal(
			"create tenant database directory",
			"tenantDBDir", tenantDBDir,
			"err", err,
		)
	}

	systemStore := stores.NewSystemStore(db)

	middleware := []wish.Middleware{
		server.NewCmdMiddleware(systemStore),
		server.NewIdentityMiddleware(systemStore),
		server.ClientMiddleware,
		server.LoggingMiddleware,
	}

	s, err := server.New(host, port, key, middleware...)
	if err != nil {
		log.Fatal("failed to create server", "err", err)
	}

	done := make(chan os.Signal, 1)

	signal.Notify(done, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		if err := s.ListenAndServe(); err != nil && err != ssh.ErrServerClosed {
			log.Error("server stopped", "err", err)
		}

		done <- nil
	}()

	log.Info("server started", "host", host, "port", port)

	<-done

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
	defer cancel()

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
