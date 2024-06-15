package server

import (
	"github.com/charmbracelet/ssh"
	"github.com/nixpig/syringe.sh/server/cmd"
)

func cobraHandler(s Server) func(next ssh.Handler) ssh.Handler {
	return func(next ssh.Handler) ssh.Handler {
		return func(sess ssh.Session) {
			db, err := newUserDBConnection(sess.PublicKey())
			if err != nil {
				s.logger.Err(err).
					Str("session", sess.Context().SessionID()).
					Msg("failed to obtain user database connection")
				sess.Stderr().Write([]byte("Failed to obtain database connection using the provided public key"))
				return
			}

			defer db.Close()

			if err := cmd.Execute(
				sess.PublicKey(),
				sess.Command(),
				sess,
				sess,
				sess.Stderr(),
				db,
			); err != nil {
				s.logger.Err(err).
					Str("session", sess.Context().SessionID()).
					Any("command", sess.Command()).
					Msg("failed to execute command")
			}

			next(sess)
		}
	}
}
