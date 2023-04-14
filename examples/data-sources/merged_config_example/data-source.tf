data "configmerger_merged" "example" {
  config_path = "config/production/us-west-2/s3bucket"
}

locals {
  output = yamldecode(data.configmerger_merged.example["result"])
}
