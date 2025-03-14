package stores

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
)

type TenantStore struct {
	ctx context.Context
	db  *sql.DB
	mu  sync.Mutex
}

func NewTenantStore(ctx context.Context, db *sql.DB) *TenantStore {
	return &TenantStore{
		ctx: ctx,
		db:  db,
		mu:  sync.Mutex{},
	}
}

func (s *TenantStore) Context() context.Context {
	return s.ctx
}

func (s *TenantStore) SetItem(item *Item) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	query := `insert into store_ (key_, value_) values ($key, $value) 
on conflict(key_) do update set value_ = $value`

	if _, err := s.db.ExecContext(
		s.ctx,
		query,
		sql.Named("key", item.Key),
		sql.Named("value", item.Value),
	); err != nil {
		return fmt.Errorf("insert key-value in database: %w", err)
	}

	return nil
}

func (s *TenantStore) GetItemByKey(key string) (*Item, error) {
	query := `select id_, key_, value_ from store_
where key_ = $key`

	row := s.db.QueryRowContext(s.ctx, query, sql.Named("key", key))

	var item Item

	if err := row.Scan(&item.ID, &item.Key, &item.Value); err != nil {
		return nil, fmt.Errorf("get key-value from database: %w", err)
	}

	return &item, nil
}

func (s *TenantStore) ListItems() ([]Item, error) {
	query := `select id_, key_, value_ from store_`

	rows, err := s.db.QueryContext(s.ctx, query)
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

func (s *TenantStore) RemoveItemByKey(key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	query := `delete from store_ where key_ = $key`

	if _, err := s.db.ExecContext(
		s.ctx, query, sql.Named("key", key),
	); err != nil {
		return fmt.Errorf("delete item: %w", err)
	}

	return nil
}
