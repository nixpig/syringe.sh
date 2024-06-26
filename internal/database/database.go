package database

import (
	"crypto/sha1"
	"database/sql"
	"fmt"
	"net/http"
	"os"

	"github.com/charmbracelet/ssh"
	"github.com/nixpig/syringe.sh/pkg/serrors"
	"github.com/nixpig/syringe.sh/pkg/turso"
	_ "github.com/tursodatabase/libsql-client-go/libsql"
	gossh "golang.org/x/crypto/ssh"
)

type DBConfig struct {
	Location string
}

func Connection(databaseURL, databaseToken string) (*sql.DB, error) {
	databaseConnectionString := databaseURL + "?authToken=" + databaseToken

	db, err := sql.Open("libsql", databaseConnectionString)
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
	api := turso.New(
		os.Getenv("DATABASE_ORG"),
		os.Getenv("API_TOKEN"),
		http.Client{},
	)

	marshalledKey := gossh.MarshalAuthorizedKey(publicKey)

	hashedKey := fmt.Sprintf("%x", sha1.Sum(marshalledKey))
	expiration := "30s"

	token, err := api.CreateToken(hashedKey, expiration)
	if err != nil {
		return nil, fmt.Errorf("failed to create token:\n%s", err)
	}

	db, err := Connection(
		"libsql://"+hashedKey+"-"+os.Getenv("DATABASE_ORG")+".turso.io",
		string(token.Jwt),
	)
	if err != nil {
		return nil, fmt.Errorf("error creating database connection:\n%s", err)
	}

	return db, nil
}
