package ctxkeys

type ContextKey string

const (
	DB                 = ContextKey("DB_CTX")
	ProjectService     = ContextKey("PROJECT_CTX")
	SecretService      = ContextKey("SECRET_CTX")
	EnvironmentService = ContextKey("ENVIRONMENT_CTX")
	UserService        = ContextKey("USER_CTX")
	Authenticated      = ContextKey("AUTHENTICATED_CTX")
)
