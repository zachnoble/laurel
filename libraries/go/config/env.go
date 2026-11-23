package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// Env reads typed values from environment variables, collecting parse errors
// so a service can report every bad variable at once via Err.
type Env struct {
	errs []error
}

// Load reads .env files (if present) into the environment without overriding
// variables that are already set. With no arguments it loads ".env".
func Load(files ...string) *Env {
	if len(files) == 0 {
		files = []string{".env"}
	}
	for _, f := range files {
		_ = godotenv.Load(f)
	}
	return &Env{}
}

func (e *Env) String(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func (e *Env) List(key string) []string {
	var out []string
	for item := range strings.SplitSeq(os.Getenv(key), ",") {
		if item = strings.TrimSpace(item); item != "" {
			out = append(out, item)
		}
	}
	return out
}

func (e *Env) Int(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}

	n, err := strconv.Atoi(v)
	if err != nil {
		e.errs = append(e.errs, fmt.Errorf("parse %s: %w", key, err))
		return fallback
	}
	return n
}

// Required returns the variable, recording an error if it is unset.
func (e *Env) Required(key string) string {
	v := os.Getenv(key)
	if v == "" {
		e.errs = append(e.errs, fmt.Errorf("%s is required", key))
	}
	return v
}

func (e *Env) Bool(key string, fallback bool) bool {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		e.errs = append(e.errs, fmt.Errorf("parse %s: %w", key, err))
		return fallback
	}
	return b
}

func (e *Env) Duration(key string, fallback time.Duration) time.Duration {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		e.errs = append(e.errs, fmt.Errorf("parse %s: %w", key, err))
		return fallback
	}
	return d
}

// Map reads "k=v,k=v" pairs.
func (e *Env) Map(key string) map[string]string {
	out := map[string]string{}
	for _, item := range e.List(key) {
		k, v, ok := strings.Cut(item, "=")
		if !ok || k == "" {
			e.errs = append(e.errs, fmt.Errorf("parse %s: %q is not key=value", key, item))
			continue
		}
		out[strings.TrimSpace(k)] = strings.TrimSpace(v)
	}
	return out
}

// Err returns every error collected while reading variables.
func (e *Env) Err() error {
	return errors.Join(e.errs...)
}
