package database

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

func NewConnection(
	filename,
	user,
	password string,
) (*sql.DB, error) {
	databaseConnectionString := fmt.Sprintf(
		"file:%s?_auth&_auth_user=%s&_auth_pass=%s&_auth_crypt=sha1",
		filename,
		user,
		password,
	)

	db, err := sql.Open("sqlite3", databaseConnectionString)
	if err != nil {
		return nil, fmt.Errorf("failed to open database file (%s): %w", databaseConnectionString, err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database (%s): %w", databaseConnectionString, err)
	}

	return db, nil
}
