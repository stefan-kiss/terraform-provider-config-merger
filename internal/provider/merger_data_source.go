package provider

import (
	"bytes"
	"context"
	"fmt"
	"github.com/gookit/goutil/maputil"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/k0kubun/pp"
	"github.com/stefan-kiss/terraform-provider-config-merger/pkg/envfacts"
	"github.com/stefan-kiss/terraform-provider-config-merger/pkg/finder"
	"github.com/stefan-kiss/terraform-provider-config-merger/pkg/merger"
	"gopkg.in/yaml.v3"
	"io"
	"os"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &MergerDataSource{}

func NewMergerDataSource() datasource.DataSource {
	return &MergerDataSource{}
}

// MergerDataSource defines the data source implementation.
type MergerDataSource struct {
	projectConfig string
	configGlobs   []string
}

// MergerDataSourceModel describes the data source data model.
type MergerDataSourceModel struct {
	Id         types.String `tfsdk:"id"`
	ConfigPath types.String `tfsdk:"config_path"`
	Result     types.String `tfsdk:"result"`
}

func (d *MergerDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_result"
}

func (d *MergerDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Merged data source",
		Attributes: map[string]schema.Attribute{
			// https://github.com/hashicorp/terraform-plugin-testing/issues/84
			"id": schema.StringAttribute{
				MarkdownDescription: "Example identifier",
				Computed:            true,
			},
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
		},
	}
}

type C struct {
	ProjectConfig basetypes.StringValue
	ConfigGlobs   []basetypes.StringValue
}

func (d *MergerDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}
	tflog.Trace(ctx, pp.Sprintln(req.ProviderData))

	providerConfig, ok := req.ProviderData.(ConfigMergerProviderModel)
	tflog.Trace(ctx, pp.Sprintln(providerConfig))
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected ConfigMergerProviderModel, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.projectConfig = providerConfig.ProjectConfig.ValueString()
	d.configGlobs = make([]string, len(providerConfig.ConfigGlobs))
	for i, v := range providerConfig.ConfigGlobs {
		d.configGlobs[i] = v.ValueString()
	}
	tflog.Trace(ctx, pp.Sprintln(d.configGlobs))
}

func (d *MergerDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data MergerDataSourceModel

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
	outMap := make(map[string]interface{}, 0)
	tflog.Trace(ctx, pp.Sprintln(outMap))
	for _, v := range p.Vars {
		pathKeys := strings.TrimPrefix(v.VariableName, ".")
		err := maputil.SetByPath(&outMap, pathKeys, v.VariableValue)
		if err != nil {
			resp.Diagnostics.AddError("Client Error: ", fmt.Sprintf("Unable to add key( %s ): %q", v.VariableName, err))
			return
		}
	}
	tflog.Trace(ctx, pp.Sprintln(outMap))
	out, err := yaml.Marshal(&outMap)
	if err != nil {
		resp.Diagnostics.AddError("Client Error: ", fmt.Sprintf("Unable Marshal output, got error: %s", err))
		return
	}

	mergeFileNames, err := finder.FindConfigFiles(p, d.configGlobs)
	if err != nil {
		resp.Diagnostics.AddError("Client Error: ", fmt.Sprintf("Unable FindConfigFiles, got error: %s", err))
		return
	}
	yamlFiles := make([]merger.YamlFile, 0)

	for _, filePath := range mergeFileNames {
		y, err := merger.LoadYamlFile(filePath)
		if err != nil {
			resp.Diagnostics.AddError("Client Error: ", fmt.Sprintf("Unable to LoadYamlFile, got error: %s", err))
			return
		}

		yamlFiles = append(yamlFiles, y)
	}

	yamlFiles = append(yamlFiles, merger.YamlFile{
		Path:   "facts.yaml",
		Reader: io.NopCloser(bytes.NewReader(out)),
	})

	ev, err := merger.MergeAllDocs(yamlFiles, merger.MergeOpts{})
	if err != nil {
		resp.Diagnostics.AddError("Client Error: ", fmt.Sprintf("Unable merger.MergeAllDocs, got error: %s", err))
		return
	}
	merged, err := yaml.Marshal(ev.Tree)
	if err != nil {
		resp.Diagnostics.AddError("Client Error: ", fmt.Sprintf("Unable yaml.Marshal, got error: %s", err))
		return
	}

	data.Result = types.StringValue(string(merged))
	// https://developer.hashicorp.com/terraform/plugin/framework/acctests#implement-id-attribute
	// We also need to set this (should be a hash)
	data.Id = types.StringValue(string(out))

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Trace(ctx, "read a data source")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
