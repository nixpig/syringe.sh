package stores

import "github.com/nixpig/syringe.sh/internal/items"

type Store interface {
	Set(item *items.Item) error
	Get(key string) (*items.Item, error)
	List() ([]items.Item, error)
	Remove(key string) error
}
