package middleware

import (
	"database/sql"
	"errors"
	"fmt"
	"net/mail"
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

func NewCmdMiddleware(systemStore *stores.SystemStore) wish.Middleware {
	return func(next ssh.Handler) ssh.Handler {
		return func(sess ssh.Session) {
			log.Debug(sess.RawCommand())
			cmd := rootCmd()
			cmd.SetArgs(sess.Command())
			cmd.SetIn(sess)
			cmd.SetOut(sess)
			cmd.SetErr(sess.Stderr())

			sessionID := sess.Context().SessionID()

			sess.Context().SetValue(contextKeyUsername, sess.Context().User())

			publicKeyHash, ok := sess.Context().Value(contextKeyHash).(string)
			if !ok {
				sess.Stderr().Write([]byte("failed to get public key"))
				sess.Exit(1)
			}

			db, err := tenantDB(publicKeyHash)
			if err != nil {
				log.Error("connect to tenant database", "session", sessionID, "err", err)
				sess.Stderr().Write([]byte("database connection error"))
				sess.Exit(1)
				return
			}

			tenantStore := stores.NewTenantStore(db)
			cmd.AddCommand(
				setCmd(tenantStore),
				getCmd(tenantStore),
				listCmd(tenantStore),
				removeCmd(tenantStore),
				registerCmd(systemStore),
			)
			cmd.AddCommand()

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
				sess.Stderr().Write([]byte("timed out"))
				sess.Exit(1)
				return

			case err := <-errCh:
				log.Error("cmd", "session", sessionID, "err", err)
				sess.Stderr().Write([]byte(err.Error()))
				sess.Exit(1)
				return

			case <-doneCh:
				log.Debug("done", "session", sessionID)
				next(sess)
			}
		}
	}
}

func rootCmd() *cobra.Command {
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
		PreRunE: func(c *cobra.Command, args []string) error {
			authenticated, ok := c.Context().Value(contextKeyAuthenticated).(bool)
			if !ok || !authenticated {
				return fmt.Errorf("not authenticated")
			}
			return nil
		},
		RunE: func(c *cobra.Command, args []string) error {
			if err := s.SetItem(
				c.Context(),
				&stores.Item{
					Key:   args[0],
					Value: args[1],
				},
			); err != nil {
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
		PreRunE: func(c *cobra.Command, args []string) error {
			authenticated, ok := c.Context().Value(contextKeyAuthenticated).(bool)
			if !ok || !authenticated {
				return fmt.Errorf("not authenticated")
			}
			return nil
		},
		RunE: func(c *cobra.Command, args []string) error {
			item, err := s.GetItemByKey(c.Context(), args[0])
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
		PreRunE: func(c *cobra.Command, args []string) error {
			authenticated, ok := c.Context().Value(contextKeyAuthenticated).(bool)
			if !ok || !authenticated {
				return fmt.Errorf("not authenticated")
			}
			return nil
		},
		RunE: func(c *cobra.Command, args []string) error {
			items, err := s.ListItems(c.Context())
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
		PreRunE: func(c *cobra.Command, args []string) error {
			authenticated, ok := c.Context().Value(contextKeyAuthenticated).(bool)
			if !ok || !authenticated {
				return fmt.Errorf("not authenticated")
			}
			return nil
		},
		RunE: func(c *cobra.Command, args []string) error {
			if err := s.RemoveItemByKey(c.Context(), args[0]); err != nil {
				return err
			}

			return nil
		},
	}
}

func registerCmd(s *stores.SystemStore) *cobra.Command {
	return &cobra.Command{
		Use:  "register",
		Args: cobra.ExactArgs(0),
		PreRunE: func(c *cobra.Command, args []string) error {
			authenticated, ok := c.Context().Value(contextKeyAuthenticated).(bool)
			if ok && authenticated {
				return fmt.Errorf("already registered")
			}
			return nil
		},
		RunE: func(c *cobra.Command, args []string) error {
			username, ok := c.Context().Value(contextKeyUsername).(string)
			if !ok {
				return fmt.Errorf("failed to get username")
			}

			publicKeyHash, ok := c.Context().Value(contextKeyHash).(string)
			if !ok {
				return fmt.Errorf("failed to get public key")
			}

			// TODO: this needs to be passed from client/maybe username should just be email??
			email := "nixpig@example.org"
			if _, err := mail.ParseAddress(email); err != nil {
				return fmt.Errorf("invalid email address")
			}

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
func tenantDB(publicKeyHash string) (*sql.DB, error) {
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
