data "config-merger_result" "example" {
  config_path = "config/production/us-west-2/s3bucket"
}

locals {
  output = yamldecode(data.config-merger_result["result"])
}
