package db

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Config configures the connection pool. Zero fields use the defaults below,
// unless the same option is already set in URL (e.g. ?pool_max_conns=20), so
// ops can tune a deployment through DATABASE_URL alone. Precedence:
// Config field > URL parameter > default.
type Config struct {
	URL string

	// ApplicationName is shown in pg_stat_activity and Postgres logs.
	// Set it to the service name.
	ApplicationName string

	// MaxConns per process. Default 10. Total connections are
	// MaxConns x tasks x services, so size it against the database's
	// max_connections (e.g. RDS), or put RDS Proxy / PgBouncer in front.
	MaxConns int32
	// MinConns kept open when idle. Default 0.
	MinConns int32
	// MaxConnLifetime before a connection is recycled, plus up to 5 minutes
	// of jitter. Default 30m. Recycling picks up failovers, DNS changes and
	// rotated credentials.
	MaxConnLifetime time.Duration
	// MaxConnIdleTime before an idle connection is closed. Default 5m.
	MaxConnIdleTime time.Duration

	// ConnectTimeout for dialing a new connection. Default 5s.
	ConnectTimeout time.Duration
	// StatementTimeout cancels any single query running longer. Default 30s;
	// negative disables. Migrations run through the migrate CLI, not this pool.
	StatementTimeout time.Duration
	// IdleInTransactionTimeout ends sessions that sit in an open transaction,
	// so a stuck request can't hold locks forever. Default 60s; negative disables.
	IdleInTransactionTimeout time.Duration
}

const (
	defaultMaxConns                 = 10
	defaultMaxConnLifetime          = 30 * time.Minute
	defaultMaxConnLifetimeJitter    = 5 * time.Minute
	defaultMaxConnIdleTime          = 5 * time.Minute
	defaultConnectTimeout           = 5 * time.Second
	defaultStatementTimeout         = 30 * time.Second
	defaultIdleInTransactionTimeout = 60 * time.Second
	startupPingTimeout              = 10 * time.Second
)

// Connect opens a Postgres connection pool and verifies it with a ping.
func Connect(ctx context.Context, cfg Config) (*pgxpool.Pool, error) {
	poolCfg, err := poolConfig(cfg)
	if err != nil {
		return nil, err
	}

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, fmt.Errorf("connect to database: %w", err)
	}

	pingCtx, cancel := context.WithTimeout(ctx, startupPingTimeout)
	defer cancel()
	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}

	return pool, nil
}

func poolConfig(cfg Config) (*pgxpool.Config, error) {
	poolCfg, err := pgxpool.ParseConfig(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("parse database url: %w", err)
	}

	// pgxpool.ParseConfig fills in its own defaults, so look at the raw
	// parameters to tell which options the URL actually set.
	raw, err := pgconn.ParseConfig(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("parse database url: %w", err)
	}
	inURL := func(param string) bool {
		_, ok := raw.RuntimeParams[param]
		return ok
	}

	switch {
	case cfg.MaxConns > 0:
		poolCfg.MaxConns = cfg.MaxConns
	case !inURL("pool_max_conns"):
		poolCfg.MaxConns = defaultMaxConns
	}
	if cfg.MinConns > 0 {
		poolCfg.MinConns = cfg.MinConns
	}
	switch {
	case cfg.MaxConnLifetime > 0:
		poolCfg.MaxConnLifetime = cfg.MaxConnLifetime
	case !inURL("pool_max_conn_lifetime"):
		poolCfg.MaxConnLifetime = defaultMaxConnLifetime
	}
	if !inURL("pool_max_conn_lifetime_jitter") {
		poolCfg.MaxConnLifetimeJitter = defaultMaxConnLifetimeJitter
	}
	switch {
	case cfg.MaxConnIdleTime > 0:
		poolCfg.MaxConnIdleTime = cfg.MaxConnIdleTime
	case !inURL("pool_max_conn_idle_time"):
		poolCfg.MaxConnIdleTime = defaultMaxConnIdleTime
	}

	conn := poolCfg.ConnConfig
	switch {
	case cfg.ConnectTimeout > 0:
		conn.ConnectTimeout = cfg.ConnectTimeout
	case conn.ConnectTimeout == 0: // not set by connect_timeout or PGCONNECT_TIMEOUT
		conn.ConnectTimeout = defaultConnectTimeout
	}

	if cfg.ApplicationName != "" {
		conn.RuntimeParams["application_name"] = cfg.ApplicationName
	}
	setTimeout(conn.RuntimeParams, "statement_timeout", cfg.StatementTimeout, defaultStatementTimeout)
	setTimeout(conn.RuntimeParams, "idle_in_transaction_session_timeout", cfg.IdleInTransactionTimeout, defaultIdleInTransactionTimeout)

	return poolCfg, nil
}

// setTimeout sets a Postgres timeout session parameter (in milliseconds).
func setTimeout(params map[string]string, name string, value, fallback time.Duration) {
	switch {
	case value < 0:
		delete(params, name)
	case value > 0:
		params[name] = strconv.FormatInt(value.Milliseconds(), 10)
	default:
		if _, ok := params[name]; !ok {
			params[name] = strconv.FormatInt(fallback.Milliseconds(), 10)
		}
	}
}
