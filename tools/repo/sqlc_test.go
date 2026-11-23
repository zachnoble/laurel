package main

import (
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func TestSqlcProjects(t *testing.T) {
	root := fixture(t)
	write(t, root, "services/a/sqlc.yaml", "version: \"2\"\nsql:\n  - engine: postgresql\n    queries: q\n    schema: s\n    gen:\n      go:\n        out: internal/db/sqlc\n")
	r, err := LoadRepo(root)
	if err != nil {
		t.Fatal(err)
	}
	got, err := sqlcProjects(r)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Config != "services/a/sqlc.yaml" || !slices.Equal(got[0].Outs, []string{"services/a/internal/db/sqlc"}) {
		t.Fatalf("got %+v", got)
	}

	write(t, root, "services/b/sqlc.yaml", "version: \"2\"\nsql:\n  - gen:\n      go:\n        out: ../a/sqlc\n")
	if _, err := sqlcProjects(r); err == nil {
		t.Fatal("want an error for output outside the project")
	}
}

func TestRemoveSqlcOutput(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "gone.sql.go", sqlcHeader+"\n\npackage sqlc\n")
	write(t, dir, "mine.go", "package sqlc\n")
	if err := removeSqlcOutput(dir); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, "gone.sql.go")); !os.IsNotExist(err) {
		t.Error("generated file not removed")
	}
	if _, err := os.Stat(filepath.Join(dir, "mine.go")); err != nil {
		t.Error("hand-written file removed")
	}
}
