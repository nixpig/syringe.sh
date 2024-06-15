package pkg

type ContextKey string

const (
	DBCtxKey          = ContextKey("DB_CTX")
	ProjectCtxKey     = ContextKey("PROJECT_CTX")
	SecretCtxKey      = ContextKey("SECRET_CTX")
	EnvironmentCtxKey = ContextKey("ENVIRONMENT_CTX")
)
