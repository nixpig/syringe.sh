package main

import (
	"crypto/sha1"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/charmbracelet/ssh"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/spf13/cobra"
)

func cmdMiddleware(next ssh.Handler) ssh.Handler {
	return func(sess ssh.Session) {
		dbName := fmt.Sprintf("%x.db", sha1.Sum(sess.PublicKey().Marshal()))

		homeDir, _ := os.UserHomeDir()

		// TODO: check if a database exists, if not then send code to email and await entry before continuing

		dbDir := filepath.Join(homeDir, ".syringe")
		if err := os.MkdirAll(dbDir, 0755); err != nil {
			log.Error(
				"create store directory",
				"session", sess.Context().SessionID(),
				"dbDir", dbDir,
				"err", err,
			)

			sess.Stderr().Write([]byte(ErrServer.Error()))
			return
		}

		dbPath := filepath.Join(dbDir, dbName)

		db, err := NewConnection(dbPath)
		if err != nil {
			log.Error(
				"new database connection",
				"session", sess.Context().SessionID(),
				"dbPath", dbPath,
				"err", err,
			)

			sess.Stderr().Write([]byte(ErrServer.Error()))
			return
		}

		driver, err := iofs.New(Migrations, "sql")
		if err != nil {
			log.Error(
				"new driver",
				"session", sess.Context().SessionID(),
				"err", err,
			)

			sess.Stderr().Write([]byte(ErrServer.Error()))
			return
		}

		migrator, err := NewMigration(db, driver)
		if err != nil {
			log.Error(
				"new migration",
				"session", sess.Context().SessionID(),
				"err", err,
			)

			sess.Stderr().Write([]byte(ErrServer.Error()))
			return
		}

		if err := migrator.Up(); err != nil {
			if !errors.Is(err, migrate.ErrNoChange) {
				log.Error(
					"run migration",
					"session", sess.Context().SessionID(),
					"err", err,
				)

				sess.Stderr().Write([]byte(ErrServer.Error()))
				return
			}
		}

		store := NewStore(db)

		cmd := rootCmd(sess, store)

		cmd.SetArgs(sess.Command())

		// NOTE: we don't pipe cobra errs, since we write custom error codes
		cmd.SetIn(sess)
		cmd.SetOut(sess)

		if err := cmd.Execute(); err != nil {
			log.Error(
				"root cmd exec",
				"session", sess.Context().SessionID(),
				"err", err,
			)

			sess.Stderr().Write([]byte(ErrCmd.Error()))
			sess.Exit(1)
			return
		}

		next(sess)
	}
}

func rootCmd(sess ssh.Session, store *Store) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "syringe",
		SilenceUsage: true,
		Run: func(c *cobra.Command, args []string) {
			log.Error(
				"root cmd exec",
				"session", sess.Context().SessionID(),
				"err", "no command specified",
			)

			sess.Stderr().Write([]byte(ErrCmd.Error()))
			sess.Exit(1)
		},
	}

	cmd.CompletionOptions.DisableDefaultCmd = true

	cmd.AddCommand(
		setCmd(sess, store),
		getCmd(sess, store),
		listCmd(sess, store),
		removeCmd(sess, store),
	)

	return cmd
}

func setCmd(sess ssh.Session, store *Store) *cobra.Command {
	return &cobra.Command{
		Use:  "set",
		Args: cobra.ExactArgs(2),
		Run: func(c *cobra.Command, args []string) {
			// TODO: verify the value being saved is encrypted with a private key
			//       that corresponds to the public key so that we're not storing
			//       unencrypted data or data that can't be decrypted by the user

			if err := store.Set(&Item{
				Key:   args[0],
				Value: args[1],
			}); err != nil {
				log.Error(
					"set cmd",
					"session", sess.Context().SessionID(),
					"err", err,
				)

				sess.Stderr().Write([]byte(ErrCmd.Error()))
				sess.Exit(1)
			}
		},
	}
}

func getCmd(sess ssh.Session, store *Store) *cobra.Command {
	return &cobra.Command{
		Use:  "get",
		Args: cobra.ExactArgs(1),
		Run: func(c *cobra.Command, args []string) {
			item, err := store.Get(args[0])
			if err != nil {
				log.Error(
					"get cmd",
					"session", sess.Context().SessionID(),
					"err", err,
				)

				sess.Stderr().Write([]byte(ErrCmd.Error()))
				sess.Exit(1)
			}

			if _, err := c.OutOrStdout().Write([]byte(item.Value)); err != nil {
				log.Error(
					"write get value",
					"session", sess.Context().SessionID(),
					"err", err,
				)
			}
		},
	}
}

func listCmd(sess ssh.Session, store *Store) *cobra.Command {
	return &cobra.Command{
		Use:  "list",
		Args: cobra.NoArgs,
		Run: func(c *cobra.Command, args []string) {
			items, err := store.List()
			if err != nil {
				log.Error(
					"list cmd",
					"session", sess.Context().SessionID(),
					"err", err,
				)

				sess.Stderr().Write([]byte(ErrCmd.Error()))
				sess.Exit(1)
			}

			keys := make([]string, len(items))
			for i, item := range items {
				keys[i] = item.Key
			}

			if _, err := c.OutOrStdout().Write([]byte(strings.Join(keys, "\n"))); err != nil {
				log.Error(
					"write keys list",
					"session", sess.Context().SessionID(),
					"err", err,
				)
			}
		},
	}
}

func removeCmd(sess ssh.Session, store *Store) *cobra.Command {
	return &cobra.Command{
		Use:  "remove",
		Args: cobra.ExactArgs(1),
		Run: func(c *cobra.Command, args []string) {
			if err := store.Remove(args[0]); err != nil {
				log.Error(
					"remove cmd",
					"session", sess.Context().SessionID(),
					"err", err,
				)

				sess.Stderr().Write([]byte(ErrCmd.Error()))
				sess.Exit(1)
			}
		},
	}
}
