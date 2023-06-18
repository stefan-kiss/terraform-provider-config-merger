package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccExampleDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,

		Steps: []resource.TestStep{
			// Read testing
			{
				Config: testAccExampleDataSourceConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.config_merger.test", "result", testAccExampleDataSourceResult),
				),
			},
		},
	})
}

const testAccExampleDataSourceConfig = `
provider "config" {
  project_config = "config/{{facts.environment}}/{{facts.region}}/{{facts.project}}"
  config_globs   = [
    "config.yaml",
    "*.config.yaml",
  ]
}
data "config_merger" "test" {
  config_path = "../../tests/config/production/us-west-2/s3bucket"
}
`

const testAccExampleDataSourceResult = `facts:
    environment: production
    project: s3bucket
    region: us-west-2
root_key:
    key_1: s3bucket_value_1
    key_2: production-s3bucket_value_2
    key_3: s3bucket_value_1
`
