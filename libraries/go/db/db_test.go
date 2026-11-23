package db

import (
	"testing"
	"time"
)

const url = "postgres://u:p@localhost:5432/app?sslmode=disable"

func TestDefaults(t *testing.T) {
	c, err := poolConfig(Config{URL: url, ApplicationName: "auth"})
	if err != nil {
		t.Fatal(err)
	}
	if c.MaxConns != 10 || c.MinConns != 0 {
		t.Errorf("conns = %d/%d, want 10/0", c.MaxConns, c.MinConns)
	}
	if c.MaxConnLifetime != 30*time.Minute || c.MaxConnLifetimeJitter != 5*time.Minute || c.MaxConnIdleTime != 5*time.Minute {
		t.Errorf("lifetime/jitter/idle = %v/%v/%v", c.MaxConnLifetime, c.MaxConnLifetimeJitter, c.MaxConnIdleTime)
	}
	if c.ConnConfig.ConnectTimeout != 5*time.Second {
		t.Errorf("connect timeout = %v", c.ConnConfig.ConnectTimeout)
	}
	params := c.ConnConfig.RuntimeParams
	if params["statement_timeout"] != "30000" || params["idle_in_transaction_session_timeout"] != "60000" || params["application_name"] != "auth" {
		t.Errorf("runtime params = %v", params)
	}
}

func TestURLOverridesDefaults(t *testing.T) {
	c, err := poolConfig(Config{URL: url + "&pool_max_conns=25&pool_max_conn_idle_time=1m&connect_timeout=2&statement_timeout=1000"})
	if err != nil {
		t.Fatal(err)
	}
	if c.MaxConns != 25 || c.MaxConnIdleTime != time.Minute || c.ConnConfig.ConnectTimeout != 2*time.Second {
		t.Errorf("max/idle/connect = %d/%v/%v", c.MaxConns, c.MaxConnIdleTime, c.ConnConfig.ConnectTimeout)
	}
	if got := c.ConnConfig.RuntimeParams["statement_timeout"]; got != "1000" {
		t.Errorf("statement_timeout = %q, want URL value 1000", got)
	}
}

func TestFieldsOverrideURL(t *testing.T) {
	c, err := poolConfig(Config{URL: url + "&pool_max_conns=25", MaxConns: 3, StatementTimeout: -1})
	if err != nil {
		t.Fatal(err)
	}
	if c.MaxConns != 3 {
		t.Errorf("max conns = %d, want field value 3", c.MaxConns)
	}
	if _, ok := c.ConnConfig.RuntimeParams["statement_timeout"]; ok {
		t.Error("negative StatementTimeout should disable it")
	}
}
