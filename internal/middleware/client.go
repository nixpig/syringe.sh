package middleware

import (
	"slices"

	"github.com/charmbracelet/log"
	"github.com/charmbracelet/ssh"
)

// TODO: read allowed clients from config or database?
var allowedClients = []string{
	"SSH-2.0-Syringe",
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
			sess.Stderr().Write([]byte("unsupported client"))
			sess.Exit(1)
			return
		}

		next(sess)
	}
}
