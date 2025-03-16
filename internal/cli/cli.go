package cli

import (
	"bytes"
	"fmt"
	"io"
	"net/mail"
	"os"
	"os/user"
	"path/filepath"
	"strings"

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
	configFlag   = "config"

	defaultHost = "ssh.syringe.sh"
	defaultPort = 2323
)

func New(v *viper.Viper) *cobra.Command {
	a := api.New()

	rootCmd := &cobra.Command{
		Use:          "syringe",
		Short:        "Encrypted key-value store",
		Version:      "",
		SilenceUsage: true,
		PersistentPreRunE: func(c *cobra.Command, args []string) error {
			applyFlags(c, v)

			configPath := v.GetString(configFlag)
			v.SetConfigType("env")
			v.SetConfigFile(configPath)
			v.ReadInConfig()

			identity := v.GetString(identityFlag)
			if identity == "" {
				c.Help()
				return fmt.Errorf("no identity")
			}

			host := v.GetString(hostFlag)
			if host == "" {
				c.Help()
				return fmt.Errorf("no host")
			}

			port := v.GetInt(portFlag)
			if port < 1 || port > 65535 {
				c.Help()
				return fmt.Errorf("invalid port number")
			}

			username := v.GetString(usernameFlag)
			if username == "" {
				c.Help()
				return fmt.Errorf("username is empty")
			}

			email := v.GetString(emailFlag)
			if email == "" {
				f, err := os.ReadFile(identity + ".pub")
				if err == nil {
					parts := strings.Split(string(f), " ")
					if len(parts) > 0 {
						email = strings.TrimSpace(parts[len(parts)-1])
					}
				}

			}
			if _, err := mail.ParseAddress(email); err != nil {
				c.Help()
				return fmt.Errorf("invalid email")
			}

			authMethod, err := ssh.AuthMethod(identity, c.OutOrStdout())
			if err != nil {
				return fmt.Errorf("failed to create auth method: %w", err)
			}

			client, err := ssh.NewSSHClient(
				host,
				port,
				username,
				authMethod,
				filepath.Join(os.Getenv("HOME"), ".ssh", "known_hosts"),
			)
			if err != nil {
				return fmt.Errorf("failed to create ssh client: %w", err)
			}

			a.SetClient(client)
			a.SetOut(c.OutOrStdout())

			return nil
		},
		PersistentPostRun: func(c *cobra.Command, args []string) {
			a.Close()
		},
	}

	defaultUsername := ""
	currentUser, err := user.Current()
	if err == nil {
		defaultUsername = currentUser.Username
	}

	defaultConfigPath := ""
	configDir, _ := os.UserConfigDir()
	if configDir != "" {
		defaultConfigPath = filepath.Join(configDir, "syringe.cfg")
	}

	rootCmd.CompletionOptions.HiddenDefaultCmd = true
	rootCmd.PersistentFlags().StringP(identityFlag, "i", "", "Path to SSH key")
	rootCmd.PersistentFlags().StringP(usernameFlag, "u", defaultUsername, "Username")
	rootCmd.PersistentFlags().StringP(emailFlag, "e", "", "Email")
	rootCmd.PersistentFlags().StringP(hostFlag, "d", defaultHost, "Host")
	rootCmd.PersistentFlags().IntP(portFlag, "p", defaultPort, "Port")
	rootCmd.PersistentFlags().StringP(configFlag, "c", defaultConfigPath, "Config file location")

	bindFlags(rootCmd, v)

	rootCmd.AddCommand(
		registerCmd(v, a),
		setCmd(v, a),
		getCmd(v, a),
		listCmd(v, a),
		removeCmd(v, a),
	)

	return rootCmd
}

func registerCmd(v *viper.Viper, a *api.HostAPI) *cobra.Command {
	return &cobra.Command{
		Use:   "register [flags]",
		Short: "Register a user and key",
		Args:  cobra.ExactArgs(0),
		RunE: func(c *cobra.Command, args []string) error {
			return a.Register()
		},
	}
}

func setCmd(v *viper.Viper, a *api.HostAPI) *cobra.Command {
	return &cobra.Command{
		Use:     "set [flags] KEY VALUE",
		Short:   "Set a key-value",
		Args:    cobra.ExactArgs(2),
		Example: "  syringe set username nixpig",
		RunE: func(c *cobra.Command, args []string) error {
			identity := v.GetString(identityFlag)
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

func getCmd(v *viper.Viper, a *api.HostAPI) *cobra.Command {
	return &cobra.Command{
		Use:     "get [flags] KEY",
		Short:   "Get a value from the store",
		Args:    cobra.ExactArgs(1),
		Example: "  syringe get username",
		RunE: func(c *cobra.Command, args []string) error {
			identity := v.GetString(identityFlag)

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

func removeCmd(v *viper.Viper, a *api.HostAPI) *cobra.Command {
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

func listCmd(v *viper.Viper, a *api.HostAPI) *cobra.Command {
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
