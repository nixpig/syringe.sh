package api

import "github.com/nixpig/syringe.sh/stores"

/*
	Calls server API over SSH
*/

type hostAPI struct {
	// calls remote API over SSH
}

func newHostAPI(url string) (*hostAPI, error) {
	return &hostAPI{}, nil
}

func (l *hostAPI) Set(item stores.StoreItem) error {
	return nil
}

func (l *hostAPI) Get(key string) (*stores.StoreItem, error) {
	return nil, nil
}

func (l *hostAPI) List() ([]stores.StoreItem, error) {
	return nil, nil
}

func (l *hostAPI) Delete(key string) error {
	return nil
}
