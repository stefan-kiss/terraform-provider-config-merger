package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure ConfigMergerProvider satisfies various provider interfaces.
var _ provider.Provider = &ConfigMergerProvider{}

// ConfigMergerProvider defines the provider implementation.
type ConfigMergerProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// ConfigMergerProviderModel describes the provider data model.
type ConfigMergerProviderModel struct {
	ProjectConfig types.String   `tfsdk:"project_config"`
	ConfigGlobs   []types.String `tfsdk:"config_globs"`
}

func (p *ConfigMergerProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "config"
	resp.Version = p.version
}

func (p *ConfigMergerProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"project_config": schema.StringAttribute{
				MarkdownDescription: "Project Configuration",
				Required:            true,
			},
			"config_globs": schema.ListAttribute{
				ElementType:         types.StringType,
				Required:            true,
				Optional:            false,
				MarkdownDescription: "List of globs to search for config files. Only last segment of each glob is considered",
			},
		},
	}
}

func (p *ConfigMergerProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data ConfigMergerProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Configuration values are now available.
	// if data.Endpoint.IsNull() { /* ... */ }

	// Example client configuration for data sources and resources

	resp.DataSourceData = data

}

func (p *ConfigMergerProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{}
}

func (p *ConfigMergerProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewMergerDataSource,
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &ConfigMergerProvider{
			version: version,
		}
	}
}
