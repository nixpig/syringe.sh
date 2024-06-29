package ctxkeys

type ContextKey string

const (
	Authenticated = ContextKey("AUTHENTICATED_CTX")
	Username      = ContextKey("USERNAME_CTX")
	PublicKey     = ContextKey("PUBLIC_KEY_CTX")
)
