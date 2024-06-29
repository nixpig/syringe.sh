package test

import (
	"testing"
)

const (
	host = "localhost"
	port = "23234"
)

func TestEndToEnd(t *testing.T) {
	scenarios := map[string]func(t *testing.T){
		"test register with new key": testRegisterWithNewKey,
	}

	for scenario, fn := range scenarios {
		t.Run(scenario, func(t *testing.T) {
			// sshServer := server.NewServer(appService, &log)

			// if err := sshServer.Start(
			// 	HOST, PORT,
			// ); err != nil {
			// 	t.Fatal("failed to start server")
			// }

			// todo: how do we stop it?

			fn(t)
		})
	}
}

func testRegisterWithNewKey(t *testing.T) {
}
