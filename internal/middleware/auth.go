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
				logger.Warn().Str("clientVersion", clientVersion).Msg("invalid client")
				sess.Stderr().Write([]byte(fmt.Sprintf("Unsupported client %s.\nPlease use the syringe CLI, available at: \n  https://github.com/nixpig/syringe.sh\n", clientVersion)))
				return
			}
			user, err := authService.AuthenticateUser(auth.AuthenticateUserRequest{
				Username:  sess.User(),
				PublicKey: sess.PublicKey(),
			})
			if err != nil {
				logger.Warn().Msg("user not authenticated")

				sess.Stderr().Write([]byte("Public key not recognised.\n"))

				return
			}

			sess.Context().SetValue(ctxkeys.Authenticated, user.Auth)

			next(sess)
		}
	}
}
