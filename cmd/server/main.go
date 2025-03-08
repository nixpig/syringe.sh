package main

import (
	"context"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/charmbracelet/log"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	"github.com/joho/godotenv"
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

	middleware := []wish.Middleware{
		cmdMiddleware,
		identityMiddleware,
		loggingMiddleware,
	}

	server, err := syringeServer{}.New(
		net.JoinHostPort(host, port),
		key,
		middleware...,
	)
	if err != nil {
		log.Fatal("failed to create server", "err", err)
	}

	done := make(chan os.Signal, 1)

	signal.Notify(done, os.Interrupt, syscall.SIGKILL, syscall.SIGINT)

	log.Info("starting server", "host", host, "port", port)
	go func() {
		if err := server.s.ListenAndServe(); err != nil && err != ssh.ErrServerClosed {
			log.Error("server stopped", "err", err)
		}

		done <- nil
	}()

	log.Info("server started")

	<-done

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	log.Info("stopping server")
	if err := server.s.Shutdown(ctx); err != nil && err != ssh.ErrServerClosed {
		log.Fatal("server failed to stop gracefully", "err", err)
	}

	log.Info("server stopped")
}

// TODO: prevent concurrent access for same user

func rateLimitingMiddleware(next ssh.Handler) ssh.Handler {
	return func(sess ssh.Session) {
		// TODO: rate limiting
		next(sess)
	}
}
