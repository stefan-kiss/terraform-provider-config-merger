terraform {
  required_providers {
    config-merger = {
      source  = "registry.terraform.io/stefan-kiss/config-merger"
      version = "=1.0.0"
    }
  }
}

provider "merger" {
  project_config = "config/{{facts.environment}}/{{facts.region}}/{{facts.project}}"
  config_globs = [
    "config.yaml",
    "*.config.yaml",
  ]
}
