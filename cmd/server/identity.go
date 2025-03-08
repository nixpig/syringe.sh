package main

import (
	"bufio"
	"fmt"
	"net/http"

	"github.com/charmbracelet/log"
	"github.com/charmbracelet/ssh"
)

// TODO: review whether this is even needed, given new solution design
func identityMiddleware(next ssh.Handler) ssh.Handler {
	return func(sess ssh.Session) {
		username := sess.Context().User()
		publicKey := sess.PublicKey()

		publicKeysURL := fmt.Sprintf("https://github.com/%s.keys", username)

		resp, err := http.Get(publicKeysURL)
		if err != nil || resp.StatusCode != http.StatusOK {
			log.Warn("failed to get public keys", "publicKeysURL", publicKeysURL)
			sess.Stderr().Write([]byte(fmt.Sprintf("Error: failed to get public keys from %s\n", publicKeysURL)))
			return
		}
		defer resp.Body.Close()

		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			k := scanner.Text()
			authorisedKey, _, _, _, err := ssh.ParseAuthorizedKey([]byte(k))
			if err != nil {
				log.Debug("failed to parse authorised key", "key", k, "err", err)
				continue
			}

			if ssh.KeysEqual(publicKey, authorisedKey) {
				next(sess)
				return
			}
		}

		if err := scanner.Err(); err != nil {
			log.Error("failed to read keys response body", "err", err)
			sess.Stderr().Write([]byte(fmt.Sprintf("Error: failed to read keys\n")))
		}

		sess.Stderr().Write([]byte("Error: no matching keys found\n"))
		sess.Exit(1)
		return
	}
}
