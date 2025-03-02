package store

import (
	"database/sql"
	"fmt"
)

type Store interface {
	Set(item StoreItem) error
	Get(key string) (*StoreItem, error)
	List() ([]StoreItem, error)
	Delete(key string) error
}

type StoreItem struct {
	Id    int
	Key   string
	Value string
}

type SqliteStore struct {
	db *sql.DB
}

func NewSqliteStore(db *sql.DB) *SqliteStore {
	return &SqliteStore{db}
}

func (s *SqliteStore) Set(item StoreItem) error {
	query := `
		insert into store_ (key_, value_) values ($key, $value) 
		on conflict(key_) do update set value_ = $value
	`

	if _, err := s.db.Exec(
		query,
		sql.Named("key", item.Key),
		sql.Named("value", item.Value),
	); err != nil {
		return fmt.Errorf("insert key-value in database: %w", err)
	}

	return nil
}

func (s *SqliteStore) Get(key string) (*StoreItem, error) {
	query := `
		select id_, key_, value_ from store_
		where key_ = $key
	`

	row := s.db.QueryRow(query, sql.Named("key", key))

	var item StoreItem

	if err := row.Scan(&item.Id, &item.Key, &item.Value); err != nil {
		return nil, fmt.Errorf("get key-value from database: %w", err)
	}

	return &item, nil
}

func (s *SqliteStore) List() ([]StoreItem, error) {
	query := `
		select id_, key_, value_ from store_
	`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("get all key-values from database: %w", err)
	}
	defer rows.Close()

	var items []StoreItem

	for rows.Next() {
		var item StoreItem

		if err := rows.Scan(&item.Id, &item.Key, &item.Value); err != nil {
			return nil, fmt.Errorf("scan row item: %w", err)
		}

		items = append(items, item)
	}

	return items, nil
}

func (s *SqliteStore) Delete(key string) error {
	query := `
		delete from store_ where key_ = $key
	`

	if _, err := s.db.Exec(query, sql.Named("key", key)); err != nil {
		return fmt.Errorf("delete item: %w", err)
	}

	return nil
}
