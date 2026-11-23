package config

import (
	"github.com/zachnoble/laurel/libraries/go/config"
)

// Config is auth's own configuration; the common variables (PORT,
// DATABASE_URL, keys for verifying tokens, ...) are read by libraries/go/service.
type Config struct {
	CookieSecret string

	// IdentityPrivateKey signs user identity tokens. Only auth holds it.
	IdentityPrivateKey string

	ResendAPIKey     string
	EmailFromAddress string
}

// New reads the config, reporting every missing variable at once.
func New(env *config.Env) (*Config, error) {
	cfg := &Config{
		CookieSecret:       env.Required("COOKIE_SECRET"),
		IdentityPrivateKey: env.Required("IDENTITY_PRIVATE_KEY"),
		ResendAPIKey:       env.Required("RESEND_API_KEY"),
		EmailFromAddress:   env.Required("EMAIL_FROM_ADDRESS"),
	}
	return cfg, env.Err()
}
