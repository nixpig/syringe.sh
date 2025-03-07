package items

type Item struct {
	ID    string
	Key   string
	Value string
}

func New(key, value string) *Item {
	return &Item{
		Key:   key,
		Value: value,
	}
}
