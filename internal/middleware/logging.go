package middleware

import (
	"time"

	"github.com/charmbracelet/log"
	"github.com/charmbracelet/ssh"
)

func LoggingMiddleware(next ssh.Handler) ssh.Handler {
	return func(sess ssh.Session) {
		command := ""
		if len(sess.Command()) > 0 {
			command = sess.Command()[0]
		}

		log.Info(
			"connect",
			"session", sess.Context().SessionID(),
			"command", command,
			"user", sess.Context().User(),
			"address", sess.Context().RemoteAddr().String(),
			"public", sess.PublicKey() != nil,
			"client", sess.Context().ClientVersion(),
			"publicKeyType", sess.PublicKey().Type(),
		)

		now := time.Now()

		next(sess)

		log.Info(
			"disconnect",
			"session", sess.Context().SessionID(),
			"elapsed", time.Since(now),
		)
	}
}
