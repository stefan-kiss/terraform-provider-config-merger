package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/k0kubun/pp"
	"github.com/stefan-kiss/terraform-provider-config-merger/pkg/envfacts"
	"github.com/stefan-kiss/terraform-provider-config-merger/pkg/yutils"
	"gopkg.in/yaml.v3"
	"os"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &MergedDataSource{}

func NewMergedDataSource() datasource.DataSource {
	return &MergedDataSource{}
}

// MergedDataSource defines the data source implementation.
type MergedDataSource struct {
	projectConfig string
}

// MergedDataSourceModel describes the data source data model.
type MergedDataSourceModel struct {
	Id         types.String `tfsdk:"id"`
	ConfigPath types.String `tfsdk:"config_path"`
	Result     types.String `tfsdk:"result"`
}

func (d *MergedDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_merged"
}

func (d *MergedDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Merged data source",
		Attributes: map[string]schema.Attribute{
			"config_path": schema.StringAttribute{
				MarkdownDescription: "Path to the most specific configuration file",
				Required:            true,
			},
			"result": schema.StringAttribute{
				MarkdownDescription: "Path to the most specific configuration file",
				Required:            false,
				Optional:            false,
				Computed:            true,
			},
			// https://github.com/hashicorp/terraform-plugin-testing/issues/84
			"id": schema.StringAttribute{
				MarkdownDescription: "Example identifier",
				Computed:            true,
			},
		},
	}
}

func (d *MergedDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}
	tflog.Trace(ctx, pp.Sprintln(req.ProviderData))

	projectConfig, ok := req.ProviderData.(basetypes.StringValue)
	tflog.Trace(ctx, pp.Sprintln(projectConfig))

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected basetypes.StringValue, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}
	tflog.Trace(ctx, pp.Sprintln(projectConfig.ValueString()))

	d.projectConfig = projectConfig.ValueString()
}

func (d *MergedDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data MergedDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// If applicable, this is a great opportunity to initialize any necessary
	// provider client data and make a call using it.
	// httpResp, err := d.client.Do(httpReq)
	// if err != nil {
	//     resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read example, got error: %s", err))
	//     return
	// }

	// For the purposes of this example code, hardcoding a response value to
	// save into the Terraform state.

	p, err := envfacts.ParseProjectStructure(d.projectConfig)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable parse project, got error: %s", err))
		return

	}

	err = p.MapPathToProject(data.ConfigPath.ValueString(), os.UserHomeDir)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable parse config dir, got error: %s", err))
		return
	}

	tflog.Trace(ctx, pp.Sprintln(p))
	outNodes := yaml.Node{
		Kind: yaml.DocumentNode,
		Content: []*yaml.Node{
			{
				Kind:    yaml.MappingNode,
				Tag:     "!!map",
				Content: []*yaml.Node{},
			},
		},
	}
	tflog.Trace(ctx, pp.Sprintln(outNodes))
	for _, v := range p.Vars {
		pathKeys := strings.Split(v.VariableName, ".")
		if pathKeys[0] == "" {
			pathKeys = pathKeys[1:]
		}
		err := yutils.SetValueAtPath(pathKeys, v.VariableValue, &outNodes)
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to add key( %s ): %q", v.VariableName, err))
			return
		}
	}
	tflog.Trace(ctx, pp.Sprintln(outNodes))
	out, err := yaml.Marshal(&outNodes)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable Marshal output, got error: %s", err))
		return
	}
	data.Result = types.StringValue(string(out))
	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Trace(ctx, "read a data source")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
