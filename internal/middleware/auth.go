package middleware

import (
	"fmt"

	"github.com/charmbracelet/ssh"
	"github.com/nixpig/syringe.sh/config"
	"github.com/nixpig/syringe.sh/internal/auth"
	"github.com/nixpig/syringe.sh/pkg/ctxkeys"
	"github.com/rs/zerolog"
)

func NewMiddlewareAuth(
	logger *zerolog.Logger,
	authService auth.AuthService,
) func(next ssh.Handler) ssh.Handler {
	return func(next ssh.Handler) ssh.Handler {
		return func(sess ssh.Session) {
			clientVersion := sess.Context().ClientVersion()

			if clientVersion != config.Client {
				logger.Error().Str("clientVersion", clientVersion).Msg("invalid client")
				sess.Stderr().Write([]byte(fmt.Sprintf("Error: unsupported client %s.\nPlease use the syringe CLI, available at: \n  https://github.com/nixpig/syringe.sh\n", clientVersion)))
				return
			}

			user, err := authService.AuthenticateUser(auth.AuthenticateUserRequest{
				Username:  sess.User(),
				PublicKey: sess.PublicKey(),
			})
			if err != nil {
				logger.Error().Err(err).Msg("failed to authenticate user")

				sess.Stderr().Write([]byte("Error: failed to authenticate.\n"))

				return
			}

			sess.Context().SetValue(ctxkeys.Authenticated, user.Auth)

			next(sess)
		}
	}
}
