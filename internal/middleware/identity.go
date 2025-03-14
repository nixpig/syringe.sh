package middleware

import (
	"crypto/sha1"
	"fmt"
	"net/mail"

	"github.com/charmbracelet/log"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	"github.com/nixpig/syringe.sh/internal/stores"
)

var contextKeyHash = struct{ string }{"publicKeyHash"}
var contextKeyEmail = struct{ string }{"email"}
var contextKeyAuthenticated = struct{ string }{"authenticated"}

func NewIdentityMiddleware(s *stores.SystemStore) wish.Middleware {
	return func(next ssh.Handler) ssh.Handler {
		return func(sess ssh.Session) {
			publicKeyHash := fmt.Sprintf("%x", sha1.Sum(sess.PublicKey().Marshal()))
			sess.Context().SetValue(contextKeyHash, publicKeyHash)

			// TODO: pull email from key? reject keys without email?
			email := "nixpig@example.org"
			if _, err := mail.ParseAddress(email); err != nil {
				sess.Stderr().Write([]byte("invalid email address"))
				sess.Exit(1)
				return
			}
			sess.Context().SetValue(contextKeyEmail, email)

			authenticated := false
			user, err := s.GetUser(sess.Context().User())
			if err == nil && user != nil && user.PublicKeySHA1 == publicKeyHash {
				authenticated = true
			}
			sess.Context().SetValue(contextKeyAuthenticated, authenticated)

			log.Debug("authenticate", "authenticated", authenticated)

			next(sess)
		}
	}
}
