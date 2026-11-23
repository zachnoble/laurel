package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"go.yaml.in/yaml/v3"
)

// The pnpm workspace: package directories and their workspace:* edges.

type tsPackage struct {
	Name string
	Dir  string
	// Kit marks a SvelteKit app.
	Kit bool
	// Deps are the directories of workspace packages it depends on, directly.
	Deps    []string
	Scripts map[string]string
}

type workspace struct {
	packages map[string]*tsPackage // by dir
}

func loadWorkspace(root string) (*workspace, error) {
	data, err := os.ReadFile(filepath.Join(root, "pnpm-workspace.yaml"))
	if err != nil {
		return nil, err
	}
	var cfg struct {
		Packages []string `yaml:"packages"`
	}
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("pnpm-workspace.yaml: %w", err)
	}

	type manifest struct {
		Name            string            `json:"name"`
		Dependencies    map[string]string `json:"dependencies"`
		DevDependencies map[string]string `json:"devDependencies"`
		Scripts         map[string]string `json:"scripts"`
	}
	manifests := map[string]manifest{}
	byName := map[string]string{}
	for _, pattern := range cfg.Packages {
		matches, err := filepath.Glob(filepath.Join(root, pattern, "package.json"))
		if err != nil {
			return nil, err
		}
		for _, m := range matches {
			data, err := os.ReadFile(m)
			if err != nil {
				return nil, err
			}
			var pkg manifest
			if err := json.Unmarshal(data, &pkg); err != nil {
				return nil, fmt.Errorf("%s: %w", m, err)
			}
			rel, _ := filepath.Rel(root, filepath.Dir(m))
			dir := filepath.ToSlash(rel)
			manifests[dir] = pkg
			byName[pkg.Name] = dir
		}
	}

	ws := &workspace{packages: map[string]*tsPackage{}}
	for dir, m := range manifests {
		p := &tsPackage{Name: m.Name, Dir: dir, Scripts: m.Scripts}
		for _, deps := range []map[string]string{m.Dependencies, m.DevDependencies} {
			for name, spec := range deps {
				if name == "@sveltejs/kit" {
					p.Kit = true
				}
				if strings.HasPrefix(spec, "workspace:") {
					if d, ok := byName[name]; ok {
						p.Deps = append(p.Deps, d)
					}
				}
			}
		}
		slices.Sort(p.Deps)
		ws.packages[dir] = p
	}
	return ws, nil
}

// closure returns dir and every workspace package it transitively depends on.
func (w *workspace) closure(dir string) []string {
	seen := map[string]bool{}
	var visit func(string)
	visit = func(d string) {
		if seen[d] {
			return
		}
		seen[d] = true
		if p, ok := w.packages[d]; ok {
			for _, dep := range p.Deps {
				visit(dep)
			}
		}
	}
	visit(dir)
	return slices.Sorted(func(yield func(string) bool) {
		for d := range seen {
			if !yield(d) {
				return
			}
		}
	})
}
