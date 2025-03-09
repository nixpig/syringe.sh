package main

import (
	"database/sql"
	"fmt"
	"sync"
)

type Store struct {
	db *sql.DB
	mu sync.Mutex
}

type Item struct {
	ID    string
	Key   string
	Value string
}

func NewStore(db *sql.DB) *Store {
	return &Store{
		db: db,
		mu: sync.Mutex{},
	}
}

func (s *Store) Set(item *Item) error {
	s.mu.Lock()
	defer s.mu.Unlock()

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

func (s *Store) Get(key string) (*Item, error) {
	query := `select id_, key_, value_ from store_
where key_ = $key`

	row := s.db.QueryRow(query, sql.Named("key", key))

	var item Item

	if err := row.Scan(&item.ID, &item.Key, &item.Value); err != nil {
		return nil, fmt.Errorf("get key-value from database: %w", err)
	}

	return &item, nil
}

func (s *Store) List() ([]Item, error) {
	query := `select id_, key_, value_ from store_`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("get all key-values from database: %w", err)
	}
	defer rows.Close()

	var allItems []Item

	for rows.Next() {
		var item Item

		if err := rows.Scan(&item.ID, &item.Key, &item.Value); err != nil {
			return nil, fmt.Errorf("scan row item: %w", err)
		}

		allItems = append(allItems, item)
	}

	return allItems, nil
}

func (s *Store) Remove(key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	query := `delete from store_ where key_ = $key`

	if _, err := s.db.Exec(query, sql.Named("key", key)); err != nil {
		return fmt.Errorf("delete item: %w", err)
	}

	return nil
}
