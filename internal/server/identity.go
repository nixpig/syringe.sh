package server

import (
	"github.com/charmbracelet/ssh"
)

// TODO: review whether this is even needed, given new solution design
func IdentityMiddleware(next ssh.Handler) ssh.Handler {
	return func(sess ssh.Session) {
		// sessionId := sess.Context().SessionID()
		//
		// publicKey := sess.PublicKey()
		// hashedPublicKey := sha1.Sum(publicKey.Marshal())
		// username := sess.Context().User()

		next(sess)
	}
}
