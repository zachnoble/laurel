package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"slices"
	"strings"
)

// Affected answers "what does this change touch?" precisely enough that a
// change to one service builds, tests and deploys that service and nothing
// else:
//
//   - Go: the real package import graph (go list), not directories.
//   - TypeScript: the pnpm workspace graph.
//
// Dependency files are coarse on purpose: a go.mod change affects all Go, a
// pnpm-lock.yaml change all TypeScript. That costs CI time only; deploys
// still skip services whose image digest, and apps whose build, didn't change.
//
// Each project can be compared against its own base (e.g. the commit last
// deployed to an environment), which is how deploys never skip changes.

type goPackage struct {
	ImportPath   string
	Imports      []string
	TestImports  []string
	XTestImports []string
	EmbedFiles   []string
}

// goGraph is the module's own packages (dependencies outside it are covered
// by the go.mod rule).
type goGraph struct {
	module   string
	packages map[string]*goPackage
}

func loadGoGraph(root string) (*goGraph, error) {
	module, err := goCmd(root, "list", "-m")
	if err != nil {
		return nil, err
	}
	out, err := goCmd(root, "list", "-e", "-json=ImportPath,Imports,TestImports,XTestImports,EmbedFiles", "./...")
	if err != nil {
		return nil, err
	}

	g := &goGraph{module: strings.TrimSpace(module), packages: map[string]*goPackage{}}
	dec := json.NewDecoder(strings.NewReader(out))
	for {
		var p goPackage
		if err := dec.Decode(&p); errors.Is(err, io.EOF) {
			break
		} else if err != nil {
			return nil, err
		}
		g.packages[p.ImportPath] = &p
	}
	return g, nil
}

func goCmd(root string, args ...string) (string, error) {
	cmd := exec.Command("go", args...)
	cmd.Dir = root
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("go %s: %w: %s", strings.Join(args, " "), err, stderr.String())
	}
	return string(out), nil
}

// relDir is a package's directory relative to the repo root.
func (g *goGraph) relDir(importPath string) string {
	return strings.TrimPrefix(strings.TrimPrefix(importPath, g.module), "/")
}

// deps is what a set of packages depends on within the module: package
// directories and embedded files.
type deps struct {
	dirs  map[string]bool
	files map[string]bool
}

// depsOf walks the import graph from roots. withTests adds the roots' own
// test imports (and what those import).
func (g *goGraph) depsOf(roots []string, withTests bool) deps {
	d := deps{dirs: map[string]bool{}, files: map[string]bool{}}
	var visit func(string)
	visit = func(ip string) {
		p, ok := g.packages[ip]
		if !ok || d.dirs[g.relDir(ip)] {
			return
		}
		dir := g.relDir(ip)
		d.dirs[dir] = true
		for _, f := range p.EmbedFiles {
			d.files[path.Join(dir, f)] = true
		}
		for _, imp := range p.Imports {
			visit(imp)
		}
	}
	for _, r := range roots {
		visit(r)
		if p, ok := g.packages[r]; ok && withTests {
			for _, imp := range slices.Concat(p.TestImports, p.XTestImports) {
				visit(imp)
			}
		}
	}
	return d
}

// packagesUnder lists packages in dir and below.
func (g *goGraph) packagesUnder(dir string) []string {
	var out []string
	for ip := range g.packages {
		if rel := g.relDir(ip); rel == dir || strings.HasPrefix(rel, dir+"/") {
			out = append(out, ip)
		}
	}
	slices.Sort(out)
	return out
}

// change is what differs between a base commit and head.
type change struct {
	files []string
	// goAll / tsAll: a dependency file changed (go.mod, pnpm-lock.yaml).
	goAll, tsAll bool
}

func (c change) touches(d deps) bool {
	if c.goAll {
		return true
	}
	for _, f := range c.files {
		if d.files[f] || d.dirs[path.Dir(f)] {
			return true
		}
	}
	return false
}

func (c change) touchesAny(prefixes ...string) bool {
	for _, f := range c.files {
		for _, p := range prefixes {
			if f == p || strings.HasPrefix(f, strings.TrimSuffix(p, "/")+"/") {
				return true
			}
		}
	}
	return false
}

func computeChange(root, base, head string) (change, error) {
	var c change
	args := []string{"diff", "--name-only", "--no-renames", base}
	if head != "" {
		args = append(args, head)
	}
	out, err := git(root, args...)
	if err != nil {
		return c, err
	}
	c.files = strings.Fields(out)
	if head == "" { // the working tree: include new, untracked files
		untracked, err := git(root, "ls-files", "--others", "--exclude-standard")
		if err != nil {
			return c, err
		}
		c.files = append(c.files, strings.Fields(untracked)...)
	}
	c.goAll = slices.Contains(c.files, "go.mod") || slices.Contains(c.files, "go.sum")
	c.tsAll = slices.Contains(c.files, "pnpm-lock.yaml") || slices.Contains(c.files, "pnpm-workspace.yaml")
	return c, nil
}

