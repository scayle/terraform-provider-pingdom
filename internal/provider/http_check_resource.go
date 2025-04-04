package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/scayle/terraform-provider-pingdom/internal/api"
	api_types "github.com/scayle/terraform-provider-pingdom/internal/api/types"
	"strconv"
	"strings"
	"time"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &HTTPCheckResource{}
var _ resource.ResourceWithImportState = &HTTPCheckResource{}

func NewHTTPCheckResource() resource.Resource {
	return &HTTPCheckResource{}
}

type HTTPCheckResource struct {
	client api.Client
}

type HTTPCheckResourceModel struct {
	Id     types.String `tfsdk:"id"`
	Name   types.String `tfsdk:"name"`
	Paused types.Bool   `tfsdk:"paused"`

	Host types.String        `tfsdk:"host"`
	Url  types.String        `tfsdk:"url"`
	Auth *HTTPCheckAuthModel `tfsdk:"auth"`

	Frequency  types.String `tfsdk:"frequency"`
	Message    types.String `tfsdk:"message"`
	ContactIds types.Set    `tfsdk:"contact_ids"`
	// Triggers a down alert if the response time exceeds threshold specified in ms.
	ResponseTimeThreshold types.Int64 `tfsdk:"response_time_threshold"`
	// Send notification when down X times
	NotifyWhenDown types.Int64 `tfsdk:"notify_when_down"`
	// Notify again every n result. 0 means that no extra notifications will be sent.
	NotifyAgainEvery types.Int64 `tfsdk:"notify_again_every"`
	// Notify when back up again
	NotifyWhenBackUp types.Bool `tfsdk:"notify_when_back_up"`

	// SSL configs
	SSLDownDaysBefore types.Int64 `tfsdk:"ssl_down_days_before"`
	VerifyCertificate types.Bool  `tfsdk:"verify_certificate"`

	Regions types.Set `tfsdk:"regions"`

	Tags types.Map `tfsdk:"tags"`
}

type HTTPCheckAuthModel struct {
	Username types.String `tfsdk:"username"`
	Password types.String `tfsdk:"password"`
}

func (r *HTTPCheckResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_http_check"
}

func (r *HTTPCheckResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "CheckDetail resource",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The ID of the check in Pingdom.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the check.",
				Required:            true,
			},
			"paused": schema.BoolAttribute{
				MarkdownDescription: "Whether the check is paused.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},

			"host": schema.StringAttribute{
				MarkdownDescription: "The host of the check.",
				Required:            true,
			},
			"url": schema.StringAttribute{
				MarkdownDescription: "A specific URL to check against.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("/"),
			},
			"auth": schema.SingleNestedAttribute{
				MarkdownDescription: "Authentication configuration in case the host is protected by basic auth.",
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"username": schema.StringAttribute{
						MarkdownDescription: "The username for basic auth.",
						Required:            true,
					},
					"password": schema.StringAttribute{
						MarkdownDescription: "The password for basic auth.",
						Required:            true,
						Sensitive:           true,
					},
				},
			},

			"frequency": schema.StringAttribute{
				MarkdownDescription: "Define how frequent the check should run. Allowed values are: 1m, 5m, 15m, 30m and 60m.",
				Optional:            true,
				Computed:            true,
				Validators: []validator.String{
					stringvalidator.OneOf("1m", "5m", "15m", "30m", "60m"),
				},
				Default: stringdefault.StaticString("5m"),
			},
			"message": schema.StringAttribute{
				MarkdownDescription: "A custom message for the check to be send in the notifications.",
				Optional:            true,
			},
			"contact_ids": schema.SetAttribute{
				MarkdownDescription: "A list of contact IDs that will be notified.",
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Default:             setdefault.StaticValue(types.SetValueMust(types.StringType, []attr.Value{})),
			},
			"response_time_threshold": schema.Int64Attribute{
				MarkdownDescription: "Triggers a downtime if the response time exceeds this threshold (in ms). The default value is 30s (30000ms).",
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(30000),
			},
			"notify_when_down": schema.Int64Attribute{
				MarkdownDescription: "Notify the contacts when the check is down for X times. The default value is 2.",
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(2),
			},
			"notify_again_every": schema.Int64Attribute{
				MarkdownDescription: "Notify the contacts again when the check continues to be down after X times. The default value is 0.",
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(0),
			},
			"notify_when_back_up": schema.BoolAttribute{
				MarkdownDescription: "Notify the contacts when the check is back-up.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
			},

			"ssl_down_days_before": schema.Int64Attribute{
				MarkdownDescription: "Trigger a downtime if the SSL certificate expires in the given days. The default value is 7 days.",
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(7),
			},
			"verify_certificate": schema.BoolAttribute{
				MarkdownDescription: "Trigger a downtime if the SSL certificate is invalid or unverifiable. The default value is true.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
			},

			"regions": schema.SetAttribute{
				MarkdownDescription: "A list of regions from which the check will be performed.",
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Default:             setdefault.StaticValue(types.SetValueMust(types.StringType, []attr.Value{})),
				Validators: []validator.Set{
					setvalidator.ValueStringsAre(
						stringvalidator.OneOf("EU", "NA", "APAC", "LATAM"),
					),
				},
			},

			"tags": schema.MapAttribute{
				MarkdownDescription: "A list of tags for the check.",
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Default:             mapdefault.StaticValue(types.MapValueMust(types.StringType, map[string]attr.Value{})),
			},
		},
	}
}

