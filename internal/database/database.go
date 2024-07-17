package database

import (
	"crypto/sha1"
	"database/sql"
	"fmt"
	"os"

	"github.com/charmbracelet/ssh"
	_ "github.com/mattn/go-sqlite3"
	"github.com/nixpig/syringe.sh/pkg/serrors"
	gossh "golang.org/x/crypto/ssh"
)

type DBConfig struct {
	Location string
}

func Connection(
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
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}

func MigrateAppDB(db *sql.DB) error {
	dropKeysTable := `drop table if exists keys_`
	if _, err := db.Exec(dropKeysTable); err != nil {
		return serrors.ErrDatabaseExec(err)
	}

	dropUsersTable := `drop table if exists users_`
	if _, err := db.Exec(dropUsersTable); err != nil {
		return serrors.ErrDatabaseExec(err)
	}

	createUsersTable := `
		create table if not exists users_ (
			id_ integer primary key autoincrement,
			username_ varchar(256) not null,
			email_ varchar(256) not null,
			created_at_ datetime without time zone default current_timestamp,
			status_ varchar(8) not null
		)
	`

	createKeysTable := `
		create table if not exists keys_ (
			id_ integer primary key autoincrement,
			ssh_public_key_ varchar(1024) not null,
			user_id_ integer not null,
			created_at_ datetime without time zone default current_timestamp,

			foreign key (user_id_) references users_(id_)
		)
	`

	if _, err := db.Exec(createUsersTable); err != nil {
		return serrors.ErrDatabaseExec(err)
	}

	if _, err := db.Exec(createKeysTable); err != nil {
		return serrors.ErrDatabaseExec(err)
	}

	return nil
}

func NewUserDBConnection(publicKey ssh.PublicKey) (*sql.DB, error) {
	marshalledKey := gossh.MarshalAuthorizedKey(publicKey)

	hashedKey := fmt.Sprintf("%x", sha1.Sum(marshalledKey))
	db, err := Connection(
		fmt.Sprintf("%s.db", hashedKey),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
	)
	if err != nil {
		return nil, fmt.Errorf("error creating database connection:\n%s", err)
	}

	return db, nil
}
