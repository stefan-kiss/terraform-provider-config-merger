data "merger_result" "example" {
  config_path = "config/production/us-west-2/s3bucket"
}

locals {
  output = yamldecode(data.merger_result.example["result"])
}
