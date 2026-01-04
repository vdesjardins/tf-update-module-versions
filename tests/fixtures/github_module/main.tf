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
  source = "github.com/hashicorp/example"
  name   = "github-module"
}

output "example_name" {
  value = module.example.name
}
