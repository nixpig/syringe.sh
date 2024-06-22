package server

import (
	"github.com/charmbracelet/ssh"
	"github.com/nixpig/syringe.sh/server/internal/services"
)

func authAndRegisterHandler(s Server) func(next ssh.Handler) ssh.Handler {
	return func(next ssh.Handler) ssh.Handler {
		return func(sess ssh.Session) {
			if user, err := s.app.AuthenticateUser(services.UserAuthRequest{
				Username:  sess.User(),
				PublicKey: sess.PublicKey(),
			}); err != nil || !user.Auth {
				s.logger.Warn().Msg("user not logged in")
				s.logger.Warn().Msg("prompt to register and call 'register' command if answer is 'Y', else return/exit")
				s.logger.Warn().Msg("auto-registering for now...")

				sess.Write([]byte("You are not logged in or registered.\n"))
				sess.Write([]byte("Auto-registering using provided public key...\n"))

				_, err := s.app.RegisterUser(services.RegisterUserRequest{
					Username:  sess.User(),
					Email:     "not_used_yet@example.org",
					PublicKey: sess.PublicKey(),
				})
				if err != nil {
					s.logger.Err(err).Msg("failed to register user")
					sess.Stderr().Write([]byte("Failed to register user using provided public key.\n"))
					return
				}
			}

			next(sess)
		}
	}
}
