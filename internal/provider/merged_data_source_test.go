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
					resource.TestCheckResourceAttr("data.configmerger_merged.test", "result", testAccExampleDataSourceResult),
				),
			},
		},
	})
}

const testAccExampleDataSourceConfig = `
provider "configmerger" {
  project_config = "config/{{.globals.environment}}/{{.globals.region}}/{{.globals.project}}"
}
data "configmerger_merged" "test" {
  config_path = "/home/user/project/config/development/us-east-2/s3bucket"
}
`

const testAccExampleDataSourceResult = `globals:
    environment: development
    region: us-east-2
    project: s3bucket
`
