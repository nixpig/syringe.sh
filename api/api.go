package api

import (
	"github.com/charmbracelet/log"
	"github.com/nixpig/syringe.sh/stores"
)

type API interface {
	Set(item stores.StoreItem) error
	Get(key string) (*stores.StoreItem, error)
	List() ([]stores.StoreItem, error)
	Delete(key string) error
}

func New(store string) (API, error) {
	// if _, err := url.ParseRequestURI(store); err != nil {
	log.Debug("file api")
	return newFileAPI(store)
	// }

	// log.Debug("host api")
	// return newHostAPI(store)
}
