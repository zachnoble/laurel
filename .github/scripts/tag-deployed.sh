#!/usr/bin/env bash
# tag-deployed.sh <env> <sha> <name>...: record that <sha> is deployed to
# <env> by moving deploy/<env>/<name> tags, and push them.
#
# Tags only move forward (to a descendant) unless ALLOW_BACKWARDS=true
# (rollbacks), so an older run finishing late can't rewind them.
set -euo pipefail
env="${1:?env}" sha="${2:?sha}"
shift 2

refs=()
for name in "$@"; do
	tag="deploy/$env/$name"
	if current=$(git rev-parse -q --verify "refs/tags/$tag^{commit}"); then
		if [ "${ALLOW_BACKWARDS:-false}" != "true" ] && ! git merge-base --is-ancestor "$current" "$sha"; then
			echo "$tag is at $current, not an ancestor of $sha; leaving it"
			continue
		fi
	fi
	git tag -f "$tag" "$sha" >/dev/null
	refs+=("+refs/tags/$tag")
done

if [ ${#refs[@]} -gt 0 ]; then
	git push origin "${refs[@]}"
fi
