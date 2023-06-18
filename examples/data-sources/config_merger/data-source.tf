data "config_merger" "example" {
  config_path = "config/production/us-west-2/s3bucket"
}

locals {
  output = yamldecode(data.config_merger.example["result"])
}
