#!/usr/bin/env bash
# deploy-go.sh <env> <service> <sha> [--skip-migrations]
#
# Rolls out <repo>:<sha> to the service's ECS service:
#   1. resolve the image digest; skip if it is already what's running
#   2. register a task definition revision with the new image
#   3. if the service has a database, run `server migrate` as a one-off task
#      with the new revision and require exit 0 (old tasks keep serving, so
#      migrations must be backward compatible)
#   4. update the service and wait for the rollout; ECS's deployment circuit
#      breaker rolls back failed rollouts, which this treats as failure
#   5. record the image for infra (SSM <prefix>/services/<name>/image)
set -euo pipefail
env="${1:?env}" name="${2:?service}" sha="${3:?sha}" skip_migrations="${4:-}"
prefix="/laurel/$env"
param() { aws ssm get-parameter --name "$1" --query Parameter.Value --output text; }

cluster=$(param "$prefix/cluster")
repo="laurel/$name"

if ! digest=$(aws ecr describe-images --repository-name "$repo" --image-ids imageTag="$sha" --query 'imageDetails[0].imageDigest' --output text 2>/dev/null); then
	echo "$name: no image for $sha; building it"
	"$(dirname "$0")/build-image.sh" "$name" "$sha"
	digest=$(aws ecr describe-images --repository-name "$repo" --image-ids imageTag="$sha" --query 'imageDetails[0].imageDigest' --output text)
fi
uri=$(aws ecr describe-repositories --repository-names "$repo" --query 'repositories[0].repositoryUri' --output text)
image="$uri@$digest"

current_td=$(aws ecs describe-services --cluster "$cluster" --services "$name" --query 'services[0].taskDefinition' --output text)
current_def=$(aws ecs describe-task-definition --task-definition "$current_td" --query taskDefinition)
current_image=$(jq -r '.containerDefinitions[] | select(.name == "app") | .image' <<<"$current_def")
if [ "$current_image" = "$image" ]; then
	echo "$name: $image is already running; nothing to deploy"
	exit 0
fi

new_def=$(jq --arg image "$image" --arg sha "$sha" '
	del(.taskDefinitionArn, .revision, .status, .requiresAttributes, .compatibilities, .registeredAt, .registeredBy, .deregisteredAt)
	| .containerDefinitions |= map(
		if .name == "app" then
			.image = $image
			| .environment = (((.environment // []) | map(select(.name != "SERVICE_VERSION"))) + [{name: "SERVICE_VERSION", value: $sha}])
		else . end)' <<<"$current_def")
new_td=$(aws ecs register-task-definition --cli-input-json "$new_def" --query 'taskDefinition.taskDefinitionArn' --output text)
echo "$name: registered $new_td"

has_db=$(jq -r --arg n "$name" '.services[$n].database // false' infra/projects.json)
if [ "$has_db" = "true" ] && [ "$skip_migrations" != "--skip-migrations" ]; then
	network=$(param "$prefix/network")
	task=$(aws ecs run-task --cluster "$cluster" --task-definition "$new_td" --launch-type FARGATE \
		--network-configuration "$network" \
		--overrides '{"containerOverrides":[{"name":"app","command":["migrate"]}]}' \
		--started-by "deploy-$sha" --query 'tasks[0].taskArn' --output text)
	echo "$name: migrating ($task)"
	aws ecs wait tasks-stopped --cluster "$cluster" --tasks "$task"
	code=$(aws ecs describe-tasks --cluster "$cluster" --tasks "$task" --query 'tasks[0].containers[?name==`app`].exitCode | [0]' --output text)
	if [ "$code" != "0" ]; then
		echo "::error::$name: migrations failed (exit $code); not rolling out. Logs: CloudWatch /laurel/$env/$name"
		exit 1
	fi
fi

aws ecs update-service --cluster "$cluster" --service "$name" --task-definition "$new_td" >/dev/null
echo "$name: rolling out"
aws ecs wait services-stable --cluster "$cluster" --services "$name"

primary=$(aws ecs describe-services --cluster "$cluster" --services "$name" --query 'services[0].deployments[?status==`PRIMARY`] | [0]')
if [ "$(jq -r .taskDefinition <<<"$primary")" != "$new_td" ] || [ "$(jq -r .rolloutState <<<"$primary")" != "COMPLETED" ]; then
	echo "::error::$name: rollout did not complete ($(jq -r '.taskDefinition + " " + .rolloutState' <<<"$primary")); ECS rolled back"
	exit 1
fi

aws ssm put-parameter --name "$prefix/services/$name/image" --value "$image" --overwrite >/dev/null
echo "$name: deployed $sha"
