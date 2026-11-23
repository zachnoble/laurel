# A whole environment (staging, production): the shared platform plus every
# service and app from infra/projects.json. Adding a project to the repo adds
# it here on the next apply; nothing in infra lists projects by hand.

terraform {
  required_providers {
    aws = { source = "hashicorp/aws", version = "~> 6.0" }
  }
}

variable "name" {
  description = "Environment name: staging or production."
  type        = string
}

variable "github_repository" {
  description = "owner/name; the deploy role trusts this repo's \"<name>\" GitHub environment."
  type        = string
}

variable "vpc_id" {
  type = string
}

variable "private_subnet_ids" {
  description = "For tasks. Needs a NAT gateway or VPC endpoints for ECR, SSM, Secrets Manager, Logs."
  type        = list(string)
}

variable "public_subnet_ids" {
  description = "For the load balancer."
  type        = list(string)
}

variable "domain" {
  description = "The environment's domain: the API is api.<domain>, apps are <subdomain>.<domain>. Apps and API must share it so the session cookie (SameSite=Lax) is same-site."
  type        = string
}

variable "zone_id" {
  description = "Route 53 zone for domain (optional; otherwise create the DNS records yourself)."
  type        = string
  default     = null
}

variable "api_certificate_arn" {
  description = "ACM certificate (in this region) for api.<domain>."
  type        = string
}

variable "apps_certificate_arn" {
  description = "ACM certificate in us-east-1 for *.<domain> (CloudFront)."
  type        = string
}

variable "identity_public_keys" {
  description = "Public keys for user identity tokens (`just keygen`); auth's IDENTITY_PRIVATE_KEY secret holds the private half. List two while rotating."
  type        = list(string)
}

variable "service_env" {
  description = "Per-service environment for this environment, e.g. { auth = { EMAIL_FROM_ADDRESS = \"...\" } }."
  type        = map(map(string))
  default     = {}
}

variable "otel_exporter_endpoint" {
  description = "OTLP/HTTP endpoint for traces and metrics (e.g. an ADOT collector). Empty disables export."
  type        = string
  default     = ""
}

variable "log_retention_days" {
  type    = number
  default = 30
}

locals {
  projects = jsondecode(file("${path.module}/../../projects.json"))
  services = local.projects.services
  apps     = local.projects.apps
  prefix   = "/laurel/${var.name}"

  public_services = [for n, s in local.services : n if s.public]
  api_host        = "api.${var.domain}"
  app_origins     = [for n, a in local.apps : "https://${a.subdomain}.${var.domain}"]
}

data "aws_region" "current" {}
data "aws_caller_identity" "current" {}

check "one_public_service" {
  assert {
    condition     = length(local.public_services) == 1
    error_message = "Exactly one service (the gateway) may set deploy.public in its project.yaml."
  }
}

# --- Cluster and networking --------------------------------------------------------

resource "aws_ecs_cluster" "this" {
  name = "laurel-${var.name}"
  setting {
    name  = "containerInsights"
    value = "enhanced"
  }
}

resource "aws_service_discovery_http_namespace" "this" {
  name = "laurel-${var.name}"
}

resource "aws_security_group" "services" {
  name        = "laurel-${var.name}-services"
  description = "Every service: reachable on 8080 from other services only."
  vpc_id      = var.vpc_id
}

resource "aws_vpc_security_group_ingress_rule" "services_from_services" {
  security_group_id            = aws_security_group.services.id
  referenced_security_group_id = aws_security_group.services.id
  ip_protocol                  = "tcp"
  from_port                    = 8080
  to_port                      = 8080
}

resource "aws_vpc_security_group_egress_rule" "services_out" {
  security_group_id = aws_security_group.services.id
  ip_protocol       = "-1"
  cidr_ipv4         = "0.0.0.0/0"
}

resource "aws_security_group" "alb" {
  name        = "laurel-${var.name}-alb"
  description = "Public load balancer."
  vpc_id      = var.vpc_id
}

resource "aws_vpc_security_group_ingress_rule" "alb_https" {
  security_group_id = aws_security_group.alb.id
  ip_protocol       = "tcp"
  from_port         = 443
  to_port           = 443
  cidr_ipv4         = "0.0.0.0/0"
}

