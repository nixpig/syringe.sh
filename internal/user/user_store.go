package user

import (
	"database/sql"
	"errors"
)

type UserStore interface {
	Insert(username, email, status string) (*User, error)
	GetByUsername(username string) (*User, error)
	DeleteByUsername(username string) error
	Update(user User) (*User, error)
}

type SqliteUserStore struct {
	db *sql.DB
}

func NewSqliteUserStore(db *sql.DB) SqliteUserStore {
	return SqliteUserStore{db}
}

func (u SqliteUserStore) Insert(username, email, status string) (*User, error) {
	query := `
		insert into users_ (username_, email_, status_) 
		values ($username, $email, $status) 
		returning id_, username_, email_, status_
	`

	row := u.db.QueryRow(
		query,
		sql.Named("username", username),
		sql.Named("email", email),
		sql.Named("status", status),
		// sql.Named("createdAt", time.Now().UTC()),
	)

	var insertedUser User

	if err := row.Scan(
		&insertedUser.Id,
		&insertedUser.Username,
		&insertedUser.Email,
		&insertedUser.Status,
		// &insertedUser.CreatedAt,
	); err != nil {
		return nil, err
	}

	return &insertedUser, nil
}

func (u SqliteUserStore) GetByUsername(username string) (*User, error) {
	query := `
		select id_, username_, email_, status_, created_at_ 
		from users_ 
		where username_ = $1
	`

	row := u.db.QueryRow(query, username)

	var user User

	if err := row.Scan(
		&user.Id,
		&user.Username,
		&user.Email,
		&user.Status,
		&user.CreatedAt,
	); err != nil {
		return nil, err
	}

	return &user, nil
}

func (u SqliteUserStore) DeleteByUsername(username string) error {
	query := `delete from users_ where username_ = $1`

	res, err := u.db.Exec(query, username)
	if err != nil {
		return err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("no user deleted")
	}

	return nil
}

func (u SqliteUserStore) Update(user User) (*User, error) {
	query := `
		update users_ set email_ = $2, set status_ = $3
		where username_ = $1 
		returning id_, username_, email_, status_, created_at_
	`

	row := u.db.QueryRow(query, user.Username, user.Email, user.Status)

	var updatedUser User

	if err := row.Scan(
		&updatedUser.Id,
		&updatedUser.Username,
		&updatedUser.Email,
		&updatedUser.Status,
		&updatedUser.CreatedAt,
	); err != nil {
		return nil, err
	}

	return &user, nil
}