// Affected is the output of `repo affected`.
type Affected struct {
	// GoServices are deployable Go services to build/deploy.
	GoServices []string `json:"go_services"`
	// Apps are SvelteKit apps to build/deploy.
	Apps []string `json:"apps"`
	// GoPackages are local Go packages to vet and test.
	GoPackages []string `json:"go_packages"`
	// TSPackages are workspace package dirs to typecheck and test.
	TSPackages []string `json:"ts_packages"`
	// Proto: the API schema or its codegen changed.
	Proto bool `json:"proto"`
	// Migrations are added or changed up migrations, to lint.
	Migrations []string `json:"migrations"`
}

type affectedOptions struct {
	// Head is the commit to compare, or "" for the working tree.
	Head string
	// Base compares every project against one commit (PRs: the merge base).
	Base string
	// BaseTag compares each project against its own tag, e.g.
	// "deploy/production/{name}". Projects without the tag are affected.
	BaseTag string
}

// goServiceImageInputs are files outside the import graph that change every
// service's image.
var goServiceImageInputs = []string{"docker/go-service.Dockerfile"}

func computeAffected(r *Repo, opts affectedOptions) (*Affected, error) {
	g, err := loadGoGraph(r.Root)
	if err != nil {
		return nil, err
	}
	ws, err := loadWorkspace(r.Root)
	if err != nil {
		return nil, err
	}

	changes := map[string]*change{}
	changeFor := func(project string) (*change, error) {
		base := opts.Base
		if opts.BaseTag != "" {
			tag := strings.ReplaceAll(opts.BaseTag, "{name}", project)
			sha, err := git(r.Root, "rev-parse", "--verify", "--quiet", "refs/tags/"+tag+"^{commit}")
			if err != nil || strings.TrimSpace(sha) == "" {
				return nil, nil // never deployed: affected
			}
			base = strings.TrimSpace(sha)
		}
		if c, ok := changes[base]; ok {
			return c, nil
		}
		c, err := computeChange(r.Root, base, opts.Head)
		if err != nil {
			return nil, err
		}
		changes[base] = &c
		return &c, nil
	}

	a := &Affected{GoServices: []string{}, Apps: []string{}, GoPackages: []string{}, TSPackages: []string{}, Migrations: []string{}}

	for _, p := range r.OfKind(KindGoService) {
		c, err := changeFor(p.Name)
		if err != nil {
			return nil, err
		}
		pkgs := g.packagesUnder(p.Dir)
		if c == nil || c.touches(g.depsOf(pkgs, false)) || c.touchesAny(p.Dir) || c.touchesAny(goServiceImageInputs...) {
			a.GoServices = append(a.GoServices, p.Name)
		}
	}

	tsAffected := func(dir string, c *change) bool {
		if c == nil || c.tsAll {
			return true
		}
		for _, d := range ws.closure(dir) {
			if c.touchesAny(d) {
				return true
			}
		}
		return false
	}
	for _, p := range r.OfKind(KindSvelteApp) {
		c, err := changeFor(p.Name)
		if err != nil {
			return nil, err
		}
		if tsAffected(p.Dir, c) {
			a.Apps = append(a.Apps, p.Name)
		}
	}

	// Checks (packages, schema, migrations) only make sense against one base.
	if opts.Base != "" {
		c, err := changeFor("")
		if err != nil {
			return nil, err
		}
		for _, ip := range slices.Sorted(mapKeys(g.packages)) {
			if c.touches(g.depsOf([]string{ip}, true)) {
				a.GoPackages = append(a.GoPackages, ip)
			}
		}
		for _, dir := range slices.Sorted(mapKeys(ws.packages)) {
			if tsAffected(dir, c) {
				a.TSPackages = append(a.TSPackages, dir)
			}
		}
		a.Proto = c.touchesAny("rpc/proto", "buf.yaml", "buf.lock", "rpc/generators")
		for _, f := range c.files {
			if strings.HasSuffix(f, ".up.sql") && strings.Contains(f, "/internal/db/migrations/") {
				if _, err := readFileAt(r.Root, opts.Head, f); err == nil {
					a.Migrations = append(a.Migrations, f)
				}
			}
		}
	}
	return a, nil
}

// shard splits items into at most n groups (round robin, so groups stay
// balanced), each a space-separated string: GitHub matrices are capped at
// 256 entries, and hundreds of near-empty jobs are slow anyway.
func shard(items []string, n int) []map[string]string {
	out := []map[string]string{}
	if len(items) == 0 {
		return out
	}
	if n <= 0 || n > len(items) {
		n = len(items)
	}
	groups := make([][]string, n)
	for i, it := range items {
		groups[i%n] = append(groups[i%n], it)
	}
	for i, grp := range groups {
		out = append(out, map[string]string{
			"shard":    fmt.Sprintf("%d/%d", i+1, n),
			"projects": strings.Join(grp, " "),
		})
	}
	return out
}

func mapKeys[K comparable, V any](m map[K]V) func(func(K) bool) {
	return func(yield func(K) bool) {
		for k := range m {
			if !yield(k) {
				return
			}
		}
	}
}

func git(root string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = root
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git %s: %w: %s", strings.Join(args, " "), err, stderr.String())
	}
	return string(out), nil
}

// readFileAt reads a file at a commit, or from the working tree if rev is "".
func readFileAt(root, rev, name string) ([]byte, error) {
	if rev == "" {
		return os.ReadFile(filepath.Join(root, name))
	}
	out, err := git(root, "show", rev+":"+name)
	return []byte(out), err
}
