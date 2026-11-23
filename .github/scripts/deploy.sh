#!/usr/bin/env bash
# deploy.sh <env> <sha>: deploy every project whose code changed since it was
# last deployed to <env>, then move its deploy/<env>/<name> tag to <sha>.
#
# Each project is compared against its own tag, not against the previous
# push, so a run that was cancelled or failed never loses changes: they're
# picked up by the next run. FORCE=true deploys everything.
set -euo pipefail
env="${1:?env}" sha="${2:?sha}"
here="$(cd "$(dirname "$0")" && pwd)"
logs="$(mktemp -d)"

base_tag="deploy/$env/{name}"
[ "${FORCE:-false}" = "true" ] && base_tag="force-all/{name}" # no such tags: everything is affected
affected=$(go run ./tools/repo affected --base-tag "$base_tag" --head "$sha")
echo "$affected" | jq '{go_services, apps}'

jobs=()
for name in $(jq -r '.go_services[]' <<<"$affected"); do jobs+=("go $name"); done
for name in $(jq -r '.apps[]' <<<"$affected"); do jobs+=("svelte $name"); done
if [ ${#jobs[@]} -eq 0 ]; then
	echo "Nothing to deploy to $env"
	exit 0
fi

run_one() {
	kind="$1" name="$2"
	if "$here/deploy-$kind.sh" "$env" "$name" "$sha" >"$logs/$name.log" 2>&1; then
		echo "$name" >>"$logs/ok"
	else
		echo "$name" >>"$logs/failed"
	fi
	echo "::group::$name"; cat "$logs/$name.log"; echo "::endgroup::"
}
export -f run_one
export env sha here logs

printf '%s\n' "${jobs[@]}" | xargs -P "${PARALLEL:-8}" -L 1 bash -c 'run_one $0 $1'

if [ -s "$logs/ok" ]; then
	# shellcheck disable=SC2046
	"$here/tag-deployed.sh" "$env" "$sha" $(cat "$logs/ok")
fi
if [ -s "$logs/failed" ]; then
	echo "::error::Failed to deploy to $env: $(tr '\n' ' ' <"$logs/failed")"
	exit 1
fi
