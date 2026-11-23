package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// fixture is a tiny monorepo in a temp git repo: services a (imports
// library x) and b, and apps web (uses library ui) and admin.
func fixture(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	files := map[string]string{
		"go.mod":              "module example.com/m\n\ngo 1.27.4\n",
		"pnpm-workspace.yaml": "packages:\n  - apps/*\n  - libraries/ts/*\n",

		"libraries/go/x/x.go": "package x\n\nfunc X() {}\n",

		"services/a/project.yaml":     "name: a\nkind: go-service\ncalls: [b]\n",
		"services/a/cmd/http/main.go": "package main\n\nimport \"example.com/m/libraries/go/x\"\n\nfunc main() { x.X() }\n",
		"services/b/project.yaml":     "name: b\nkind: go-service\n",
		"services/b/cmd/http/main.go": "package main\n\nfunc main() {}\n",

		"libraries/ts/ui/package.json":  `{"name": "@laurel/ui"}`,
		"libraries/ts/ui/src/index.ts":  "export {}\n",
		"apps/web/project.yaml":         "name: web\nkind: svelte-app\nsubdomain: app\nport: 5173\n",
		"apps/web/package.json":         `{"name": "web", "dependencies": {"@laurel/ui": "workspace:*"}, "devDependencies": {"@sveltejs/kit": "^3.0.0"}}`,
		"apps/admin/project.yaml":       "name: admin\nkind: svelte-app\nsubdomain: admin\nport: 5174\n",
		"apps/admin/package.json":       `{"name": "admin", "devDependencies": {"@sveltejs/kit": "^3.0.0"}}`,
		"apps/admin/src/routes/page.ts": "export {}\n",
	}
	for name, content := range files {
		write(t, root, name, content)
	}
	sh(t, root, "git", "init", "-q", "-b", "main")
	sh(t, root, "git", "-c", "user.email=t@example.com", "-c", "user.name=t", "add", ".")
	commit(t, root, "init")
	return root
}

func write(t *testing.T, root, name, content string) {
	t.Helper()
	path := filepath.Join(root, name)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func sh(t *testing.T, dir string, name string, args ...string) string {
	t.Helper()
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GOWORK=off", "GOFLAGS=-mod=mod")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("%s %v: %v\n%s", name, args, err, out)
	}
	return strings.TrimSpace(string(out))
}

func commit(t *testing.T, root, msg string) string {
	t.Helper()
	sh(t, root, "git", "add", "-A")
	sh(t, root, "git", "-c", "user.email=t@example.com", "-c", "user.name=t", "commit", "-q", "-m", msg)
	return sh(t, root, "git", "rev-parse", "HEAD")
}

func affected(t *testing.T, root string, opts affectedOptions) *Affected {
	t.Helper()
	r, err := LoadRepo(root)
	if err != nil {
		t.Fatal(err)
	}
	a, err := computeAffected(r, opts)
	if err != nil {
		t.Fatal(err)
	}
	return a
}

func TestLibraryChangeAffectsOnlyImporters(t *testing.T) {
	root := fixture(t)
	base := sh(t, root, "git", "rev-parse", "HEAD")
	write(t, root, "libraries/go/x/x.go", "package x\n\nfunc X() { println() }\n")
	head := commit(t, root, "change x")

	a := affected(t, root, affectedOptions{Base: base, Head: head})
	if !slices.Equal(a.GoServices, []string{"a"}) {
		t.Errorf("go services = %v, want [a]", a.GoServices)
	}
	if slices.Contains(a.GoPackages, "example.com/m/services/b/cmd/http") || !slices.Contains(a.GoPackages, "example.com/m/libraries/go/x") {
		t.Errorf("go packages = %v", a.GoPackages)
	}
	if len(a.Apps) != 0 {
		t.Errorf("apps = %v, want none", a.Apps)
	}
}

func TestWorkspaceChangeAffectsDependentApps(t *testing.T) {
	root := fixture(t)
	base := sh(t, root, "git", "rev-parse", "HEAD")
	write(t, root, "libraries/ts/ui/src/index.ts", "export const x = 1\n")
	head := commit(t, root, "change ui")

	a := affected(t, root, affectedOptions{Base: base, Head: head})
	if !slices.Equal(a.Apps, []string{"web"}) {
		t.Errorf("apps = %v, want [web]", a.Apps)
	}
	if !slices.Equal(a.TSPackages, []string{"apps/web", "libraries/ts/ui"}) {
		t.Errorf("ts packages = %v", a.TSPackages)
	}
	if len(a.GoServices) != 0 {
		t.Errorf("go services = %v, want none", a.GoServices)
	}
}

