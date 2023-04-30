provider "configmerger" {
  project_config = "config/{{facts.environment}}/{{facts.region}}/{{facts.project}}"
  config_globs = [
    "config.yaml",
    "*.config.yml",
  ]
}
