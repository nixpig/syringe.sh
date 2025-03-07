package api

import (
	"github.com/nixpig/syringe.sh/internal/items"
)

/*
	Calls server API over SSH
*/

type hostAPI struct {
	// calls remote API over SSH
}

func newHostAPI(url string) (*hostAPI, error) {
	return &hostAPI{}, nil
}

func (l *hostAPI) Set(item *items.Item) error {
	return nil
}

func (l *hostAPI) Get(key string) (*items.Item, error) {
	return nil, nil
}

func (l *hostAPI) List() ([]items.Item, error) {
	return nil, nil
}

func (l *hostAPI) Remove(key string) error {
	return nil
}
