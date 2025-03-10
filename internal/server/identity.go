package server

import (
	"crypto/sha1"
	"fmt"
	"time"

	"github.com/charmbracelet/log"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	"github.com/nixpig/syringe.sh/internal/serrors"
	"github.com/nixpig/syringe.sh/internal/stores"
)

// TODO: review whether this is even needed, given new solution design
func NewIdentityMiddleware(s *stores.SystemStore) wish.Middleware {
	return func(next ssh.Handler) ssh.Handler {
		return func(sess ssh.Session) {
			sessionID := sess.Context().SessionID()
			username := sess.Context().User()
			publicKeyHash := fmt.Sprintf("%x", sha1.Sum(sess.PublicKey().Marshal()))
			email := time.Now().GoString()

			user, err := s.GetUser(username, publicKeyHash)
			if err != nil || user == nil {
				sess.Write([]byte(fmt.Sprintf("User '%s' not found.\n", username)))

				log.Debug(
					"creating user",
					"session", sessionID,
					"username", username,
					"email", email,
					"publicKeyHash", publicKeyHash,
					"getUserErr", err,
				)

				user = &stores.User{
					Username:      username,
					PublicKeySHA1: publicKeyHash,
					Email:         email,
				}
				userID, err := s.CreateUser(user)
				if err != nil {
					log.Error(
						"failed to create user",
						"session", sessionID,
						"err", err,
					)

					sess.Stderr().Write([]byte(serrors.New("user", fmt.Sprintf("failed to create user '%s'", username), sessionID).Error()))
					sess.Exit(1)
					return
				}

				sess.Write([]byte(fmt.Sprintf("Created user '%s'.\n", username)))

				user.ID = userID
			}

			// TODO: add email verification

			next(sess)
		}
	}
}
