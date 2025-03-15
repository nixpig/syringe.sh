package cli

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/user"
	"path/filepath"

	"github.com/charmbracelet/log"
	"github.com/nixpig/syringe.sh/internal/api"
	"github.com/nixpig/syringe.sh/pkg/ssh"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"golang.org/x/term"
)

const (
	identityFlag = "identity"
	usernameFlag = "username"
	emailFlag    = "email"
	hostFlag     = "host"
	portFlag     = "port"
)

var a *api.HostAPI

func New(v *viper.Viper) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:     "syringe",
		Short:   "Encrypted key-value store",
		Version: "",
		PersistentPreRunE: func(c *cobra.Command, args []string) error {
			applyFlags(c, v)

			authMethod, err := ssh.AuthMethod(v.GetString(identityFlag), c.OutOrStdout())
			if err != nil {
				log.Fatal("failed to create auth method: %s", err)
			}

			client, err := ssh.NewSSHClient(
				v.GetString(hostFlag),
				v.GetInt(portFlag),
				v.GetString(usernameFlag),
				authMethod,
				filepath.Join(os.Getenv("HOME"), ".ssh", "known_hosts"),
			)
			if err != nil {
				log.Fatal("failed to create ssh client", "err", err)
			}

			a = api.New(client, c.OutOrStdout())
			a.SetOut(c.OutOrStdout())

			return nil
		},
		PersistentPostRun: func(c *cobra.Command, args []string) {
			a.Close()
		},
	}

	username := ""
	currentUser, err := user.Current()
	if err == nil {
		username = currentUser.Username
	}

	rootCmd.CompletionOptions.HiddenDefaultCmd = true
	rootCmd.PersistentFlags().StringP(identityFlag, "i", "", "Path to SSH key")
	rootCmd.PersistentFlags().StringP(usernameFlag, "u", username, "Username")
	rootCmd.PersistentFlags().StringP(emailFlag, "e", "", "Email")
	rootCmd.PersistentFlags().StringP(hostFlag, "d", "localhost", "Host")
	rootCmd.PersistentFlags().IntP(portFlag, "p", 22, "Port")

	rootCmd.MarkPersistentFlagRequired(identityFlag)
	rootCmd.MarkPersistentFlagRequired(usernameFlag)
	// rootCmd.MarkPersistentFlagRequired(emailFlag)
	rootCmd.MarkPersistentFlagRequired(hostFlag)
	rootCmd.MarkPersistentFlagRequired(portFlag)

	bindFlags(rootCmd, v)

	rootCmd.AddCommand(
		registerCmd(),
		setCmd(),
		getCmd(),
		listCmd(),
		removeCmd(),
	)

	return rootCmd
}

func registerCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "register [flags]",
		Short: "Register a user and key",
		Args:  cobra.ExactArgs(0),
		RunE: func(c *cobra.Command, args []string) error {
			return a.Register()
		},
	}
}

func setCmd() *cobra.Command {
	return &cobra.Command{
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

			encrypt := ssh.NewEncryptor(publicKey)

			encryptedValue, err := encrypt(args[1])
			if err != nil {
				return fmt.Errorf("encrypt: %w", err)
			}

			if err := a.Set(args[0], encryptedValue); err != nil {
				return fmt.Errorf("set '%s' in store: %w", args[0], err)
			}

			return nil
		},
	}
}

func getCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "get [flags] KEY",
		Short:   "Get a value from the store",
		Args:    cobra.ExactArgs(1),
		Example: "  syringe get username",
		RunE: func(c *cobra.Command, args []string) error {
			identity, _ := c.Flags().GetString(identityFlag)
			privateKey, err := ssh.GetPrivateKey(
				identity, c.OutOrStderr(), term.ReadPassword,
			)
			if err != nil {
				return fmt.Errorf("failed to get private key from identity: %w", err)
			}

			decrypt := ssh.NewDecryptor(privateKey)

			var b bytes.Buffer
			a.SetOut(io.Writer(&b))

			err = a.Get(args[0])
			if err != nil {
				return err
			}

			d, err := io.ReadAll(&b)
			if err != nil {
				return err
			}

			decryptedValue, err := decrypt(string(d))
			if err != nil {
				return err
			}

			c.OutOrStdout().Write([]byte(decryptedValue))

			return nil
		},
	}
}

func removeCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "remove [flags] KEY",
		Short:   "Remove a record from the store",
		Args:    cobra.ExactArgs(1),
		Example: "  syringe remove username",
		RunE: func(c *cobra.Command, args []string) error {
			return a.Remove(args[0])
		},
	}
}

func listCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "list [flags]",
		Short:   "List all records in store",
		Args:    cobra.ExactArgs(0),
		Example: "  syringe list",
		RunE: func(c *cobra.Command, args []string) error {
			return a.List()
		},
	}
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
