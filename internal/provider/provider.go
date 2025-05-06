package provider

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/scayle/terraform-provider-pingdom/internal/api"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ provider.Provider = &pingdomProvider{}
)

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &pingdomProvider{
			version: version,
		}
	}
}

type pingdomProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

type pingdomProviderModel struct {
	ApiToken types.String `tfsdk:"api_token"`
}

func (p *pingdomProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "pingdom"
	resp.Version = p.version
}

func (p *pingdomProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"api_token": schema.StringAttribute{
				Required:  true,
				Sensitive: true,
				MarkdownDescription: `API token used to authenticate against the Pingdom API.

The API token needs to have Read/Write permissions.

See the Pingdom API documentation for more information: https://docs.pingdom.com/api/#section/Authentication.,
`,
			},
		},
	}
}

func (p *pingdomProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	// Check environment variables
	apiToken := os.Getenv("PINGDOM_API_TOKEN")

	var config pingdomProviderModel

	// Read configuration data into model
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)

	// Check configuration data, which should take precedence over
	// environment variable data, if found.
	if config.ApiToken.ValueString() != "" {
		apiToken = config.ApiToken.ValueString()
	}

	// If any of the expected configurations are missing, return
	// errors with provider-specific guidance.
	if apiToken == "" {
		resp.Diagnostics.AddError(
			"Missing API Token Configuration",
			"While configuring the provider, the API token was not found in "+
				"the PINGDOM_API_TOKEN environment variable or provider "+
				"configuration block api_token attribute.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	client := api.New(apiToken)
	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *pingdomProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewContactDataSource,
	}
}

func (p *pingdomProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewHTTPCheckResource,
	}
}
