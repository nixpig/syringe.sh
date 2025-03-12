package server

import (
	"slices"

	"github.com/charmbracelet/log"
	"github.com/charmbracelet/ssh"
	"github.com/nixpig/syringe.sh/internal/serrors"
)

var allowedClients = []string{
	"SSH-2.0-Syringe_0.0.4",
	"SSH-2.0-OpenSSH_9.9",
}

func ClientMiddleware(next ssh.Handler) ssh.Handler {
	return func(sess ssh.Session) {
		clientVersion := sess.Context().ClientVersion()
		if !slices.Contains(allowedClients, clientVersion) {
			log.Error(
				"disallowed client",
				"session", sess.Context().SessionID(),
				"version", clientVersion,
			)
			sess.Stderr().Write([]byte(serrors.New(
				"client", "unsupported client", sess.Context().SessionID(),
			).Error()))
			sess.Exit(1)
			return
		}

		next(sess)
	}
}
