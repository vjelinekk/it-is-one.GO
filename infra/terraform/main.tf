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
  default = "https://github.com/vjelinekk/it-is-one.GO.git"
}

variable "ses_from_email" {
  default = "vojtech.jelinekk@gmail.com"
}
