#!/bin/sh
# dev-go <name>: build and run services/<name>, rebuilding when it or the
# shared Go code changes. Migrations run on start (AUTO_MIGRATE).
set -eu
name="${1:?usage: dev-go <service>}"
exec air -c /etc/air.toml \
	-build.cmd "go build -o /tmp/air/$name ./services/$name/cmd/http" \
	-build.bin "/tmp/air/$name" \
	-build.include_dir "services/$name,libraries/go,rpc/generated/go"
