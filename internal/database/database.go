package database

import (
	"database/sql"

	_ "github.com/tursodatabase/libsql-client-go/libsql"
)

type DbConfig struct {
	Location string
}

func Connection(url string) (*sql.DB, error) {
	db, err := sql.Open("libsql", url)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}

func CreateTables(db *sql.DB) error {
	// query := `
	// 	create table if not exists variables_ (
	// 		id_ integer primary key autoincrement not null,
	// 		key_ text not null,
	// 		value_ text not null,
	// 		secret_ boolean,
	// 		project_name_ text,
	// 		environment_name_ text,
	// 		unique (key_, project_name_, environment_name_)
	// 	)
	// `

	dropQuery := `drop table if exists users_`
	_, err := db.Exec(dropQuery)
	if err != nil {
		return err
	}

	query := `
		create table if not exists users_ (
			id_ integer primary key autoincrement not null,
			username_ text,
			email_ text,
			created_at_ datetime,
			password_ text
		)
	`

	_, err = db.Exec(query)
	if err != nil {
		return err
	}

	return nil
}
