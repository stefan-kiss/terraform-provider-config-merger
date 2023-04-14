provider "configmerger" {
  project_config = "config/{{facts.environment}}/{{facts.region}}/{{facts.project}}"
}
