---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "config-merger_result Data Source - terraform-provider-config-merger"
subcategory: ""
description: |-
  Merged data source
---

# config-merger_result (Data Source)

Merged data source

## Example Usage

```terraform
data "config-merger_result" "example" {
  config_path = "config/production/us-west-2/s3bucket"
}

locals {
  output = yamldecode(data.config-merger_result["result"])
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `config_path` (String) Path to the most specific configuration file

### Read-Only

- `id` (String) Example identifier
- `result` (String) Path to the most specific configuration file
