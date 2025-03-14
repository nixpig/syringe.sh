package server

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
	_ "github.com/mattn/go-sqlite3"
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
				log.Debug("no command")
				next(sess)
				return
			}

			sessionID := sess.Context().SessionID()
			username := sess.Context().User()

			publicKeyHash, ok := sess.Context().Value("publicKeyHash").(string)
			if !ok {
				log.Error("failed to get public key hash from context")
				sess.Stderr().Write([]byte("Error: "))
				sess.Exit(1)
			}

			email, ok := sess.Context().Value("email").(string)
			if !ok {
				log.Error("failed to get email from context")
				sess.Stderr().Write([]byte("Error: "))
				sess.Exit(1)
				return
			}

			authenticated, ok := sess.Context().Value("authenticated").(bool)
			if !ok {
				log.Warn("failed to get authenticated from context, defaulting to 'false'")
				authenticated = false
			}

			if authenticated {
				db, err := tenantDB(publicKeyHash, sessionID)
				if err != nil {
					log.Error("connect to tenant db (authenticated)", "err", err)
					sess.Stderr().Write([]byte("Error: failed to connect to database"))
					sess.Exit(1)
					return

				}

				tenantStore := stores.NewTenantStore(db)

				switch cmd[0] {
				case "set":
					if len(cmd) != 3 {
						sess.Stderr().Write([]byte(fmt.Sprintf("Error: 'set' accepts 2 args, received %d", len(cmd)-1)))
						sess.Exit(1)
						return
					}

					if err := setCmd(tenantStore, cmd[1], cmd[2]); err != nil {
						sess.Stderr().Write([]byte(fmt.Sprintf("Error: 'set': %s", err)))
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
						sess.Stderr().Write([]byte(fmt.Sprintf("Error: 'get': %s", err)))
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
						sess.Stderr().Write([]byte(fmt.Sprintf("Error: 'remove': %s", err)))
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
						sess.Stderr().Write([]byte(fmt.Sprintf("Error: 'list': %s", err)))
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
					log.Debug("register", "err", err)
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

			// done := make(chan bool, 1)
			//
			// ctx, cancel := context.WithTimeout(sess.Context(), commandTimeout)
			// defer cancel()
			//
			// go func() {
			// 	if err := cmd.Execute(); err != nil {
			// 		log.Error(
			// 			"exec cmd",
			// 			"session", sessionID,
			// 			"err", err,
			// 		)
			//
			// 		sess.Stderr().Write([]byte(serrors.New(
			// 			"cmd", "encountered error while running command", sessionID,
			// 		).Error()))
			// 		sess.Exit(1)
			// 	}
			// 	done <- true
			// }()

			// select {
			// case <-ctx.Done():
			// 	log.Error("request timed out", "session", sessionID, "err", err)
			// 	sess.Stderr().Write([]byte(serrors.New(
			// 		"timeout", "request timed out", sessionID,
			// 	).Error()))
			// 	sess.Exit(1)
			// 	return
			// case <-done:
			// 	next(sess)
			// }
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
