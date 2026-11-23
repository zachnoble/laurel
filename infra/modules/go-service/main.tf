# One Go service in one environment: its ECS service (on Service Connect, so
# others reach it at http://<name>:8080), task definition, IAM roles, logs,
# and secrets.

terraform {
  required_providers {
    aws = { source = "hashicorp/aws", version = "~> 6.0" }
  }
}

variable "name" {
  type = string
}

variable "environment" {
  type = string
}

variable "config" {
  description = "The service's entry in infra/projects.json (from its project.yaml)."
  type = object({
    database = bool
    public   = bool
    cpu      = number
    memory   = number
    count    = number
    calls    = list(string)
    secrets  = list(string)
    env      = map(string)
  })
}

variable "extra_env" {
  description = "Environment for this environment only (overrides project.yaml env)."
  type        = map(string)
  default     = {}
}

variable "cluster_arn" {
  type = string
}

variable "namespace_arn" {
  type = string
}

variable "subnet_ids" {
  type = list(string)
}

variable "security_group_ids" {
  type = list(string)
}

variable "target_group_arn" {
  description = "Load balancer target group (public services only)."
  type        = string
  default     = null
}

variable "repository_url" {
  type = string
}

variable "identity_public_keys" {
  type = list(string)
}

variable "otel_exporter_endpoint" {
  type    = string
  default = ""
}

variable "log_retention_days" {
  type    = number
  default = 30
}

data "aws_region" "current" {}

locals {
  prefix      = "/laurel/${var.environment}"
  secret_path = "laurel/${var.environment}/${var.name}"
  secrets     = toset(concat(var.config.secrets, var.config.database ? ["DATABASE_URL"] : []))

  environment = merge(
    var.config.env,
    {
      APP_ENV              = var.environment
      IDENTITY_PUBLIC_KEYS = join(",", var.identity_public_keys)
      UPSTREAM_URL         = "http://{name}:8080"
    },
    var.otel_exporter_endpoint != "" ? { OTEL_EXPORTER_OTLP_ENDPOINT = var.otel_exporter_endpoint } : {},
    var.extra_env,
  )
}

resource "aws_cloudwatch_log_group" "this" {
  name              = "${local.prefix}/${var.name}"
  retention_in_days = var.log_retention_days
}

# --- Secrets -----------------------------------------------------------------
# Declared in project.yaml; values are set out of band (see infra/README.md),
# so they never pass through Terraform state.

resource "aws_secretsmanager_secret" "declared" {
  for_each                = local.secrets
  name                    = "${local.secret_path}/${each.key}"
  recovery_window_in_days = 7
}

# --- Image ---------------------------------------------------------------------
# Deploys (.github/scripts/deploy-go.sh) record the running image here, so a
# Terraform apply registers task definitions with the deployed image rather
# than rolling anything back.

resource "aws_ssm_parameter" "image" {
  name           = "${local.prefix}/services/${var.name}/image"
  type           = "String"
  insecure_value = "${var.repository_url}:bootstrap"
  lifecycle {
    ignore_changes = [insecure_value]
  }
}

# --- IAM -------------------------------------------------------------------------

data "aws_iam_policy_document" "ecs_tasks" {
  statement {
    actions = ["sts:AssumeRole"]
    principals {
      type        = "Service"
      identifiers = ["ecs-tasks.amazonaws.com"]
    }
  }
}

resource "aws_iam_role" "execution" {
  name               = "laurel-${var.environment}-${var.name}-execution"
  assume_role_policy = data.aws_iam_policy_document.ecs_tasks.json
}

resource "aws_iam_role_policy_attachment" "execution" {
  role       = aws_iam_role.execution.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AmazonECSTaskExecutionRolePolicy"
}

resource "aws_iam_role_policy" "execution_secrets" {
  count = length(local.secrets) > 0 ? 1 : 0
  role  = aws_iam_role.execution.id
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect   = "Allow"
      Action   = ["secretsmanager:GetSecretValue"]
      Resource = [for s in aws_secretsmanager_secret.declared : s.arn]
    }]
  })
}

