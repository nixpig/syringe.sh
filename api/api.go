package api

import (
	"github.com/charmbracelet/log"
	"github.com/nixpig/syringe.sh/internal/items"
)

type API interface {
	Set(*items.Item) error
	Get(key string) (*items.Item, error)
	List() ([]items.Item, error)
	Delete(key string) error
	Close() error
}

func New(store string) (API, error) {
	// if _, err := url.ParseRequestURI(store); err != nil {
	log.Debug("new store", "store", store)
	return newFileAPI(store)
	// }

	// log.Debug("host api")
	// return newHostAPI(store)
}
