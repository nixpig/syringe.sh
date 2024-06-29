package database_test

// import (
// 	"database/sql"
// 	"regexp"
// 	"testing"
//
// 	"github.com/DATA-DOG/go-sqlmock"
// 	"github.com/nixpig/syringe.sh/internal/database"
// 	"github.com/stretchr/testify/require"
// )
//
// func TestDatabase(t *testing.T) {
// 	scenarios := map[string]func(t *testing.T, db *sql.DB, mock sqlmock.Sqlmock){
// 		"test migrate app database happy path": testMigrateAppDBHappyPath,
// 	}
//
// 	for scenario, fn := range scenarios {
// 		t.Run(scenario, func(t *testing.T) {
// 			db, mock, err := sqlmock.New()
// 			if err != nil {
// 				t.Fatal("unable to create mock database")
// 			}
//
// 			fn(t, db, mock)
// 		})
// 	}
// }
//
// func testMigrateAppDBHappyPath(t *testing.T, db *sql.DB, mock sqlmock.Sqlmock) {
// 	dropKeysTable := `drop table if exists keys_`
// 	dropUsersTable := `drop table if exists users_`
//
// 	createUsersTable := `
// 		create table if not exists users_ (
// 			id_ integer primary key autoincrement,
// 			username_ varchar(256) not null,
// 			email_ varchar(256) not null,
// 			created_at_ datetime without time zone default current_timestamp,
// 			status_ varchar(8) not null
// 		)
// 	`
//
// 	createKeysTable := `
// 		create table if not exists keys_ (
// 			id_ integer primary key autoincrement,
// 			ssh_public_key_ varchar(1024) not null,
// 			user_id_ integer not null,
// 			created_at_ datetime without time zone default current_timestamp,
//
// 			foreign key (user_id_) references users_(id_)
// 		)
// 	`
//
// 	mock.ExpectExec(regexp.QuoteMeta(dropKeysTable)).WillReturnResult(sqlmock.NewResult(0, 0))
// 	mock.ExpectExec(regexp.QuoteMeta(dropUsersTable)).WillReturnResult(sqlmock.NewResult(0, 0))
// 	mock.ExpectExec(regexp.QuoteMeta(createUsersTable)).WillReturnResult(sqlmock.NewResult(0, 0))
// 	mock.ExpectExec(regexp.QuoteMeta(createKeysTable)).WillReturnResult(sqlmock.NewResult(0, 0))
//
// 	err := database.MigrateAppDB(db)
//
// 	require.NoError(t, err)
// 	require.NoError(t, mock.ExpectationsWereMet())
// }