# The service's own AWS permissions: none by default.
resource "aws_iam_role" "task" {
  name               = "laurel-${var.environment}-${var.name}-task"
  assume_role_policy = data.aws_iam_policy_document.ecs_tasks.json
}

# --- ECS -------------------------------------------------------------------------

resource "aws_ecs_task_definition" "this" {
  family                   = "laurel-${var.environment}-${var.name}"
  requires_compatibilities = ["FARGATE"]
  network_mode             = "awsvpc"
  cpu                      = var.config.cpu
  memory                   = var.config.memory
  execution_role_arn       = aws_iam_role.execution.arn
  task_role_arn            = aws_iam_role.task.arn
  runtime_platform {
    cpu_architecture        = "X86_64"
    operating_system_family = "LINUX"
  }

  container_definitions = jsonencode([{
    name      = "app"
    image     = aws_ssm_parameter.image.insecure_value
    essential = true
    portMappings = [{
      name          = "http"
      containerPort = 8080
      protocol      = "tcp"
      appProtocol   = "http2" # h2c between services
    }]
    environment = [for k in sort(keys(local.environment)) : { name = k, value = local.environment[k] }]
    secrets     = [for k in sort(tolist(local.secrets)) : { name = k, valueFrom = aws_secretsmanager_secret.declared[k].arn }]
    stopTimeout = 30 # libraries/go/server drains in 25s
    logConfiguration = {
      logDriver = "awslogs"
      options = {
        awslogs-group         = aws_cloudwatch_log_group.this.name
        awslogs-region        = data.aws_region.current.region
        awslogs-stream-prefix = "app"
      }
    }
  }])
}

resource "aws_ecs_service" "this" {
  name            = var.name
  cluster         = var.cluster_arn
  task_definition = aws_ecs_task_definition.this.arn
  desired_count   = var.config.count
  launch_type     = "FARGATE"

  network_configuration {
    subnets          = var.subnet_ids
    security_groups  = var.security_group_ids
    assign_public_ip = false
  }

  service_connect_configuration {
    enabled   = true
    namespace = var.namespace_arn
    service {
      port_name      = "http"
      discovery_name = var.name
      client_alias {
        port     = 8080
        dns_name = var.name
      }
    }
  }

  dynamic "load_balancer" {
    for_each = var.target_group_arn == null ? [] : [var.target_group_arn]
    content {
      target_group_arn = load_balancer.value
      container_name   = "app"
      container_port   = 8080
    }
  }

  # A rollout whose tasks don't become healthy is rolled back automatically.
  deployment_circuit_breaker {
    enable   = true
    rollback = true
  }
  deployment_minimum_healthy_percent = 100
  deployment_maximum_percent         = 200

  lifecycle {
    # Autoscaling owns the count after creation.
    ignore_changes = [desired_count]
  }
}

resource "aws_appautoscaling_target" "this" {
  service_namespace  = "ecs"
  scalable_dimension = "ecs:service:DesiredCount"
  resource_id        = "service/${split("/", var.cluster_arn)[1]}/${aws_ecs_service.this.name}"
  min_capacity       = var.config.count
  max_capacity       = var.config.count * 4
}

resource "aws_appautoscaling_policy" "cpu" {
  name               = "cpu"
  policy_type        = "TargetTrackingScaling"
  service_namespace  = aws_appautoscaling_target.this.service_namespace
  scalable_dimension = aws_appautoscaling_target.this.scalable_dimension
  resource_id        = aws_appautoscaling_target.this.resource_id
  target_tracking_scaling_policy_configuration {
    target_value = 60
    predefined_metric_specification {
      predefined_metric_type = "ECSServiceAverageCPUUtilization"
    }
  }
}

output "task_role_arn" {
  value = aws_iam_role.task.arn
}

output "execution_role_arn" {
  value = aws_iam_role.execution.arn
}

output "secret_names" {
  description = "Secrets to fill in (out of band) before the first deploy."
  value       = [for s in aws_secretsmanager_secret.declared : s.name]
}
