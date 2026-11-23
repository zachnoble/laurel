package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strings"

	"go.yaml.in/yaml/v3"
)

// Project is a project.yaml: the single description of a service or app
// that everything else (local dev, CI, infra) is generated from.
type Project struct {
	Name        string `yaml:"name"`
	Kind        string `yaml:"kind"`
	Description string `yaml:"description"`

	// go-service

	// Database: the service has its own Postgres database (migrations in
	// internal/db/migrations, embedded and applied by `server migrate`).
	Database bool `yaml:"database,omitempty"`
	// Calls lists the services this one calls (local dev starts them too).
	Calls []string `yaml:"calls,omitempty"`
	// Secrets are environment variables provided from Secrets Manager in
	// deployed environments (values set out of band, see infra/README.md).
	Secrets []string `yaml:"secrets,omitempty"`
	// Env is plain environment for every deployed environment.
	Env map[string]string `yaml:"env,omitempty"`
	// Deploy sizes the ECS service.
	Deploy *Deploy `yaml:"deploy,omitempty"`

	// svelte-app

	// Subdomain the app is served at, under the environment's domain.
	Subdomain string `yaml:"subdomain,omitempty"`
	// Port for local dev (unique per app).
	Port int `yaml:"port,omitempty"`

	// Dir is the project's directory, relative to the repo root.
	Dir string `yaml:"-"`
}

type Deploy struct {
	CPU    int `yaml:"cpu,omitempty" json:"cpu"`
	Memory int `yaml:"memory,omitempty" json:"memory"`
	Count  int `yaml:"count,omitempty" json:"count"`
	// Public puts the service behind the load balancer. Only the gateway.
	Public bool `yaml:"public,omitempty" json:"public"`
}

const (
	KindGoService = "go-service"
	KindSvelteApp = "svelte-app"
	manifestFile  = "project.yaml"
)

// projectRoots are where project.yaml files live, one level deep. Libraries
// have none.
var projectRoots = []string{"services", "apps"}

var (
	namePattern = regexp.MustCompile(`^[a-z][a-z0-9]*$`)
	envPattern  = regexp.MustCompile(`^[A-Z][A-Z0-9_]*$`)
)

type Repo struct {
	Root     string
	Projects []*Project
}

func (r *Repo) ByName(name string) *Project {
	for _, p := range r.Projects {
		if p.Name == name {
			return p
		}
	}
	return nil
}

func (r *Repo) OfKind(kind string) []*Project {
	var out []*Project
	for _, p := range r.Projects {
		if p.Kind == kind {
			out = append(out, p)
		}
	}
	return out
}

func LoadRepo(root string) (*Repo, error) {
	r := &Repo{Root: root}
	for _, base := range projectRoots {
		entries, err := os.ReadDir(filepath.Join(root, base))
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			return nil, err
		}
		for _, e := range entries {
			if !e.IsDir() {
				continue
			}
			dir := filepath.ToSlash(filepath.Join(base, e.Name()))
			data, err := os.ReadFile(filepath.Join(root, dir, manifestFile))
			if errors.Is(err, os.ErrNotExist) {
				return nil, fmt.Errorf("%s has no %s (every project needs one)", dir, manifestFile)
			}
			if err != nil {
				return nil, err
			}
			p := &Project{}
			dec := yaml.NewDecoder(strings.NewReader(string(data)))
			dec.KnownFields(true)
			if err := dec.Decode(p); err != nil {
				return nil, fmt.Errorf("%s/%s: %w", dir, manifestFile, err)
			}
			p.Dir = dir
			r.Projects = append(r.Projects, p)
		}
	}
	sort.Slice(r.Projects, func(i, j int) bool { return r.Projects[i].Dir < r.Projects[j].Dir })
	return r, r.validate()
}

