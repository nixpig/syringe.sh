package ctxkeys

type ContextKey string

const (
	APP_DB             = ContextKey("DB_CTX")
	USER_DB            = ContextKey("USER_DB_CTX")
	ProjectService     = ContextKey("PROJECT_CTX")
	SecretService      = ContextKey("SECRET_CTX")
	EnvironmentService = ContextKey("ENVIRONMENT_CTX")
	UserService        = ContextKey("USER_CTX")
	Authenticated      = ContextKey("AUTHENTICATED_CTX")
	Username           = ContextKey("USERNAME_CTX")
	PublicKey          = ContextKey("PUBLIC_KEY_CTX")
)
