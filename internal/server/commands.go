package server

import (
	"context"
	"crypto/sha1"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/log"
	"github.com/charmbracelet/ssh"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "github.com/mattn/go-sqlite3"
	"github.com/nixpig/syringe.sh/database"
	"github.com/nixpig/syringe.sh/internal/serrors"
	"github.com/nixpig/syringe.sh/internal/stores"
	"github.com/spf13/cobra"
)

const (
	commandTimeout = time.Second * 5
)

func CmdMiddleware(next ssh.Handler) ssh.Handler {
	return func(sess ssh.Session) {
		sessionId := sess.Context().SessionID()

		hashedPublicKey := fmt.Sprintf("%x", sha1.Sum(sess.PublicKey().Marshal()))

		// TODO: should be database directory on server
		homeDir, _ := os.UserHomeDir()

		// check for user and key in system database
		systemDBName := "system.db"
		systemDBDir := filepath.Join(homeDir, ".syringe")
		if err := os.MkdirAll(systemDBDir, 0755); err != nil {
			log.Error(
				"create tenant database directory",
				"session", sessionId,
				"systemDBDir", systemDBDir,
				"err", err,
			)

			sess.Stderr().Write([]byte(serrors.New("server", sessionId).Error()))
			return
		}

		systemDBPath := filepath.Join(systemDBDir, systemDBName)
		systemDB, err := database.NewConnection(systemDBPath)
		if err != nil {
			log.Error(
				"new system database connection",
				"session", sessionId,
				"systemDBPath", systemDBPath,
				"err", err,
			)

			sess.Stderr().Write([]byte(serrors.New("server", sessionId).Error()))
			return
		}

		systemDriver, err := iofs.New(database.SystemMigrations, "sql")
		if err != nil {
			log.Error(
				"new system driver",
				"session", sessionId,
				"err", err,
			)

			sess.Stderr().Write([]byte(serrors.New("server", sessionId).Error()))
			return
		}

		systemMigrator, err := database.NewMigration(systemDB, systemDriver)
		if err != nil {
			log.Error(
				"new system migration",
				"session", sessionId,
				"err", err,
			)

			sess.Stderr().Write([]byte(serrors.New("server", sessionId).Error()))
			return
		}

		if err := systemMigrator.Up(); err != nil {
			if !errors.Is(err, migrate.ErrNoChange) {
				log.Error(
					"run system migration",
					"session", sessionId,
					"err", err,
				)

				sess.Stderr().Write([]byte(serrors.New("server", sessionId).Error()))
				return
			}
		}

		systemStore := stores.NewSystemStore(systemDB)
		user, err := systemStore.GetUser(sess.Context().User(), hashedPublicKey)
		if err != nil || user == nil {
			log.Info("user not found, create user", "username", sess.Context().User(), "hashedPublicKey", hashedPublicKey)
			user = &stores.User{
				Username:      sess.Context().User(),
				PublicKeySHA1: hashedPublicKey,
				Email:         time.Now().GoString(),
			}
			if err := systemStore.CreateUser(user); err != nil {
				log.Fatal("create user failed", "err", err)
			}
		}

		log.Info("USER", "user", user)

		// TODO: check if a database exists, if not then send code to email and await entry before continuing
		tenantDBName := fmt.Sprintf("%x.db", hashedPublicKey)
		tenantDBDir := filepath.Join(homeDir, ".syringe")
		if err := os.MkdirAll(tenantDBDir, 0755); err != nil {
			log.Error(
				"create tenant database directory",
				"session", sessionId,
				"tenantDBDir", tenantDBDir,
				"err", err,
			)

			sess.Stderr().Write([]byte(serrors.New("server", sessionId).Error()))
			return
		}

		tenantDBPath := filepath.Join(tenantDBDir, tenantDBName)

		tenantDB, err := database.NewConnection(tenantDBPath)
		if err != nil {
			log.Error(
				"new tenant database connection",
				"session", sessionId,
				"tenantDBPath", tenantDBPath,
				"err", err,
			)

			sess.Stderr().Write([]byte(serrors.New("server", sessionId).Error()))
			return
		}

		driver, err := iofs.New(database.TenantMigrations, "sql")
		if err != nil {
			log.Error(
				"new driver",
				"session", sessionId,
				"err", err,
			)

			sess.Stderr().Write([]byte(serrors.New("server", sessionId).Error()))
			return
		}

		migrator, err := database.NewMigration(tenantDB, driver)
		if err != nil {
			log.Error(
				"new migration",
				"session", sessionId,
				"err", err,
			)

			sess.Stderr().Write([]byte(serrors.New("server", sessionId).Error()))
			return
		}

		if err := migrator.Up(); err != nil {
			if !errors.Is(err, migrate.ErrNoChange) {
				log.Error(
					"run migration",
					"session", sessionId,
					"err", err,
				)

				sess.Stderr().Write([]byte(serrors.New("server", sessionId).Error()))
				return
			}
		}

		store := stores.NewTenantStore(tenantDB)

		cmd := rootCmd(sess, store)

		cmd.SetArgs(sess.Command())

		cmd.SetIn(sess)
		cmd.SetOut(sess)
		cmd.SetErr(sess.Stderr())

		done := make(chan bool, 1)

		ctx, cancel := context.WithTimeout(sess.Context(), commandTimeout)
		defer cancel()

		go func() {
			if err := cmd.Execute(); err != nil {
				log.Error(
					"exec cmd",
					"session", sessionId,
					"err", err,
				)

				sess.Stderr().Write([]byte(serrors.New("cmd", sessionId).Error()))
				sess.Exit(1)
			}
			done <- true
		}()

		select {
		case <-ctx.Done():
			sess.Stderr().Write([]byte(serrors.New("timeout", sessionId).Error()))
			sess.Exit(1)
			return
		case <-done:
			// done
		}

		next(sess)
	}
}