resource "aws_vpc_security_group_ingress_rule" "alb_http" {
  security_group_id = aws_security_group.alb.id
  ip_protocol       = "tcp"
  from_port         = 80
  to_port           = 80
  cidr_ipv4         = "0.0.0.0/0"
}

resource "aws_vpc_security_group_egress_rule" "alb_to_edge" {
  security_group_id            = aws_security_group.alb.id
  referenced_security_group_id = aws_security_group.edge.id
  ip_protocol                  = "tcp"
  from_port                    = 8080
  to_port                      = 8080
}

# Only the public service gets this group, so the load balancer can reach the
# gateway and nothing else.
resource "aws_security_group" "edge" {
  name        = "laurel-${var.name}-edge"
  description = "The public service: reachable from the load balancer."
  vpc_id      = var.vpc_id
}

resource "aws_vpc_security_group_ingress_rule" "edge_from_alb" {
  security_group_id            = aws_security_group.edge.id
  referenced_security_group_id = aws_security_group.alb.id
  ip_protocol                  = "tcp"
  from_port                    = 8080
  to_port                      = 8080
}

# --- Load balancer (gateway only) ----------------------------------------------------

resource "aws_lb" "this" {
  name               = "laurel-${var.name}"
  load_balancer_type = "application"
  subnets            = var.public_subnet_ids
  security_groups    = [aws_security_group.alb.id]
  idle_timeout       = 60 # libraries/go/server's IdleTimeout (120s) is longer
}

resource "aws_lb_target_group" "edge" {
  name             = "laurel-${var.name}-edge"
  port             = 8080
  protocol         = "HTTP"
  protocol_version = "HTTP1"
  target_type      = "ip"
  vpc_id           = var.vpc_id
  health_check {
    path    = "/readiness"
    matcher = "200"
  }
  deregistration_delay = 30
}

resource "aws_lb_listener" "https" {
  load_balancer_arn = aws_lb.this.arn
  port              = 443
  protocol          = "HTTPS"
  ssl_policy        = "ELBSecurityPolicy-TLS13-1-2-2021-06"
  certificate_arn   = var.api_certificate_arn
  default_action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.edge.arn
  }
}

resource "aws_lb_listener" "http" {
  load_balancer_arn = aws_lb.this.arn
  port              = 80
  protocol          = "HTTP"
  default_action {
    type = "redirect"
    redirect {
      port        = "443"
      protocol    = "HTTPS"
      status_code = "HTTP_301"
    }
  }
}

resource "aws_route53_record" "api" {
  count   = var.zone_id == null ? 0 : 1
  zone_id = var.zone_id
  name    = local.api_host
  type    = "A"
  alias {
    name                   = aws_lb.this.dns_name
    zone_id                = aws_lb.this.zone_id
    evaluate_target_health = true
  }
}

# --- Projects ------------------------------------------------------------------------

data "aws_ecr_repository" "service" {
  for_each = local.services
  name     = "laurel/${each.key}"
}

module "service" {
  source   = "../go-service"
  for_each = local.services

  name                   = each.key
  environment            = var.name
  config                 = each.value
  cluster_arn            = aws_ecs_cluster.this.arn
  namespace_arn          = aws_service_discovery_http_namespace.this.arn
  subnet_ids             = var.private_subnet_ids
  security_group_ids     = each.value.public ? [aws_security_group.services.id, aws_security_group.edge.id] : [aws_security_group.services.id]
  target_group_arn       = each.value.public ? aws_lb_target_group.edge.arn : null
  repository_url         = data.aws_ecr_repository.service[each.key].repository_url
  identity_public_keys   = var.identity_public_keys
  otel_exporter_endpoint = var.otel_exporter_endpoint
  log_retention_days     = var.log_retention_days
  extra_env = merge(
    # The gateway allows exactly this environment's apps.
    each.value.public ? { ALLOWED_ORIGINS = join(",", local.app_origins) } : {},
    lookup(var.service_env, each.key, {}),
  )
}

module "app" {
  source   = "../svelte-app"
  for_each = local.apps

