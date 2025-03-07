package cli

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/nixpig/syringe.sh/api"
	"github.com/nixpig/syringe.sh/internal/items"
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

func New(v *viper.Viper) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:     "syringe",
		Short:   "Encrypted key-value store",
		Version: "",
		PersistentPreRunE: func(c *cobra.Command, args []string) error {
			applyFlags(c, v)

			debugLevel, _ := c.Flags().GetBool("debug")
			if debugLevel {
				log.SetLevel(log.DebugLevel)
			}

			identityPath, _ := c.Flags().GetString(identityFlag)
			storePath, _ := c.Flags().GetString(storeFlag)
			log.Debug("flags", identityFlag, identityPath, storeFlag, storePath)

			return nil
		},
	}

	rootCmd.PersistentFlags().BoolP("debug", "d", false, "Set log level to debug")

	rootCmd.PersistentFlags().StringP(identityFlag, "i", "", "Path to SSH key")
	rootCmd.MarkPersistentFlagRequired(identityFlag)

	rootCmd.PersistentFlags().StringP(storeFlag, "s", "", "Store as parent path or parent URL")
	rootCmd.MarkPersistentFlagRequired(storeFlag)

	bindFlags(rootCmd, v)

	rootCmd.AddCommand(
		setCmd,
		getCmd,
		listCmd,
		removeCmd,
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

		store, _ := c.Flags().GetString(storeFlag)
		a, err := api.New(store)
		if err != nil {
			return fmt.Errorf("create new api: %w", err)
		}
		defer a.Close()

		return set(
			c.Context(),
			a,
			ssh.NewEncryptor(publicKey),
			items.New(args[0], args[1]),
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

		store, _ := c.Flags().GetString(storeFlag)
		a, err := api.New(store)
		if err != nil {
			return fmt.Errorf("create new api: %w", err)
		}
		defer a.Close()

		item, err := get(
			c.Context(),
			a,
			ssh.NewDecryptor(privateKey),
			args[0],
		)
		if err != nil {
			return err
		}

		c.OutOrStdout().Write([]byte(item.Value))

		return nil
	},
}

var removeCmd = &cobra.Command{
	Use:     "remove [flags] KEY",
	Short:   "Remove a record from the store",
	Args:    cobra.ExactArgs(1),
	Example: "  syringe remove username",
	RunE: func(c *cobra.Command, args []string) error {
		store, _ := c.Flags().GetString(storeFlag)
		a, err := api.New(store)
		if err != nil {
			return fmt.Errorf("create new api: %w", err)
		}
		defer a.Close()

		return remove(c.Context(), a, args[0])
	},
}

var listCmd = &cobra.Command{
	Use:     "list [flags]",
	Short:   "List all records in store",
	Args:    cobra.ExactArgs(0),
	Example: "  syringe list",
	RunE: func(c *cobra.Command, args []string) error {
		store, _ := c.Flags().GetString(storeFlag)
		a, err := api.New(store)
		if err != nil {
			return fmt.Errorf("create new api: %w", err)
		}
		defer a.Close()

		list, err := list(c.Context(), a)
		if err != nil {
			return fmt.Errorf("list: %w", err)
		}

		keys := make([]string, len(list))
		for i, l := range list {
			keys[i] = l.Key
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
