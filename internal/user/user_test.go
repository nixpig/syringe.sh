package user_test

import (
	"bytes"
	"database/sql"
	"os"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/nixpig/syringe.sh/cmd/server/servercmd"
	"github.com/nixpig/syringe.sh/internal/user"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

func TestUserCmd(t *testing.T) {
	scenarios := map[string]func(t *testing.T, mock sqlmock.Sqlmock, db *sql.DB){
		"test user command initialise cmd": testUserCmdInitialiseCmd,
	}

	for scenario, fn := range scenarios {
		t.Run(scenario, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("unable to create mock database:\n%s", err)
			}

			fn(t, mock, db)
		})
	}

}

func testUserCmdInitialiseCmd(t *testing.T, mock sqlmock.Sqlmock, db *sql.DB) {
	cmdIn := bytes.NewReader([]byte{})
	cmdOut := bytes.NewBufferString("")

	err := servercmd.Execute(
		[]*cobra.Command{user.UserCommand(nil)},
		[]string{"user"},
		cmdIn,
		cmdOut,
		os.Stderr,
		db,
	)

	require.NoError(t, err)
}
