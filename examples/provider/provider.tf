terraform {
  required_providers {
    config = {
      source         = "registry.terraform.io/stefan-kiss/config"
      version        = "=1.0.0"
    }
  }
}

provider "config" {
  project_config = "config/{{facts.environment}}/{{facts.region}}/{{facts.project}}"
  config_globs = [
    "config.yaml",
    "*.config.yaml",
  ]
}
