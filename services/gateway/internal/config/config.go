package config

import (
	"time"

	"github.com/zachnoble/laurel/libraries/go/config"
)

// Config is the gateway's own configuration; UPSTREAM_URL and
// SERVICE_URL_<NAME> (where to route each service) are read by
// libraries/go/service.
type Config struct {
	// AllowedOrigins for CORS. Entries may use one wildcard, e.g.
	// https://*.example.com. Apps must be same-site with the gateway (see
	// README, "Apps"), since the session cookie is SameSite=Lax.
	AllowedOrigins []string

	// SessionService owns the session cookie: it is the only upstream that
	// receives the Cookie header, and it mints identity tokens.
	SessionService string
	// SessionCookie is the session cookie's name.
	SessionCookie string
	// SessionCacheSize bounds the session cache (entries).
	SessionCacheSize int
	// SessionCacheTTL is how long a revoked session can keep working.
	SessionCacheTTL time.Duration
}

func New(env *config.Env) *Config {
	return &Config{
		AllowedOrigins:   env.List("ALLOWED_ORIGINS"),
		SessionService:   env.String("SESSION_SERVICE", "auth"),
		SessionCookie:    env.String("SESSION_COOKIE", "session_token"),
		SessionCacheSize: env.Int("SESSION_CACHE_SIZE", 100_000),
		SessionCacheTTL:  env.Duration("SESSION_CACHE_TTL", 30*time.Second),
	}
}
