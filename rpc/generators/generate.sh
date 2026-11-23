#!/usr/bin/env bash
# Regenerates rpc/generated/{go,typescript} from rpc/proto/ in one buf invocation.
# Run from the repo root (`just gen`).
set -euo pipefail

find rpc/generated/go -mindepth 1 -maxdepth 1 -type d -exec rm -rf {} +
find rpc/generated/typescript -mindepth 1 -maxdepth 1 -type d -not -name node_modules -exec rm -rf {} +

buf generate --template rpc/generators/buf.gen.yaml
