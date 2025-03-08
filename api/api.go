package api

import "context"

type API interface {
	Set(key, value string) error
	Get(key string) (string, error)
	List() ([]string, error)
	Remove(key string) error
	Close() error
}

type hostAPI struct {
	// calls remote API over SSH
	url string
	ctx context.Context
}

func New(url string) *hostAPI {
	return &hostAPI{
		url: url,
	}
}

func (l *hostAPI) WithContext(ctx context.Context) *hostAPI {
	l.ctx = ctx
	return l
}

func (l *hostAPI) Set(key, value string) error {
	return nil
}

func (l *hostAPI) Get(key string) (string, error) {
	return "", nil
}

func (l *hostAPI) List() ([]string, error) {
	return nil, nil
}

func (l *hostAPI) Remove(key string) error {
	return nil
}

func (l *hostAPI) Close() error {
	return nil
}
