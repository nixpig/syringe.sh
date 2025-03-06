package cmd

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/nixpig/syringe.sh/internal/database"
	"github.com/nixpig/syringe.sh/internal/migrations"
	"github.com/nixpig/syringe.sh/internal/store"
	"github.com/nixpig/syringe.sh/pkg/ssh"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"golang.org/x/term"
)

const (
	identityFlag = "identity"
	storeFlag    = "store"
)

var db *sql.DB

func New(v *viper.Viper) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:     "syringe",
		Short:   "Encrypted key-value store",
		Version: "",
		PersistentPreRunE: func(c *cobra.Command, args []string) error {
			var err error

			applyFlags(c, v)

			identityPath, _ := c.Flags().GetString(identityFlag)
			storePath, _ := c.Flags().GetString(storeFlag)
			log.Debug("flags", identityFlag, identityPath, storeFlag, storePath)

			dbDir := filepath.Dir(storePath)
			if err := os.MkdirAll(dbDir, 0755); err != nil {
				return fmt.Errorf("create store directory: %w", err)
			}

			db, err = database.NewConnection(storePath)
			if err != nil {
				return fmt.Errorf("new database connection: %w", err)
			}

			driver, err := iofs.New(migrations.Migrations, "sql")
			if err != nil {
				return fmt.Errorf("new driver: %w", err)
			}

			migrator, err := database.NewMigration(db, driver)
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
		PersistentPostRun: func(c *cobra.Command, args []string) {
			if err := db.Close(); err != nil {
				log.Error("close database", "err", err)
			}
		},
	}

	rootCmd.PersistentFlags().StringP(
		identityFlag,
		"i",
		"",
		"Path to SSH key",
	)
	rootCmd.MarkPersistentFlagRequired(identityFlag)

	rootCmd.PersistentFlags().StringP(
		storeFlag,
		"s",
		"",
		"Store as name, path or URL",
	)
	rootCmd.MarkPersistentFlagRequired(storeFlag)

	bindFlags(rootCmd, v)

	rootCmd.AddCommand(
		setCmd,
		getCmd,
		listCmd,
		deleteCmd,
	)

	return rootCmd
}

var setCmd = &cobra.Command{
	Use:     "set [flags] KEY VALUE",
	Short:   "Set a key-value",
	Args:    cobra.ExactArgs(2),
	Example: "  syringe set username nixpig",
	RunE: func(c *cobra.Command, args []string) error {
		identity, _ := c.Flags().GetString(identityFlag)

		publicKey, err := ssh.GetPublicKey(identity + ".pub")
		if err != nil {
			return fmt.Errorf("get public key: %w", err)
		}

		return Set(
			c.Context(),
			store.NewSqliteStore(db),
			ssh.NewEncryptor(publicKey),
			args[0],
			args[1],
		)
	},
}

var getCmd = &cobra.Command{
	Use:     "get [flags] KEY",
	Short:   "Get a value from the store",
	Args:    cobra.ExactArgs(1),
	Example: "  syringe get username",
	RunE: func(c *cobra.Command, args []string) error {
		identity, _ := c.Flags().GetString(identityFlag)

		privateKey, err := ssh.GetPrivateKey(identity, term.ReadPassword)
		if err != nil {
			return fmt.Errorf("get private key: %w", err)
		}

		value, err := Get(
			c.Context(),
			store.NewSqliteStore(db),
			ssh.NewDecryptor(privateKey),
			args[0],
		)
		if err != nil {
			return err
		}

		c.OutOrStdout().Write([]byte(value))

		return nil
	},
}

var deleteCmd = &cobra.Command{
	Use:     "delete [flags] KEY",
	Short:   "Delete a record from the store",
	Args:    cobra.ExactArgs(1),
	Example: "  syringe delete username",
	RunE: func(c *cobra.Command, args []string) error {
		return Delete(c.Context(), store.NewSqliteStore(db), args[0])
	},
}

var listCmd = &cobra.Command{
	Use:     "list [flags]",
	Short:   "List all records in store",
	Args:    cobra.ExactArgs(0),
	Example: "  syringe list",
	RunE: func(c *cobra.Command, args []string) error {
		keys, err := List(c.Context(), store.NewSqliteStore(db))
		if err != nil {
			return fmt.Errorf("list: %w", err)
		}

		c.OutOrStdout().Write([]byte(strings.Join(keys, "\n")))

		return nil
	},
}

func bindFlags(c *cobra.Command, v *viper.Viper) {
	c.PersistentFlags().VisitAll(func(f *pflag.Flag) {
		v.BindPFlag(f.Name, f)
	})
}

func applyFlags(c *cobra.Command, v *viper.Viper) {
	c.Flags().VisitAll(func(f *pflag.Flag) {
		if v.IsSet(f.Name) {
			c.Flags().Set(f.Name, v.GetString(f.Name))
		}
	})
}
