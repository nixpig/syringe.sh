package secret

import (
	"bytes"
	"io"
	"os/exec"
	"strings"

	"github.com/nixpig/syringe.sh/internal/cli"
	"github.com/nixpig/syringe.sh/pkg"
	"github.com/nixpig/syringe.sh/pkg/ssh"
	"github.com/spf13/cobra"
)

func NewCLIHandlerSecretInject(host string, port int, out io.Writer) pkg.CobraHandler {
	return func(cmd *cobra.Command, args []string) error {
		w := bytes.NewBufferString("")

		// TODO: this feels hacky
		injectHandler := cli.NewHandlerCLI(
			host,
			port,
			w,
			ssh.NewSSHClient,
		)

		if err := injectHandler(cmd, args); err != nil {
			return err
		}

		injection, err := io.ReadAll(w)
		if err != nil {
			return err
		}

		env := strings.Split(string(injection), " ")

		var command string
		var arguments []string

		if len(args) > 0 {
			command = args[0]
		}

		if len(args) > 1 {
			arguments = args[1:]
		}

		hostCmd := exec.Command(command, arguments...)
		hostCmd.Env = append(hostCmd.Environ(), env...)
		hostCmd.Stdout = out

		if err := hostCmd.Run(); err != nil {
			cmd.SilenceUsage = true
			return err
		}

		return nil
	}
}
