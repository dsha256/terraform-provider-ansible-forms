package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"terraform-provider-ansible-forms/internal/interfaces"
	"terraform-provider-ansible-forms/internal/utils"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ datasource.DataSource = &JobDataSource{}

// JobDataSource defines the data source implementation.
type JobDataSource struct {
	config resourceOrDataSourceConfig
}

// NewJobDataSource is a helper function to simplify the provider implementation.
func NewJobDataSource() datasource.DataSource {
	return &JobDataSource{
		config: resourceOrDataSourceConfig{
			name: "job_data_source",
		},
	}
}

// JobDataSourceModel maps the resource schema data.
type JobDataSourceModel struct {
	CxProfileName types.String `tfsdk:"cx_profile_name"`
	ID            types.Int64  `tfsdk:"id"`
	LastUpdated   types.String `tfsdk:"last_updated"`
	FormName      types.String `tfsdk:"form_name"`
	Status        types.String `tfsdk:"status"`
	Extravars     types.Map    `tfsdk:"extravars"`
	Credentials   types.Map    `tfsdk:"credentials"`
	Target        types.String `tfsdk:"target"`
	Output        types.String `tfsdk:"output"`
	Counter       types.Int64  `tfsdk:"counter"`
	NoOfRecords   types.Int64  `tfsdk:"no_of_records"`
	Start         types.String `tfsdk:"start"`
	End           types.String `tfsdk:"end"`
	Approval      types.String `tfsdk:"approval"`
}

// Metadata returns the data source type name.
func (d *JobDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + d.config.name
}

// Schema defines the schema for the data source.
func (d *JobDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Job JobGetDataSourceModel Source",

		Attributes: map[string]schema.Attribute{
			"cx_profile_name": schema.StringAttribute{
				MarkdownDescription: "Connection profile name",
				Required:            true,
			},
			"id": schema.Int64Attribute{
				MarkdownDescription: "",
				Required:            true,
			},
			"last_updated": schema.StringAttribute{
				MarkdownDescription: "",
				Computed:            true,
			},
			"form_name": schema.StringAttribute{
				Description: "Form Name.",
				Computed:    true,
			},
			"status": schema.StringAttribute{
				MarkdownDescription: "",
				Computed:            true,
			},
			"extravars": schema.MapAttribute{
				MarkdownDescription: "",
				ElementType:         types.StringType,
				Computed:            true,
			},
			"credentials": schema.MapAttribute{
				MarkdownDescription: "",
				ElementType:         types.StringType,
				Computed:            true,
			},
			"target": schema.StringAttribute{
				Computed: true,

				MarkdownDescription: "Target form of a job.",
			},
			"output": schema.StringAttribute{
				Computed: true,

				MarkdownDescription: "Output of a job.",
			},
			"counter": schema.Int64Attribute{
				Computed: true,

				MarkdownDescription: "Counter of a job.",
			},
			"no_of_records": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "Number of records of a job.",
			},
			"start": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Start time of a job.",
			},
			"end": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "End time of a job.",
			},
			"approval": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Approval of a job.",
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *JobDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}
	config, ok := req.ProviderData.(Config)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected JobGetDataSourceModel Source Configure Type",
			fmt.Sprintf("Expected Config, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
	}
	d.config.providerConfig = config
}

// Read refreshes the Terraform state with the latest data.
func (d *JobDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data JobDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	errorHandler := utils.NewErrorHandler(ctx, &resp.Diagnostics)
	// we need to defer setting the client until we can read the connection profile name
	client, err := getRestClient(errorHandler, d.config, data.CxProfileName)
	if err != nil {
		// error reporting done inside NewClient
		return
	}

	restInfo, err := interfaces.GetJobByID(errorHandler, *client, data.ID.String())
	if err != nil {
		// error reporting done inside GetSVMPeer
		return
	}

	data.ID = types.Int64Value(restInfo.ID)
	data.FormName = types.StringValue(restInfo.Form)
	data.Status = types.StringValue(restInfo.Status)
	data.Extravars = jsonStringToMapValue(ctx, &resp.Diagnostics, restInfo.Extravars)
	data.Credentials = jsonStringToMapValue(ctx, &resp.Diagnostics, restInfo.Credentials)
	data.Target = types.StringValue(restInfo.Target)
	data.Output = types.StringValue(restInfo.Output)
	data.Counter = types.Int64Value(restInfo.Counter)
	data.NoOfRecords = types.Int64Value(restInfo.NoOfRecords)
	data.Start = types.StringValue(restInfo.Start)
	data.End = types.StringValue(restInfo.End)
	data.Approval = types.StringValue(restInfo.Approval)

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Debug(ctx, fmt.Sprintf("read a data source: %#v", data))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
