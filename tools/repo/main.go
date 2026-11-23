// Command repo manages the monorepo from its project.yaml manifests.
//
//	repo gen [--check]          write (or verify) generated files and sqlc output
//	repo affected [flags]       which projects/packages a change affects (JSON)
//	repo list [--kind k]        list projects
//
// Run through just: `just gen`, `just affected`.
package main

import (
	"cmp"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "repo:", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) == 0 {
		return errors.New("usage: repo gen|affected|list")
	}
	root, err := os.Getwd()
	if err != nil {
		return err
	}
	cmd, args := args[0], args[1:]
	fset := flag.NewFlagSet(cmd, flag.ContinueOnError)

	switch cmd {
	case "gen":
		check := fset.Bool("check", false, "fail if generated files are out of date instead of writing them")
		if err := fset.Parse(args); err != nil {
			return err
		}
		r, err := LoadRepo(root)
		if err != nil {
			return err
		}
		return errors.Join(WriteGenerated(r, *check), GenerateSqlc(r, *check))

	case "affected":
		base := fset.String("base", "", "compare against this commit (for PRs, the merge base is used)")
		baseTag := fset.String("base-tag", "", `compare each project against its own tag, e.g. "deploy/production/{name}"`)
		head := fset.String("head", "", "head commit (default: the working tree, including uncommitted changes)")
		shards := fset.Int("shards", 20, "split go_services/apps into at most this many matrix entries")
		if err := fset.Parse(args); err != nil {
			return err
		}
		if (*base == "") == (*baseTag == "") {
			return errors.New("affected: pass exactly one of --base or --base-tag")
		}
		r, err := LoadRepo(root)
		if err != nil {
			return err
		}
		opts := affectedOptions{Head: *head, BaseTag: *baseTag}
		if *base != "" {
			mb, err := git(root, "merge-base", *base, cmp.Or(*head, "HEAD"))
			if err != nil {
				return err
			}
			opts.Base = strings.TrimSpace(mb)
		}
		a, err := computeAffected(r, opts)
		if err != nil {
			return err
		}
		return writeAffected(a, *shards)

	case "list":
		kind := fset.String("kind", "", "only this kind (go-service, svelte-app)")
		if err := fset.Parse(args); err != nil {
			return err
		}
		r, err := LoadRepo(root)
		if err != nil {
			return err
		}
		for _, p := range r.Projects {
			if *kind == "" || p.Kind == *kind {
				fmt.Printf("%s\t%s\t%s\n", p.Name, p.Kind, p.Dir)
			}
		}
		return nil
	}
	return fmt.Errorf("unknown command %q", cmd)
}

// writeAffected prints the result as JSON and, under GitHub Actions, as step
// outputs (lists as JSON, matrices pre-sharded, packages space-separated).
func writeAffected(a *Affected, shards int) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(a); err != nil {
		return err
	}

	path := os.Getenv("GITHUB_OUTPUT")
	if path == "" {
		return nil
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()

	js := func(v any) string {
		b, _ := json.Marshal(v)
		return string(b)
	}
	lines := []string{
		"go_services=" + js(a.GoServices),
		"go_services_matrix=" + js(shard(a.GoServices, shards)),
		"apps=" + js(a.Apps),
		"apps_matrix=" + js(shard(a.Apps, shards)),
		"go_packages=" + strings.Join(a.GoPackages, " "),
		"ts_packages=" + js(a.TSPackages),
		"ts_matrix=" + js(shard(a.TSPackages, shards)),
		fmt.Sprintf("proto=%t", a.Proto),
		"migrations=" + strings.Join(a.Migrations, " "),
		fmt.Sprintf("any=%t", len(a.GoServices)+len(a.Apps) > 0),
	}
	_, err = f.WriteString(strings.Join(lines, "\n") + "\n")
	return err
}
