package main

import (
	"github.com/charmbracelet/log"
	"github.com/charmbracelet/ssh"
)

func NewLoggingMiddleware(logger *log.Logger) func(next ssh.Handler) ssh.Handler {
	return func(next ssh.Handler) ssh.Handler {
		return func(sess ssh.Session) {
			log.WithContext(sess.Context(), logger)
			// FIXME: this doesn't work as expected
			logger := log.With("session", sess.Context().SessionID())

			logger.Info(
				"connect",
				"user", sess.Context().User(),
				"address", sess.Context().RemoteAddr().String(),
				"public", sess.PublicKey() != nil,
				"client", sess.Context().ClientVersion(),
			)

			next(sess)

			logger.Info("disconnect")
		}
	}
}
