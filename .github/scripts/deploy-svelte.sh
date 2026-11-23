#!/usr/bin/env bash
# deploy-svelte.sh <env> <app> <sha>: build the app for the environment and
# publish it to its bucket, skipping the upload if the build is unchanged.
set -euo pipefail
env="${1:?env}" name="${2:?app}" sha="${3:?sha}"
prefix="/laurel/$env"
param() { aws ssm get-parameter --name "$1" --query Parameter.Value --output text; }

bucket=$(param "$prefix/apps/$name/bucket")
distribution=$(param "$prefix/apps/$name/distribution")
dir="apps/$name"

API_URL=$(param "$prefix/api_url") pnpm --dir "$dir" build

hash=$(cd "$dir/build" && find . -type f -print0 | sort -z | xargs -0 sha256sum | sha256sum | cut -d' ' -f1)
if [ "$(aws s3 cp "s3://$bucket/_deploy/build-hash" - 2>/dev/null || true)" = "$hash" ]; then
	echo "$name: build unchanged; nothing to deploy"
	exit 0
fi

# Hashed assets first (cached forever, never deleted here, so pages already
# loaded keep working), then everything else (revalidated on every load).
aws s3 sync "$dir/build/_app/immutable" "s3://$bucket/_app/immutable" --cache-control "public,max-age=31536000,immutable"
aws s3 sync "$dir/build" "s3://$bucket" --delete --exclude "_app/immutable/*" --exclude "_deploy/*" --cache-control "no-cache"
printf '%s' "$hash" | aws s3 cp - "s3://$bucket/_deploy/build-hash"

aws cloudfront create-invalidation --distribution-id "$distribution" --paths "/*" >/dev/null
echo "$name: deployed $sha"
