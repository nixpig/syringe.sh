package user_test

import (
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestUserCmd(t *testing.T) {
	// scenarios := map[string]func(
	// 	t *testing.T,
	// 	cmd *cobra.Command,
	// 	service project.ProjectService,
	// 	mock sqlmock.Sqlmock,
	// ){
	// "test user command initialise cmd": testUserCmdInitialiseCmd,
}

// for scenario, fn := range scenarios {
// 	t.Run(scenario, func(t *testing.T) {
// 		db, mock, err := sqlmock.New()
// 		if err != nil {
// 			t.Fatalf("unable to create mock database:\n%s", err)
// 		}
//
// 		// cmd := user.NewCmdUser()
//
// 		// fn(t, mock, db)
// 	})
// }
//
// }

func testUserCmdInitialiseCmd(t *testing.T, mock sqlmock.Sqlmock, db *sql.DB) {
	// cmdIn := bytes.NewReader([]byte{})
	// cmdOut := bytes.NewBufferString("")
	//
	// userCmd := user.New(user.InitContext)
	// userCmd.AddCommand(user.RegisterCmd(func(cmd *cobra.Command, args []string) error {}))
	//
	// err := userCmd.Execute(
	// 	[]*cobra.Command{(nil)},
	// 	[]string{"user"},
	// 	cmdIn,
	// 	cmdOut,
	// 	os.Stderr,
	// 	db,
	// )
	//
	// require.NoError(t, err)
}
