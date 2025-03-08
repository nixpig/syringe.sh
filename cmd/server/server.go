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
	"github.com/spf13/cobra"
)

const (
	env     = ".env"
	portEnv = "SYRINGE_PORT"
	hostEnv = "SYRINGE_HOST"
	keyEnv  = "SYRINGE_KEY"
)

var allowedKeyTypes = []string{"ssh-rsa", "ssh-ed25519"}

type syringeServer struct {
	s   *ssh.Server
	log *log.Logger
}

func (s syringeServer) New(
	host string,
	hostKeyPath string,
	logger *log.Logger,
	m ...wish.Middleware,
) (*syringeServer, error) {
	server, err := wish.NewServer(
		wish.WithAddress(host),
		wish.WithHostKeyPath(hostKeyPath),
		wish.WithMaxTimeout(time.Second*30),
		wish.WithPublicKeyAuth(func(ctx ssh.Context, key ssh.PublicKey) bool {
			log.Debug("check public key auth", "allowed", allowedKeyTypes, "actual", key.Type())
			return slices.Contains(allowedKeyTypes, key.Type())
		}),
		wish.WithMiddleware(m...),
	)
	if err != nil && err != ssh.ErrServerClosed {
		return nil, fmt.Errorf("server stopped not gracefully: %w", err)
	}

	return &syringeServer{
		s:   server,
		log: logger,
	}, nil
}

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

	logger := &log.Logger{}

	middleware := []wish.Middleware{
		storeMiddleware,
		publicKeyMiddleware,
		NewLoggingMiddleware(logger),
	}

	server, err := syringeServer{}.New(
		net.JoinHostPort(host, port),
		key,
		logger,
		middleware...,
	)
	if err != nil {
		log.Fatal("failed to create server", "err", err)
	}

	done := make(chan os.Signal, 1)

	signal.Notify(done, os.Interrupt, syscall.SIGKILL, syscall.SIGINT)

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

	log.Info("shutting down server")
	if err := server.s.Shutdown(ctx); err != nil && err != ssh.ErrServerClosed {
		log.Fatal("server failed to shutdown gracefully", "err", err)
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

// TODO: review whether this is even needed, given new solution design
func publicKeyMiddleware(next ssh.Handler) ssh.Handler {
	return func(sess ssh.Session) {
		username := sess.Context().User()
		publicKey := sess.PublicKey()

		publicKeysURL := fmt.Sprintf("https://github.com/%s.keys", username)

		resp, err := http.Get(publicKeysURL)
		if err != nil || resp.StatusCode != http.StatusOK {
			log.Warn("failed to get public keys", "publicKeysURL", publicKeysURL)
			sess.Stderr().Write([]byte(fmt.Sprintf("Error: failed to get public keys from %s\n", publicKeysURL)))
			return
		}
		defer resp.Body.Close()

		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			k := scanner.Text()
			authorisedKey, _, _, _, err := ssh.ParseAuthorizedKey([]byte(k))
			if err != nil {
				log.Debug("failed to parse authorised key", "key", k, "err", err)
				continue
			}

			if ssh.KeysEqual(publicKey, authorisedKey) {
				next(sess)
				return
			}
		}

		if err := scanner.Err(); err != nil {
			log.Error("failed to read keys response body", "err", err)
			sess.Stderr().Write([]byte(fmt.Sprintf("Error: failed to read keys\n")))
		}

		sess.Stderr().Write([]byte("Error: no matching keys found\n"))
		sess.Exit(1)
		return
	}
}

func root() *cobra.Command {
	rootCmd := &cobra.Command{
		Use: "syringe [flags]",
		CompletionOptions: cobra.CompletionOptions{
			DisableDefaultCmd: true,
		},
		SilenceUsage: true,
		Run: func(c *cobra.Command, args []string) {
			fmt.Println("ARGS: ", args)
		},
	}

	return rootCmd
}

var setCmd = &cobra.Command{
	Use:  "set",
	Args: cobra.ExactArgs(2),
	RunE: func(c *cobra.Command, args []string) error {
		c.OutOrStdout().Write([]byte(fmt.Sprintf("set: %s=%s", args[0], args[1])))
		return nil
	},
}

var getCmd = &cobra.Command{
	Use:  "get",
	Args: cobra.ExactArgs(1),
	RunE: func(c *cobra.Command, args []string) error {
		c.OutOrStdout().Write([]byte(fmt.Sprintf("get: %s", args[0])))
		return nil
	},
}

var listCmd = &cobra.Command{
	Use:  "list",
	Args: cobra.NoArgs,
	RunE: func(c *cobra.Command, args []string) error {
		c.OutOrStdout().Write([]byte("list"))
		return nil
	},
}

var removeCmd = &cobra.Command{
	Use:  "remove",
	Args: cobra.ExactArgs(1),
	RunE: func(c *cobra.Command, args []string) error {
		c.OutOrStdout().Write([]byte(fmt.Sprintf("remove: %s", args[0])))
		return nil
	},
}

func storeMiddleware(next ssh.Handler) ssh.Handler {
	return func(sess ssh.Session) {
		rootCmd := root()

		rootCmd.AddCommand(setCmd, getCmd, listCmd, removeCmd)
		rootCmd.SetArgs(sess.Command())
		rootCmd.SetIn(sess)
		rootCmd.SetOut(sess)
		rootCmd.SetErr(sess.Stderr())

		if err := rootCmd.Execute(); err != nil {
			sess.Exit(1)
			return
		}

		next(sess)
	}
}