  name            = each.key
  environment     = var.name
  domain_name     = "${each.value.subdomain}.${var.domain}"
  certificate_arn = var.apps_certificate_arn
  zone_id         = var.zone_id
}

# --- What deploys read (.github/scripts) -----------------------------------------------

resource "aws_ssm_parameter" "cluster" {
  name  = "${local.prefix}/cluster"
  type  = "String"
  value = aws_ecs_cluster.this.name
}

resource "aws_ssm_parameter" "network" {
  name = "${local.prefix}/network"
  type = "String"
  # For one-off migration tasks (aws ecs run-task --network-configuration).
  value = jsonencode({
    awsvpcConfiguration = {
      subnets        = var.private_subnet_ids
      securityGroups = [aws_security_group.services.id]
      assignPublicIp = "DISABLED"
    }
  })
}

resource "aws_ssm_parameter" "api_url" {
  name  = "${local.prefix}/api_url"
  type  = "String"
  value = "https://${local.api_host}"
}

# --- CI deploy role ----------------------------------------------------------------------

data "aws_iam_openid_connect_provider" "github" {
  url = "https://token.actions.githubusercontent.com"
}

resource "aws_iam_role" "deploy" {
  name = "laurel-${var.name}-deploy"
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect    = "Allow"
      Principal = { Federated = data.aws_iam_openid_connect_provider.github.arn }
      Action    = "sts:AssumeRoleWithWebIdentity"
      Condition = {
        StringEquals = {
          "token.actions.githubusercontent.com:aud" = "sts.amazonaws.com"
          # Only jobs running in this GitHub environment (with its protection rules).
          "token.actions.githubusercontent.com:sub" = "repo:${var.github_repository}:environment:${var.name}"
        }
      }
    }]
  })
}

resource "aws_iam_role_policy" "deploy" {
  role = aws_iam_role.deploy.id
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect   = "Allow"
        Action   = ["ssm:GetParameter", "ssm:PutParameter"]
        Resource = "arn:aws:ssm:${data.aws_region.current.region}:${data.aws_caller_identity.current.account_id}:parameter${local.prefix}/*"
      },
      {
        Effect = "Allow"
        Action = [
          "ecs:DescribeServices", "ecs:UpdateService", "ecs:DescribeTaskDefinition", "ecs:RegisterTaskDefinition",
          "ecs:RunTask", "ecs:DescribeTasks",
        ]
        Resource = "*"
        Condition = {
          StringEqualsIfExists = { "ecs:cluster" = aws_ecs_cluster.this.arn }
        }
      },
      {
        Effect   = "Allow"
        Action   = ["iam:PassRole"]
        Resource = flatten([for m in module.service : [m.task_role_arn, m.execution_role_arn]])
      },
      { Effect = "Allow", Action = ["ecr:GetAuthorizationToken"], Resource = "*" },
      {
        # Describe images, and push one if a deploy finds it missing.
        Effect = "Allow"
        Action = [
          "ecr:BatchCheckLayerAvailability", "ecr:BatchGetImage", "ecr:CompleteLayerUpload",
          "ecr:DescribeImages", "ecr:DescribeRepositories", "ecr:GetDownloadUrlForLayer",
          "ecr:InitiateLayerUpload", "ecr:PutImage", "ecr:UploadLayerPart",
        ]
        Resource = [for r in data.aws_ecr_repository.service : r.arn]
      },
      {
        Effect   = "Allow"
        Action   = ["s3:ListBucket", "s3:GetObject", "s3:PutObject", "s3:DeleteObject"]
        Resource = flatten([for m in module.app : [m.bucket_arn, "${m.bucket_arn}/*"]])
      },
      {
        Effect   = "Allow"
        Action   = ["cloudfront:CreateInvalidation"]
        Resource = [for m in module.app : m.distribution_arn]
      },
    ]
  })
}

output "deploy_role_arn" {
  description = "Set as the AWS_DEPLOY_ROLE_ARN variable of this GitHub environment."
  value       = aws_iam_role.deploy.arn
}

output "load_balancer_dns_name" {
  value = aws_lb.this.dns_name
}

output "secrets_to_set" {
  description = "Secrets Manager secrets whose values must be set before the first deploy."
  value       = flatten([for m in module.service : m.secret_names])
}
