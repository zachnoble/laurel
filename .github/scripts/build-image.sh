#!/usr/bin/env bash
# build-image.sh <service> <sha>: build and push the service's image to ECR
# as <repo>:<sha>, unless that tag already exists.
#
# Builds are reproducible (docker/go-service.Dockerfile, SOURCE_DATE_EPOCH=0,
# rewrite-timestamp), so unchanged source yields the same digest and
# deploy-go.sh skips the rollout.
set -euo pipefail
name="${1:?service}" sha="${2:?sha}"
repo="laurel/$name"

if aws ecr describe-images --repository-name "$repo" --image-ids imageTag="$sha" >/dev/null 2>&1; then
	echo "$name: image $sha already built"
	exit 0
fi

uri=$(aws ecr describe-repositories --repository-names "$repo" --query 'repositories[0].repositoryUri' --output text)
SOURCE_DATE_EPOCH=0 docker buildx build \
	--build-arg SERVICE="$name" \
	-f docker/go-service.Dockerfile \
	--cache-from "type=gha,scope=$name" \
	--cache-to "type=gha,mode=max,scope=$name" \
	--output "type=image,name=$uri:$sha,push=true,rewrite-timestamp=true" \
	--provenance=false \
	.