func (r *HTTPCheckResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(api.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected api.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
}

func transformPingdomCheckToModel(check api_types.Check) (HTTPCheckResourceModel, diag.Diagnostics) {
	var contactIds []attr.Value
	for _, userId := range check.UserIDs {
		contactIds = append(contactIds, types.StringValue(strconv.FormatInt(userId, 10)))
	}

	tags := map[string]attr.Value{}
	for _, tag := range check.Tags {
		s := strings.Split(tag.Name, ":")
		if len(s) != 2 {
			continue
		}

		tags[s[0]] = types.StringValue(s[1])
	}

	var auth *HTTPCheckAuthModel
	if check.Type.HTTP.Username != "" && check.Type.HTTP.Password != "" {
		auth = &HTTPCheckAuthModel{
			Username: types.StringValue(check.Type.HTTP.Username),
			Password: types.StringValue(check.Type.HTTP.Password),
		}
	}

	var regions []attr.Value
	for _, filter := range check.ProbeFilters {
		if !strings.HasPrefix(filter, "region: ") {
			continue
		}

		regions = append(regions, types.StringValue(strings.ReplaceAll(filter, "region: ", "")))
	}

	tfContactIds, diagnostics := types.SetValue(types.StringType, contactIds)
	if diagnostics.HasError() {
		return HTTPCheckResourceModel{}, diagnostics
	}

	tfTags, diagnostics := types.MapValue(types.StringType, tags)
	if diagnostics.HasError() {
		return HTTPCheckResourceModel{}, diagnostics
	}

	tfRegions, diagnostics := types.SetValue(types.StringType, regions)
	if diagnostics.HasError() {
		return HTTPCheckResourceModel{}, diagnostics
	}

	message := types.StringNull()
	if check.CustomMessage != "" {
		message = types.StringValue(check.CustomMessage)
	}

	return HTTPCheckResourceModel{
		Id:     types.StringValue(strconv.FormatInt(check.Id, 10)),
		Name:   types.StringValue(check.Name),
		Paused: types.BoolValue(check.Status == "paused"),

		Host: types.StringValue(check.Hostname),
		Url:  types.StringValue(check.Type.HTTP.URL),
		Auth: auth,

		Frequency:             types.StringValue(fmt.Sprintf("%dm", check.Resolution)),
		Message:               message,
		ContactIds:            tfContactIds,
		ResponseTimeThreshold: types.Int64Value(check.ResponseTimeThreshold),
		NotifyWhenDown:        types.Int64Value(check.SendNotificationWhenDown),
		NotifyAgainEvery:      types.Int64Value(check.NotifyAgainEvery),
		NotifyWhenBackUp:      types.BoolValue(check.NotifyWhenBackup),

		SSLDownDaysBefore: types.Int64Value(check.Type.HTTP.SSLDownDaysBefore),
		VerifyCertificate: types.BoolValue(check.Type.HTTP.VerifyCertificate),

		Regions: tfRegions,

		Tags: tfTags,
	}, nil
}

func createCheckRequestModel(resourceModel HTTPCheckResourceModel) api.CreateCheckRequest {
	frequency, err := time.ParseDuration(resourceModel.Frequency.ValueString())
	if err != nil {
		panic(err)
	}

	var auth string
	if resourceModel.Auth != nil && !resourceModel.Auth.Username.IsNull() && resourceModel.Auth.Password.IsNull() {
		auth = fmt.Sprintf("%s:%s", resourceModel.Auth.Username.ValueString(), resourceModel.Auth.Password.ValueString())
	}

	userIds := []string{}
	for _, contactId := range resourceModel.ContactIds.Elements() {
		stringValue, ok := contactId.(types.String)
		if !ok {
			continue
		}

		userIds = append(userIds, stringValue.ValueString())
	}

	probeFilters := []string{}
	for _, region := range resourceModel.Regions.Elements() {
		stringValue, ok := region.(types.String)
		if !ok {
			continue
		}

		probeFilters = append(probeFilters, fmt.Sprintf("region: %s", stringValue.ValueString()))
	}

	tags := []string{}
	for key, value := range resourceModel.Tags.Elements() {
		stringValue, ok := value.(types.String)
		if !ok {
			continue
		}

		tags = append(tags, fmt.Sprintf("%s:%s", key, stringValue.ValueString()))
	}

	return api.CreateCheckRequest{
		Name:                     resourceModel.Name.ValueString(),
		Host:                     resourceModel.Host.ValueString(),
		Auth:                     auth,
		Encryption:               true,
		Type:                     "http",
		VerifyCertificate:        resourceModel.VerifyCertificate.ValueBool(),
		SSLDownDaysBefore:        resourceModel.SSLDownDaysBefore.ValueInt64(),
		NotifyWhenBackup:         resourceModel.NotifyWhenBackUp.ValueBool(),
		NotifyAgainEvery:         resourceModel.NotifyAgainEvery.ValueInt64(),
		SendNotificationWhenDown: resourceModel.NotifyWhenDown.ValueInt64(),
		Url:                      resourceModel.Url.ValueString(),
		ResponseTimeThreshold:    resourceModel.ResponseTimeThreshold.ValueInt64(),
		CustomMessage:            resourceModel.Message.ValueString(),
		Paused:                   resourceModel.Paused.ValueBool(),
		UserIds:                  strings.Join(userIds, ","),
		ProbeFilters:             probeFilters,
		Tags:                     tags,
		Resolution:               frequency.Minutes(),
	}

}

func (r *HTTPCheckResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var model HTTPCheckResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	checkId, err := r.client.CreateCheck(ctx, createCheckRequestModel(model))
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create check, got error: %s", err))
		return
	}

	check, err := r.client.GetCheck(ctx, strconv.FormatInt(*checkId, 10))
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create check, got error: %s", err))
		return
	}

	model, diagnostics := transformPingdomCheckToModel(*check)
	if diagnostics.HasError() {
		resp.Diagnostics.Append(diagnostics...)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *HTTPCheckResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var model HTTPCheckResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	check, err := r.client.GetCheck(ctx, model.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read example, got error: %s", err))
		return
	}

	model, diagnostics := transformPingdomCheckToModel(*check)
	if diagnostics.HasError() {
		resp.Diagnostics.Append(diagnostics...)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *HTTPCheckResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data HTTPCheckResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.UpdateCheck(ctx, data.Id.ValueString(), createCheckRequestModel(data))
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update example, got error: %s", err))
		return
	}

	check, err := r.client.GetCheck(ctx, data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read example, got error: %s", err))
		return
	}

	model, diagnostics := transformPingdomCheckToModel(*check)
	if diagnostics.HasError() {
		resp.Diagnostics.Append(diagnostics...)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *HTTPCheckResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data HTTPCheckResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteCheck(ctx, data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete example, got error: %s", err))
		return
	}
}

func (r *HTTPCheckResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