func TestWorkingTreeChangesCount(t *testing.T) {
	root := fixture(t)
	write(t, root, "services/b/cmd/http/extra.go", "package main\n")
	a := affected(t, root, affectedOptions{Base: "HEAD"})
	if !slices.Equal(a.GoServices, []string{"b"}) {
		t.Errorf("go services = %v, want [b] (untracked file)", a.GoServices)
	}
}

func TestPerProjectBaseTags(t *testing.T) {
	root := fixture(t)
	first := sh(t, root, "git", "rev-parse", "HEAD")
	sh(t, root, "git", "tag", "deploy/prod/a", first)
	sh(t, root, "git", "tag", "deploy/prod/web", first)

	// A change to b lands; b was never deployed, a's last deploy predates it.
	write(t, root, "services/b/cmd/http/main.go", "package main\n\nfunc main() { println() }\n")
	commit(t, root, "change b")
	// Later a change to a: a deploy that only diffed the last push would
	// still see it, but b's earlier change must not be lost either.
	write(t, root, "libraries/go/x/x.go", "package x\n\nfunc X() { println(1) }\n")
	head := commit(t, root, "change x")

	got := affected(t, root, affectedOptions{BaseTag: "deploy/prod/{name}", Head: head})
	if !slices.Equal(got.GoServices, []string{"a", "b"}) {
		t.Errorf("go services = %v, want [a b]", got.GoServices)
	}
	if !slices.Equal(got.Apps, []string{"admin"}) { // never deployed
		t.Errorf("apps = %v, want [admin]", got.Apps)
	}

	sh(t, root, "git", "tag", "-f", "deploy/prod/a", head)
	sh(t, root, "git", "tag", "deploy/prod/b", head)
	sh(t, root, "git", "tag", "deploy/prod/admin", head)
	got = affected(t, root, affectedOptions{BaseTag: "deploy/prod/{name}", Head: head})
	if len(got.GoServices)+len(got.Apps) != 0 {
		t.Errorf("after deploying everything: %v %v, want none", got.GoServices, got.Apps)
	}
}

func TestDependencyFilesAffectTheirLanguage(t *testing.T) {
	root := fixture(t)
	base := sh(t, root, "git", "rev-parse", "HEAD")
	write(t, root, "go.mod", "module example.com/m\n\ngo 1.27.4\n\n// bumped\n")
	head := commit(t, root, "bump go deps")
	a := affected(t, root, affectedOptions{Base: base, Head: head})
	if !slices.Equal(a.GoServices, []string{"a", "b"}) || len(a.Apps) != 0 {
		t.Errorf("go.mod change: services %v apps %v, want all services, no apps", a.GoServices, a.Apps)
	}

	write(t, root, "pnpm-lock.yaml", "lockfileVersion: '9.0'\n")
	next := commit(t, root, "bump ts deps")
	a = affected(t, root, affectedOptions{Base: head, Head: next})
	if !slices.Equal(a.Apps, []string{"admin", "web"}) || len(a.GoServices) != 0 {
		t.Errorf("lockfile change: apps %v services %v, want all apps, no services", a.Apps, a.GoServices)
	}
}

func TestShard(t *testing.T) {
	got := shard([]string{"a", "b", "c", "d", "e"}, 2)
	if len(got) != 2 || got[0]["projects"] != "a c e" || got[1]["projects"] != "b d" {
		t.Errorf("shards = %v", got)
	}
	if got := shard(nil, 3); len(got) != 0 {
		t.Errorf("empty = %v", got)
	}
}

func TestValidateRejectsBadManifests(t *testing.T) {
	root := fixture(t)
	write(t, root, "services/b/project.yaml", "name: b\nkind: go-service\ncalls: [a]\n") // a calls b
	if _, err := LoadRepo(root); err == nil || !strings.Contains(err.Error(), "cycle") {
		t.Errorf("err = %v, want a cycle error", err)
	}

	root = fixture(t)
	write(t, root, "apps/admin/project.yaml", "name: admin\nkind: svelte-app\nsubdomain: admin\nport: 5173\n")
	if _, err := LoadRepo(root); err == nil || !strings.Contains(err.Error(), "port 5173") {
		t.Errorf("err = %v, want a duplicate port error", err)
	}

	root = fixture(t)
	write(t, root, "services/b/project.yaml", "name: b\nkind: go-service\ndatabse: true\n")
	if _, err := LoadRepo(root); err == nil || !strings.Contains(err.Error(), "databse") {
		t.Errorf("err = %v, want an unknown field error", err)
	}
}
