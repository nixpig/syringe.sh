package server

import (
	"os"

	"github.com/charmbracelet/ssh"
	"github.com/nixpig/syringe.sh/server/cmd"
)

func cobraHandler(s Server) func(next ssh.Handler) ssh.Handler {
	return func(next ssh.Handler) ssh.Handler {
		return func(sess ssh.Session) {
			db, err := newUserDBConnection(sess.PublicKey())
			if err != nil {
				// todo: what to do here?
				return
			}

			defer db.Close()

			if err := cmd.Execute(
				sess.PublicKey(),
				sess.Command(),
				os.Stdin,
				sess,
				sess.Stderr(),
				db,
			); err != nil {
				s.logger.Err(err).
					Str("session", sess.Context().SessionID()).
					Msg("failed to execute command")
			}

			next(sess)
		}
	}
}