func rootCmd(sess ssh.Session, s *stores.TenantStore) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "syringe",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(c *cobra.Command, args []string) error {
			return fmt.Errorf("no command specified")
		},
	}

	cmd.CompletionOptions.DisableDefaultCmd = true

	cmd.AddCommand(
		setCmd(s),
		getCmd(s),
		listCmd(s),
		removeCmd(s),
	)

	return cmd
}

func setCmd(s *stores.TenantStore) *cobra.Command {
	return &cobra.Command{
		Use:  "set",
		Args: cobra.ExactArgs(2),
		RunE: func(c *cobra.Command, args []string) error {
			// TODO: verify the value being saved is encrypted with a private key
			//       that corresponds to the public key so that we're not storing
			//       unencrypted data or data that can't be decrypted by the user

			if err := s.Set(&stores.Item{
				Key:   args[0],
				Value: args[1],
			}); err != nil {
				return fmt.Errorf("set value in store: %w", err)
			}

			return nil
		},
	}
}

func getCmd(s *stores.TenantStore) *cobra.Command {
	return &cobra.Command{
		Use:  "get",
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			item, err := s.Get(args[0])
			if err != nil {
				return fmt.Errorf("get value from store: %w", err)
			}

			if _, err := c.OutOrStdout().Write([]byte(item.Value)); err != nil {
				return fmt.Errorf("write value to stdout: %w", err)
			}

			return nil
		},
	}
}

func listCmd(s *stores.TenantStore) *cobra.Command {
	return &cobra.Command{
		Use:  "list",
		Args: cobra.ExactArgs(0),
		RunE: func(c *cobra.Command, args []string) error {
			items, err := s.List()
			if err != nil {
				return fmt.Errorf("list keys in store: %w", err)
			}

			keys := make([]string, len(items))
			for i, item := range items {
				keys[i] = item.Key
			}

			if _, err := c.OutOrStdout().Write(
				[]byte(strings.Join(keys, "\n")),
			); err != nil {
				return fmt.Errorf("write keys to stdout: %w", err)
			}

			return nil
		},
	}
}

func removeCmd(s *stores.TenantStore) *cobra.Command {
	return &cobra.Command{
		Use:  "remove",
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			if err := s.Remove(args[0]); err != nil {
				return fmt.Errorf("remove value from store: %w", err)
			}

			return nil
		},
	}
}
