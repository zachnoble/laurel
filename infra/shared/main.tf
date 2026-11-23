# Resources shared by every environment: one ECR repository per Go service
# (images are built once and promoted from staging to production), GitHub's
# OIDC provider, and the role CI uses to push images.
#
# Apply this before the environments: `terraform -chdir=infra/shared apply`.

terraform {
  required_version = ">= 1.9"
  required_providers {
    aws = { source = "hashicorp/aws", version = "~> 6.0" }
  }
  # backend "s3" { bucket = "<state bucket>", key = "laurel/shared.tfstate", region = "<region>" }
}

variable "region" {
  type = string
}

variable "github_repository" {
  description = "owner/name of this repository, trusted by the CI roles."
  type        = string
}

provider "aws" {
  region = var.region
  default_tags {
    tags = { project = "laurel", stack = "shared" }
  }
}

locals {
  projects = jsondecode(file("${path.module}/../projects.json"))
}

resource "aws_ecr_repository" "service" {
  for_each             = local.projects.services
  name                 = "laurel/${each.key}"
  image_tag_mutability = "IMMUTABLE" # tags are commit SHAs
  image_scanning_configuration {
    scan_on_push = true
  }
}

resource "aws_ecr_lifecycle_policy" "service" {
  for_each   = aws_ecr_repository.service
  repository = each.value.name
  # Old images stay available for rollbacks, up to a point.
  policy = jsonencode({
    rules = [{
      rulePriority = 1
      description  = "Keep the last 200 images"
      selection    = { tagStatus = "any", countType = "imageCountMoreThan", countNumber = 200 }
      action       = { type = "expire" }
    }]
  })
}

resource "aws_iam_openid_connect_provider" "github" {
  url            = "https://token.actions.githubusercontent.com"
  client_id_list = ["sts.amazonaws.com"]
}

# Pushes images from pushes to main (the Deploy workflow's images job).
resource "aws_iam_role" "build" {
  name = "laurel-ci-build"
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect    = "Allow"
      Principal = { Federated = aws_iam_openid_connect_provider.github.arn }
      Action    = "sts:AssumeRoleWithWebIdentity"
      Condition = {
        StringEquals = {
          "token.actions.githubusercontent.com:aud" = "sts.amazonaws.com"
          "token.actions.githubusercontent.com:sub" = "repo:${var.github_repository}:ref:refs/heads/main"
        }
      }
    }]
  })
}

resource "aws_iam_role_policy" "build" {
  role = aws_iam_role.build.id
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      { Effect = "Allow", Action = ["ecr:GetAuthorizationToken"], Resource = "*" },
      {
        Effect = "Allow"
        Action = [
          "ecr:BatchCheckLayerAvailability", "ecr:BatchGetImage", "ecr:CompleteLayerUpload",
          "ecr:DescribeImages", "ecr:DescribeRepositories", "ecr:GetDownloadUrlForLayer",
          "ecr:InitiateLayerUpload", "ecr:PutImage", "ecr:UploadLayerPart",
        ]
        Resource = [for r in aws_ecr_repository.service : r.arn]
      },
    ]
  })
}

output "build_role_arn" {
  description = "Set as the repository variable AWS_BUILD_ROLE_ARN."
  value       = aws_iam_role.build.arn
}

output "oidc_provider_arn" {
  value = aws_iam_openid_connect_provider.github.arn
}
