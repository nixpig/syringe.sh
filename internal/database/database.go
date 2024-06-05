package database

import (
	"database/sql"

	_ "github.com/tursodatabase/libsql-client-go/libsql"
)

type DbConfig struct {
	Location string
}

func Connection(databaseUrl, databaseToken string) (*sql.DB, error) {
	databaseConnectionString := databaseUrl + "?authToken=" + databaseToken

	db, err := sql.Open("libsql", databaseConnectionString)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}

func MigrateUserDb(db *sql.DB) error {
	var err error

	dropKeysTable := `drop table if exists keys_`
	_, err = db.Exec(dropKeysTable)
	if err != nil {
		return err
	}

	dropUsersTable := `drop table if exists users_`
	_, err = db.Exec(dropUsersTable)
	if err != nil {
		return err
	}

	createUsersTable := `
		create table if not exists users_ (
			id_ integer primary key autoincrement not null,
			username_ text,
			email_ text,
			created_at_ datetime without time zone default current_timestamp,
			status_ text
		)
	`

	createKeysTable := `
		create table if not exists keys_ (
			id_ integer primary key autoincrement not null,
			ssh_public_key_ text,
			user_id_ integer not null,
			created_at_ datetime without time zone default current_timestamp,

			foreign key (user_id_) references users_(id_)
		)
	`

	_, err = db.Exec(createUsersTable)
	if err != nil {
		return err
	}

	_, err = db.Exec(createKeysTable)
	if err != nil {
		return err
	}

	return nil
}
