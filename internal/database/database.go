package database

import "database/sql"

type DbConfig struct {
	Location string
}

func Connect(config DbConfig) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", config.Location)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}

func Create(config DbConfig) error {
	db, err := sql.Open("sqlite3", config.Location)
	if err != nil {
		return err
	}

	if err := db.Ping(); err != nil {
		return err
	}

	query := `
		create table if not exists variables_ (
			id_ integer primary key generated not null, 
			key_ text not null, 
			value_ text not null
		)
	`

	_, err = db.Exec(query)
	if err != nil {
		return err
	}

	return nil
}
