terraform {
  required_version = ">= 1.9"
  required_providers {
    aws = { source = "hashicorp/aws", version = "~> 6.0" }
  }
  # backend "s3" { bucket = "<state bucket>", key = "laurel/staging.tfstate", region = "<region>", use_lockfile = true }
}

provider "aws" {
  region = var.region
  default_tags {
    tags = { project = "laurel", environment = "staging" }
  }
}

variable "region" {
  type = string
}

module "environment" {
  source = "../../modules/environment"

  name                 = "staging"
  github_repository    = var.github_repository
  vpc_id               = var.vpc_id
  private_subnet_ids   = var.private_subnet_ids
  public_subnet_ids    = var.public_subnet_ids
  domain               = var.domain
  zone_id              = var.zone_id
  api_certificate_arn  = var.api_certificate_arn
  apps_certificate_arn = var.apps_certificate_arn
  identity_public_keys = var.identity_public_keys
  service_env          = var.service_env
}

variable "github_repository" {
  type = string
}

variable "vpc_id" {
  type = string
}

variable "private_subnet_ids" {
  type = list(string)
}

variable "public_subnet_ids" {
  type = list(string)
}

variable "domain" {
  type = string
}

variable "zone_id" {
  type    = string
  default = null
}

variable "api_certificate_arn" {
  type = string
}

variable "apps_certificate_arn" {
  type = string
}

variable "identity_public_keys" {
  type = list(string)
}

variable "service_env" {
  type    = map(map(string))
  default = {}
}

output "deploy_role_arn" {
  value = module.environment.deploy_role_arn
}

output "secrets_to_set" {
  value = module.environment.secrets_to_set
}
