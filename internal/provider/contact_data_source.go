package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/scayle/terraform-provider-pingdom/internal/api"
	"strconv"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &ContactDataSource{}

func NewContactDataSource() datasource.DataSource {
	return &ContactDataSource{}
}

type ContactDataSource struct {
	client api.Client
}

type ContactDataSourceModel struct {
	Name types.String `tfsdk:"name"`
	Id   types.String `tfsdk:"id"`
}

func (d *ContactDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_contact"
}

func (d *ContactDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Contact data source",

		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the contact",
				Required:            true,
			},
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The ID of the contact",
			},
		},
	}
}

func (d *ContactDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(api.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected api.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.client = client
}

func (d *ContactDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ContactDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	res, err := d.client.GetContacts(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read contacts, got error: %s", err))
		return
	}

	tflog.Debug(ctx, "Received contacts", map[string]interface{}{"contacts": res})

	for _, contact := range res.Contacts {
		if contact.Name == data.Name.ValueString() {
			tflog.Info(ctx, "Contact found", map[string]interface{}{
				"contact.name": contact.Name,
				"contact.id":   strconv.FormatInt(contact.Id, 10),
			})
			data.Id = types.StringValue(strconv.FormatInt(contact.Id, 10))
			break
		}
	}

	if data.Id.IsNull() {
		resp.Diagnostics.AddError("Unable to find contact", fmt.Sprintf("Unable to find contact with name: %s", data.Name))
		return
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
