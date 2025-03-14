package stores

// TODO: validate item and user structs before making database queries!

type Item struct {
	ID    int
	Key   string
	Value string
}

type User struct {
	ID            int
	Username      string
	Email         string
	Verified      bool
	PublicKeySHA1 string
}
