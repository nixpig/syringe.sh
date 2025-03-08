package main

import (
	"crypto/sha1"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/ssh"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/spf13/cobra"
)

var db *sql.DB

func cmdMiddleware(next ssh.Handler) ssh.Handler {
	return func(sess ssh.Session) {
		cmd := rootCmd(sess)

		cmd.SetArgs(sess.Command())
		cmd.SetIn(sess)
		cmd.SetOut(sess)
		cmd.SetErr(sess.Stderr())

		if err := cmd.Execute(); err != nil {
			sess.Exit(1)
			return
		}

		next(sess)
	}
}

func rootCmd(sess ssh.Session) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "syringe",
		SilenceUsage: true,
		PersistentPreRunE: func(c *cobra.Command, args []string) error {
			var err error
			c.AddCommand()

			dbName := fmt.Sprintf("%x.db", sha1.Sum(sess.PublicKey().Marshal()))

			homeDir, _ := os.UserHomeDir()

			// TODO: check if a database exists, if not then send code to email and await entry before continuing

			dbDir := filepath.Join(homeDir, ".syringe")
			if err := os.MkdirAll(dbDir, 0755); err != nil {
				return fmt.Errorf("create store directory: %w", err)
			}

			dbPath := filepath.Join(dbDir, dbName)

			db, err = NewConnection(dbPath)
			if err != nil {
				return fmt.Errorf("new database connection: %w", err)
			}

			driver, err := iofs.New(Migrations, "sql")
			if err != nil {
				return fmt.Errorf("new driver: %w", err)
			}

			migrator, err := NewMigration(db, driver)
			if err != nil {
				return fmt.Errorf("create new migration: %w", err)
			}

			if err := migrator.Up(); err != nil {
				if errors.Is(err, migrate.ErrNoChange) {

				} else {
					return fmt.Errorf("run migration: %w", err)
				}
			}

			return nil
		},
		RunE: func(c *cobra.Command, args []string) error {
			return fmt.Errorf("no command specified")
		},
	}

	cmd.CompletionOptions.DisableDefaultCmd = true

	cmd.AddCommand(setCmd(), getCmd(), listCmd(), removeCmd())

	return cmd
}

func setCmd() *cobra.Command {
	return &cobra.Command{
		Use:  "set",
		Args: cobra.ExactArgs(2),
		RunE: func(c *cobra.Command, args []string) error {
			store := NewStore(db)

			// TODO: verify the value being saved is encrypted with a private key
			//       that corresponds to the public key so that we're not storing
			//       unencrypted data

			return store.Set(&Item{
				Key:   args[0],
				Value: args[1],
			})
		},
	}
}

func getCmd() *cobra.Command {
	return &cobra.Command{
		Use:  "get",
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			store := NewStore(db)

			item, err := store.Get(args[0])
			if err != nil {
				return err
			}

			c.OutOrStdout().Write([]byte(item.Value))
			return nil
		},
	}
}

func listCmd() *cobra.Command {
	return &cobra.Command{
		Use:  "list",
		Args: cobra.NoArgs,
		RunE: func(c *cobra.Command, args []string) error {
			store := NewStore(db)
			items, err := store.List()
			if err != nil {
				return err
			}

			keys := make([]string, len(items))
			for i, item := range items {
				keys[i] = item.Key
			}

			c.OutOrStdout().Write([]byte(strings.Join(keys, "\n")))
			return nil
		},
	}
}

func removeCmd() *cobra.Command {
	return &cobra.Command{
		Use:  "remove",
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			store := NewStore(db)
			return store.Remove(args[0])
		},
	}
}
