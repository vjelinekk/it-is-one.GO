terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
    tls = {
      source  = "hashicorp/tls"
      version = "~> 4.0"
    }
    local = {
      source  = "hashicorp/local"
      version = "~> 2.0"
    }
  }
}

provider "aws" {
  region = var.aws_region
}

variable "aws_region" {
  default = "eu-central-1"
}

variable "github_repo" {
  description = "GitHub repo URL to clone on EC2 boot (e.g. https://github.com/user/repo.git)"
  type        = string
}

variable "ses_from_email" {
  description = "Verified SES sender email address"
  type        = string
}
