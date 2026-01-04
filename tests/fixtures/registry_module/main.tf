terraform {
  required_version = ">= 1.0"
  required_providers {
    null = {
      source  = "hashicorp/null"
      version = "~> 3.0"
    }
  }
}

module "example" {
  source = "hashicorp/vault-starter/aws"
  version = "0.1.3"
}
