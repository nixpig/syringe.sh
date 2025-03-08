package main

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"slices"
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

var allowedKeyTypes = []string{"ssh-rsa"}

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

	log.Info("starting server", "host", host, "port", port)

	middleware := []wish.Middleware{
		storeMiddleware,
		authMiddleware,
		loggingMiddleware,
		rateLimitingMiddleware,
	}

	server, err := wish.NewServer(
		wish.WithAddress(net.JoinHostPort(host, port)),
		wish.WithHostKeyPath(key),
		wish.WithMaxTimeout(time.Second*30),
		wish.WithPublicKeyAuth(func(ctx ssh.Context, key ssh.PublicKey) bool {
			log.Debug("check public key auth", "allowed", allowedKeyTypes, "actual", key.Type())
			return slices.Contains(allowedKeyTypes, key.Type())
		}),
		wish.WithMiddleware(middleware...),
	)
	if err != nil && err != ssh.ErrServerClosed {
		log.Fatal("server shutting down not gracefully", "err", err)
	}

	done := make(chan os.Signal, 1)

	signal.Notify(done, os.Interrupt, syscall.SIGKILL, syscall.SIGINT)

	go func() {
		if err := server.ListenAndServe(); err != nil && err != ssh.ErrServerClosed {
			log.Error("server stopped", "err", err)
		}

		done <- nil
	}()

	log.Info("server started")

	<-done

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	log.Info("shutting down server")
	if err := server.Shutdown(ctx); err != nil && err != ssh.ErrServerClosed {
		log.Fatal("server failed to shutdown gracefully", "err", err)
	}

	log.Info("server stopped")
}

// TODO: prevent concurrent access for same user

func rateLimitingMiddleware(next ssh.Handler) ssh.Handler {
	return func(sess ssh.Session) {
		// rate limiting
		next(sess)
	}
}

func loggingMiddleware(next ssh.Handler) ssh.Handler {
	return func(sess ssh.Session) {
		log.Info(
			"connect",
			"session", sess.Context().SessionID(),
			"user", sess.Context().User(),
			"address", sess.Context().RemoteAddr().String(),
			"public", sess.PublicKey() != nil,
			"client", sess.Context().ClientVersion(),
		)

		next(sess)

		log.Info("disconnect", "session", sess.Context().SessionID())
	}
}

func authMiddleware(next ssh.Handler) ssh.Handler {
	return func(sess ssh.Session) {
		username := sess.Context().User()
		publicKey := sess.PublicKey()

		publicKeysURL := fmt.Sprintf("https://github.com/%s.keys", username)

		resp, err := http.Get(publicKeysURL)
		if err != nil || resp.StatusCode != http.StatusOK {
			log.Error("failed to get public keys", "publicKeysURL", publicKeysURL)
			sess.Stderr().Write([]byte(fmt.Sprintf("Error: failed to get public keys from %s", publicKeysURL)))
			return
		}
		defer resp.Body.Close()

		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			k := scanner.Text()
			authorisedKey, _, _, _, err := ssh.ParseAuthorizedKey([]byte(k))
			if err != nil {
				log.Warn("failed to parse key", "key", k, "err", err)
				sess.Stderr().Write([]byte(fmt.Sprintf("Error: failed to parse authorised key: %s", err)))
			}

			if ssh.KeysEqual(publicKey, authorisedKey) {
				next(sess)
				return
			}
		}

		if err := scanner.Err(); err != nil {
			log.Error("failed to read response body", "err", err)
			sess.Stderr().Write([]byte(fmt.Sprintf("Error: failed to read keys: %s", err)))
		}

		log.Error("no matching keys")
		sess.Stderr().Write([]byte("Error: no matching keys found"))
	}
}

func storeMiddleware(next ssh.Handler) ssh.Handler {
	return func(sess ssh.Session) {
		// parse command
		// interact with store
		next(sess)
	}
}
