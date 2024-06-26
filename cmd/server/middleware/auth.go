package middleware

import (
	"github.com/charmbracelet/ssh"
	"github.com/nixpig/syringe.sh/internal/auth"
	"github.com/nixpig/syringe.sh/pkg/ctxkeys"
	"github.com/rs/zerolog"
)

func NewAuthHandler(
	logger *zerolog.Logger,
	authService auth.AuthService,
) func(next ssh.Handler) ssh.Handler {
	return func(next ssh.Handler) ssh.Handler {
		return func(sess ssh.Session) {
			user, err := authService.AuthenticateUser(auth.AuthenticateUserRequest{
				Username:  sess.User(),
				PublicKey: sess.PublicKey(),
			})
			if err != nil {
				logger.Warn().Msg("user not authenticated")

				sess.Write([]byte("Public key not recognised.\n"))

				return
			}

			sess.Context().SetValue(ctxkeys.Authenticated, user.Auth)

			next(sess)
		}
	}
}
