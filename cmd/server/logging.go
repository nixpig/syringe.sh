package main

import (
	"github.com/charmbracelet/log"
	"github.com/charmbracelet/ssh"
)

func loggingMiddleware(next ssh.Handler) ssh.Handler {
	return func(sess ssh.Session) {
		log.Info(
			"connect",
			"user", sess.Context().User(),
			"address", sess.Context().RemoteAddr().String(),
			"public", sess.PublicKey() != nil,
			"client", sess.Context().ClientVersion(),
		)

		next(sess)

		log.Info("disconnect")
	}
}
