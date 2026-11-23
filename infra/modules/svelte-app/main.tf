# One SvelteKit app (adapter-static SPA) in one environment: a private S3
# bucket behind CloudFront at https://<subdomain>.<domain>.

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

variable "domain_name" {
  description = "Full host name, e.g. app.staging.example.com."
  type        = string
}

variable "certificate_arn" {
  description = "ACM certificate in us-east-1 covering domain_name (CloudFront requirement)."
  type        = string
}

variable "zone_id" {
  description = "Route 53 zone to create the DNS record in (optional)."
  type        = string
  default     = null
}

locals {
  prefix = "/laurel/${var.environment}"
}

resource "aws_s3_bucket" "this" {
  bucket_prefix = "laurel-${var.environment}-${var.name}-"
}

resource "aws_s3_bucket_public_access_block" "this" {
  bucket                  = aws_s3_bucket.this.id
  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

resource "aws_s3_bucket_lifecycle_configuration" "this" {
  bucket = aws_s3_bucket.this.id
  # Deploys never delete hashed assets (old pages may still request them);
  # expire them here instead.
  rule {
    id     = "old-immutable-assets"
    status = "Enabled"
    filter {
      prefix = "_app/immutable/"
    }
    expiration {
      days = 30
    }
  }
}

resource "aws_cloudfront_origin_access_control" "this" {
  name                              = "laurel-${var.environment}-${var.name}"
  origin_access_control_origin_type = "s3"
  signing_behavior                  = "always"
  signing_protocol                  = "sigv4"
}

data "aws_cloudfront_cache_policy" "optimized" {
  name = "Managed-CachingOptimized"
}

resource "aws_cloudfront_distribution" "this" {
  enabled             = true
  aliases             = [var.domain_name]
  default_root_object = "index.html"
  http_version        = "http2and3"
  price_class         = "PriceClass_100"

  origin {
    origin_id                = "s3"
    domain_name              = aws_s3_bucket.this.bucket_regional_domain_name
    origin_access_control_id = aws_cloudfront_origin_access_control.this.id
  }

  default_cache_behavior {
    target_origin_id       = "s3"
    viewer_protocol_policy = "redirect-to-https"
    allowed_methods        = ["GET", "HEAD"]
    cached_methods         = ["GET", "HEAD"]
    compress               = true
    # Honors the Cache-Control set by deploy-svelte.sh.
    cache_policy_id = data.aws_cloudfront_cache_policy.optimized.id
  }

  # SPA fallback: unknown paths serve the app shell.
  custom_error_response {
    error_code         = 403
    response_code      = 200
    response_page_path = "/index.html"
  }
  custom_error_response {
    error_code         = 404
    response_code      = 200
    response_page_path = "/index.html"
  }

  restrictions {
    geo_restriction {
      restriction_type = "none"
    }
  }

  viewer_certificate {
    acm_certificate_arn      = var.certificate_arn
    ssl_support_method       = "sni-only"
    minimum_protocol_version = "TLSv1.2_2021"
  }
}

resource "aws_s3_bucket_policy" "this" {
  bucket = aws_s3_bucket.this.id
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect    = "Allow"
      Principal = { Service = "cloudfront.amazonaws.com" }
      Action    = "s3:GetObject"
      Resource  = "${aws_s3_bucket.this.arn}/*"
      Condition = { StringEquals = { "AWS:SourceArn" = aws_cloudfront_distribution.this.arn } }
    }]
  })
}

resource "aws_route53_record" "this" {
  count   = var.zone_id == null ? 0 : 1
  zone_id = var.zone_id
  name    = var.domain_name
  type    = "A"
  alias {
    name                   = aws_cloudfront_distribution.this.domain_name
    zone_id                = aws_cloudfront_distribution.this.hosted_zone_id
    evaluate_target_health = false
  }
}

# Read by .github/scripts/deploy-svelte.sh.
resource "aws_ssm_parameter" "bucket" {
  name  = "${local.prefix}/apps/${var.name}/bucket"
  type  = "String"
  value = aws_s3_bucket.this.bucket
}

resource "aws_ssm_parameter" "distribution" {
  name  = "${local.prefix}/apps/${var.name}/distribution"
  type  = "String"
  value = aws_cloudfront_distribution.this.id
}

output "bucket_arn" {
  value = aws_s3_bucket.this.arn
}

output "distribution_arn" {
  value = aws_cloudfront_distribution.this.arn
}

output "origin" {
  value = "https://${var.domain_name}"
}