func (r *Repo) validate() error {
	var errs []error
	fail := func(p *Project, format string, args ...any) {
		errs = append(errs, fmt.Errorf("%s/%s: %s", p.Dir, manifestFile, fmt.Sprintf(format, args...)))
	}

	names := map[string]string{}
	ports := map[int]string{}
	subdomains := map[string]string{}
	for _, p := range r.Projects {
		if !namePattern.MatchString(p.Name) {
			fail(p, "name %q must be lowercase letters and digits, starting with a letter", p.Name)
		}
		if filepath.Base(p.Dir) != p.Name {
			fail(p, "name %q must match the directory name", p.Name)
		}
		// Services and apps share DNS-ish namespaces (compose, ECS, ECR); keep
		// names unique across kinds.
		if other, ok := names[p.Name]; ok {
			fail(p, "name %q is also used by %s", p.Name, other)
		}
		names[p.Name] = p.Dir

		wantRoot := map[string]string{KindGoService: "services", KindSvelteApp: "apps"}[p.Kind]
		if wantRoot == "" {
			fail(p, "kind %q must be go-service or svelte-app", p.Kind)
			continue
		}
		if filepath.ToSlash(filepath.Dir(p.Dir)) != wantRoot {
			fail(p, "a %s belongs in %s/", p.Kind, wantRoot)
		}

		if p.Kind != KindGoService && (p.Database || len(p.Calls) > 0 || len(p.Secrets) > 0 || len(p.Env) > 0 || p.Deploy != nil) {
			fail(p, "database, calls, secrets, env and deploy are only for go-service")
		}
		if p.Kind != KindSvelteApp && (p.Subdomain != "" || p.Port != 0) {
			fail(p, "subdomain and port are only for svelte-app")
		}

		switch p.Kind {
		case KindGoService:
			if _, err := os.Stat(filepath.Join(r.Root, p.Dir, "cmd", "http", "main.go")); err != nil {
				fail(p, "a go-service needs cmd/http/main.go")
			}
			if p.Database {
				if _, err := os.Stat(filepath.Join(r.Root, p.Dir, "internal", "db", "migrations")); err != nil {
					fail(p, "database: true needs internal/db/migrations")
				}
			}
			for _, s := range p.Secrets {
				if !envPattern.MatchString(s) {
					fail(p, "secrets: %q is not an environment variable name", s)
				}
			}
			for k := range p.Env {
				if !envPattern.MatchString(k) {
					fail(p, "env: %q is not an environment variable name", k)
				}
			}
		case KindSvelteApp:
			if p.Port == 0 {
				fail(p, "a svelte-app needs a local dev port")
			} else if other, ok := ports[p.Port]; ok {
				fail(p, "port %d is also used by %s", p.Port, other)
			}
			ports[p.Port] = p.Dir
			if p.Subdomain == "" {
				fail(p, "a svelte-app needs a subdomain")
			} else if other, ok := subdomains[p.Subdomain]; ok {
				fail(p, "subdomain %q is also used by %s", p.Subdomain, other)
			}
			subdomains[p.Subdomain] = p.Dir
		}
	}

	for _, p := range r.OfKind(KindGoService) {
		for _, c := range p.Calls {
			if q := r.ByName(c); q == nil || q.Kind != KindGoService {
				fail(p, "calls: %q is not a go-service", c)
			}
			if c == p.Name {
				fail(p, "calls: a service doesn't need to list itself")
			}
		}
	}
	if cycle := r.callCycle(); cycle != nil {
		errs = append(errs, fmt.Errorf("calls form a cycle (%s); local dev starts dependencies first, so break it", strings.Join(cycle, " -> ")))
	}
	return errors.Join(errs...)
}

func (r *Repo) callCycle() []string {
	const (
		unvisited = iota
		visiting
		done
	)
	state := map[string]int{}
	var stack []string
	var visit func(name string) []string
	visit = func(name string) []string {
		switch state[name] {
		case visiting:
			i := slices.Index(stack, name)
			return append(slices.Clone(stack[i:]), name)
		case done:
			return nil
		}
		state[name] = visiting
		stack = append(stack, name)
		if p := r.ByName(name); p != nil {
			for _, c := range p.Calls {
				if cycle := visit(c); cycle != nil {
					return cycle
				}
			}
		}
		stack = stack[:len(stack)-1]
		state[name] = done
		return nil
	}
	for _, p := range r.OfKind(KindGoService) {
		if cycle := visit(p.Name); cycle != nil {
			return cycle
		}
	}
	return nil
}
