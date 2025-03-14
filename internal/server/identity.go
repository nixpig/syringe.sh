package server

import (
	"crypto/sha1"
	"fmt"
	"net/mail"

	"github.com/charmbracelet/log"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	"github.com/nixpig/syringe.sh/internal/stores"
)

func NewIdentityMiddleware(s *stores.SystemStore) wish.Middleware {
	return func(next ssh.Handler) ssh.Handler {
		return func(sess ssh.Session) {
			publicKeyHash := fmt.Sprintf("%x", sha1.Sum(sess.PublicKey().Marshal()))
			sess.Context().SetValue("publicKeyHash", publicKeyHash)

			email := "nixpig@example.org"
			if _, err := mail.ParseAddress(email); err != nil {
				sess.Stderr().Write([]byte("Error: invalid email address"))
				sess.Exit(1)
				return
			}
			sess.Context().SetValue("email", email)

			sessionID := sess.Context().SessionID()
			username := sess.Context().User()

			cmd := sess.Command()
			if len(cmd) > 0 && cmd[0] == "register" {
				log.Debug("register user", "session", sessionID, "username", username, "email", email, "publicKeyHash", publicKeyHash)
				next(sess)
			}

			authenticated := false
			user, err := s.GetUser(username)
			if err == nil && user != nil && user.PublicKeySHA1 == publicKeyHash {
				authenticated = true
			}

			log.Debug("is user authenticated", "authenticated", authenticated)

			sess.Context().SetValue("authenticated", authenticated)

			// TODO: add email verification

			next(sess)
		}
	}
}
