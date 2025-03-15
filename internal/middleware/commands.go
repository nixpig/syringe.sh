package middleware

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	"github.com/golang-migrate/migrate/v4"
	"github.com/nixpig/syringe.sh/database"
	"github.com/nixpig/syringe.sh/internal/stores"
	"github.com/spf13/cobra"
)

// TODO: better strategy for logging, writing errors and exiting
// TODO: better logic for commands and switching when register should/shouldn't be available

func NewCmdMiddleware(systemStore *stores.SystemStore) wish.Middleware {
	return func(next ssh.Handler) ssh.Handler {
		return func(sess ssh.Session) {
			log.Debug(sess.RawCommand())
			cmd := rootCmd(sess)
			cmd.SetArgs(sess.Command())
			cmd.SetIn(sess)
			cmd.SetOut(sess)
			cmd.SetErr(sess.Stderr())

			sessionID := sess.Context().SessionID()
			username := sess.Context().User()

			publicKeyHash, ok := sess.Context().Value(contextKeyHash).(string)
			if !ok {
				sess.Stderr().Write([]byte("Error: failed to get public key"))
				sess.Exit(1)
			}

			email, ok := sess.Context().Value(contextKeyEmail).(string)
			if !ok {
				sess.Stderr().Write([]byte("Error: failed to get email"))
				sess.Exit(1)
				return
			}

			authenticated, ok := sess.Context().Value(contextKeyAuthenticated).(bool)
			if !ok {
				log.Debug("failed to get authenticated context", "session", sessionID)
				authenticated = false
			}

			if authenticated {
				db, err := tenantDB(publicKeyHash, sessionID)
				if err != nil {
					log.Error("connect to tenant database", "session", sessionID, "err", err)
					sess.Stderr().Write([]byte("Error: database connection error"))
					sess.Exit(1)
					return
				}

				tenantStore := stores.NewTenantStore(sess.Context(), db)
				cmd.AddCommand(
					setCmd(tenantStore),
					getCmd(tenantStore),
					listCmd(tenantStore),
					removeCmd(tenantStore),
				)
			} else {
				cmd.AddCommand(registerCmd(
					systemStore,
					username,
					email,
					publicKeyHash,
				))
			}

			doneCh := make(chan bool, 1)
			errCh := make(chan error, 1)

			go func() {
				if err := cmd.ExecuteContext(sess.Context()); err != nil {
					errCh <- err
					return
				}

				doneCh <- true
			}()

			select {
			case <-sess.Context().Done():
				log.Error("timeout", "session", sessionID)
				sess.Stderr().Write([]byte("Error: timed out"))
				sess.Exit(1)
				return

			case err := <-errCh:
				log.Error("cmd", "session", sessionID, "err", err)
				sess.Stderr().Write([]byte("Error: " + err.Error()))
				sess.Exit(1)
				return

			case <-doneCh:
				log.Debug("done", "session", sessionID)
				next(sess)
			}
		}
	}
}

func rootCmd(sess ssh.Session) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "syringe",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(c *cobra.Command, args []string) error {
			return fmt.Errorf("no command specified")
		},
	}

	cmd.CompletionOptions.HiddenDefaultCmd = true

	return cmd
}

func setCmd(s *stores.TenantStore) *cobra.Command {
	return &cobra.Command{
		Use:  "set",
		Args: cobra.ExactArgs(2),
		RunE: func(c *cobra.Command, args []string) error {
			if err := s.SetItem(&stores.Item{
				Key:   args[0],
				Value: args[1],
			}); err != nil {
				return err
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
			item, err := s.GetItemByKey(args[0])
			if err != nil {
				return err
			}

			c.OutOrStdout().Write([]byte(item.Value))
			return nil
		},
	}
}

func listCmd(s *stores.TenantStore) *cobra.Command {
	return &cobra.Command{
		Use:  "list",
		Args: cobra.ExactArgs(0),
		RunE: func(c *cobra.Command, args []string) error {
			items, err := s.ListItems()
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

func removeCmd(s *stores.TenantStore) *cobra.Command {
	return &cobra.Command{
		Use:  "remove",
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			if err := s.RemoveItemByKey(args[0]); err != nil {
				return err
			}

			return nil
		},
	}
}

func registerCmd(
	s *stores.SystemStore,
	username, email, publicKeyHash string,
) *cobra.Command {
	return &cobra.Command{
		Use:  "register",
		Args: cobra.ExactArgs(0),
		RunE: func(c *cobra.Command, args []string) error {
			if _, err := s.CreateUser(&stores.User{
				Username:      username,
				Email:         email,
				PublicKeySHA1: publicKeyHash,
				Verified:      false,
			}); err != nil {
				return err
			}

			return nil
		},
	}
}

// TODO: move this somewhere sensible!
func tenantDB(publicKeyHash, sessionID string) (*sql.DB, error) {
	tenantDBDir := os.Getenv("SYRINGE_DB_TENANT_DIR")
	tenantDBName := fmt.Sprintf("%x.db", publicKeyHash)
	tenantDBPath := filepath.Join(tenantDBDir, tenantDBName)

	tenantDB, err := database.NewConnection(tenantDBPath)
	if err != nil {
		return nil, err
	}

	migrator, err := database.NewMigration(tenantDB, database.TenantMigrations)
	if err != nil {
		return nil, err
	}

	if err := migrator.Up(); err != nil {
		if !errors.Is(err, migrate.ErrNoChange) {
			return nil, err
		}
	}

	return tenantDB, nil
}
