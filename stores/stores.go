package stores

type Store interface {
	Set(item StoreItem) error
	Get(key string) (*StoreItem, error)
	List() ([]StoreItem, error)
	Delete(key string) error
}

type StoreItem struct {
	ID    int
	Key   string
	Value string
}
