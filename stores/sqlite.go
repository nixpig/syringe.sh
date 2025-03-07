package stores

import (
	"database/sql"
	"fmt"

	"github.com/nixpig/syringe.sh/internal/items"
)

type SqliteStore struct {
	db *sql.DB
}

func NewSqliteStore(db *sql.DB) *SqliteStore {
	return &SqliteStore{db}
}

func (s *SqliteStore) Set(item *items.Item) error {
	query := `insert into store_ (key_, value_) values ($key, $value) 
on conflict(key_) do update set value_ = $value`

	if _, err := s.db.Exec(
		query,
		sql.Named("key", item.Key),
		sql.Named("value", item.Value),
	); err != nil {
		return fmt.Errorf("insert key-value in database: %w", err)
	}

	return nil
}

func (s *SqliteStore) Get(key string) (*items.Item, error) {
	query := `select id_, key_, value_ from store_
where key_ = $key`

	row := s.db.QueryRow(query, sql.Named("key", key))

	var item items.Item

	if err := row.Scan(&item.ID, &item.Key, &item.Value); err != nil {
		return nil, fmt.Errorf("get key-value from database: %w", err)
	}

	return &item, nil
}

func (s *SqliteStore) List() ([]items.Item, error) {
	query := `select id_, key_, value_ from store_`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("get all key-values from database: %w", err)
	}
	defer rows.Close()

	var allItems []items.Item

	for rows.Next() {
		var item items.Item

		if err := rows.Scan(&item.ID, &item.Key, &item.Value); err != nil {
			return nil, fmt.Errorf("scan row item: %w", err)
		}

		allItems = append(allItems, item)
	}

	return allItems, nil
}

func (s *SqliteStore) Delete(key string) error {
	query := `delete from store_ where key_ = $key`

	if _, err := s.db.Exec(query, sql.Named("key", key)); err != nil {
		return fmt.Errorf("delete item: %w", err)
	}

	return nil
}
