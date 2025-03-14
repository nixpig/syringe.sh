package middleware

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/log"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	"github.com/golang-migrate/migrate/v4"
	"github.com/nixpig/syringe.sh/database"
	"github.com/nixpig/syringe.sh/internal/stores"
)

const (
	commandTimeout = time.Second * 5
)

func NewCmdMiddleware(systemStore *stores.SystemStore) wish.Middleware {
	return func(next ssh.Handler) ssh.Handler {
		return func(sess ssh.Session) {
			cmd := sess.Command()
			if len(cmd) == 0 {
				sess.Stderr().Write([]byte("Error: no command specified"))
				sess.Exit(1)
				return
			}

			sessionID := sess.Context().SessionID()
			username := sess.Context().User()

			publicKeyHash, ok := sess.Context().Value("publicKeyHash").(string)
			if !ok {
				sess.Stderr().Write([]byte("Error: failed to get public key"))
				sess.Exit(1)
			}

			email, ok := sess.Context().Value("email").(string)
			if !ok {
				sess.Stderr().Write([]byte("Error: failed to get email"))
				sess.Exit(1)
				return
			}

			authenticated, ok := sess.Context().Value("authenticated").(bool)
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

				switch cmd[0] {
				case "set":
					if len(cmd) != 3 {
						sess.Stderr().Write([]byte(fmt.Sprintf("Error: 'set' accepts 2 args, received %d", len(cmd)-1)))
						sess.Exit(1)
						return
					}

					if err := setCmd(tenantStore, cmd[1], cmd[2]); err != nil {
						log.Debug("set", "session", sessionID, "err", err)
						sess.Stderr().Write([]byte(fmt.Sprintf("Error: failed to set '%s'", cmd[1])))
						sess.Exit(1)
						return
					}
					next(sess)
					return

				case "get":
					if len(cmd) != 2 {
						sess.Stderr().Write([]byte(fmt.Sprintf("Error: 'get' accepts 1 args, received %d", len(cmd)-1)))
						sess.Exit(1)
						return
					}

					value, err := getCmd(tenantStore, cmd[1])
					if err != nil {
						log.Debug("get", "session", sessionID, "key", cmd[1], "err", err)
						sess.Stderr().Write([]byte(fmt.Sprintf("Error: failed to get '%s'", cmd[1])))
						sess.Exit(1)
						return
					}

					sess.Write([]byte(value))
					next(sess)
					return

				case "remove":
					if len(cmd) != 2 {
						sess.Stderr().Write([]byte(fmt.Sprintf("Error: 'remove' accepts 1 args, received %d", len(cmd)-1)))
						sess.Exit(1)
						return
					}

					if err := removeCmd(tenantStore, cmd[1]); err != nil {
						log.Debug("remove", "session", sessionID, "key", cmd[1], "err", err)
						sess.Stderr().Write([]byte(fmt.Sprintf("Error: failed to remove '%s'", cmd[1])))
						sess.Exit(1)
						return
					}

					next(sess)
					return

				case "list":
					if len(cmd) != 1 {
						sess.Stderr().Write([]byte(fmt.Sprintf("Error: 'list' accepts 0 args, received %d", len(cmd)-1)))
						sess.Exit(1)
						return
					}

					list, err := listCmd(tenantStore)
					if err != nil {
						log.Debug("list", "session", sessionID, "err", err)
						sess.Stderr().Write([]byte("Error: failed to list"))
						sess.Exit(1)
						return
					}
					sess.Write([]byte(strings.Join(list, "\n")))
					next(sess)
					return

				default:
					sess.Stderr().Write([]byte(fmt.Sprintf("Error: unknown command '%s'", cmd[0])))
					sess.Exit(1)
					return

				}
			}

			switch cmd[0] {
			case "register":
				if err := registerCmd(
					systemStore,
					username,
					email,
					publicKeyHash,
				); err != nil {
					log.Debug(
						"register",
						"session", sessionID,
						"username", username,
						"email", email,
						"publicKeyHash", publicKeyHash,
						"err", err,
					)
					sess.Stderr().Write([]byte("Error: failed to register"))
					sess.Exit(1)
					return
				}
				sess.Exit(0)
				return
			default:
				sess.Stderr().Write([]byte("Error: you are not authenticated"))
				sess.Exit(1)
				return
			}
		}
	}
}

func setCmd(s *stores.TenantStore, key, value string) error {
	if err := s.SetItem(&stores.Item{
		Key:   key,
		Value: value,
	}); err != nil {
		return err
	}

	return nil
}

func getCmd(s *stores.TenantStore, key string) (string, error) {
	item, err := s.GetItemByKey(key)
	if err != nil {
		return "", err
	}

	return item.Value, err
}

func listCmd(s *stores.TenantStore) ([]string, error) {
	items, err := s.ListItems()
	if err != nil {
		return []string{}, err
	}

	keys := make([]string, len(items))
	for i, item := range items {
		keys[i] = item.Key
	}

	return keys, nil
}

func removeCmd(s *stores.TenantStore, key string) error {
	if err := s.RemoveItemByKey(key); err != nil {
		return err
	}

	return nil
}

func registerCmd(
	s *stores.SystemStore,
	username, email, publicKeyHash string,
) error {
	if _, err := s.CreateUser(&stores.User{
		Username:      username,
		Email:         email,
		PublicKeySHA1: publicKeyHash,
		Verified:      false,
	}); err != nil {
		return err
	}

	// TODO: add email verification

	// Note: tenant database will just be created automatically if it doesn't
	// exist when first command is run
	// ...see tenantDB function

	return nil
}

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
