package provider

import (
	"context"
	"os"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/rodriguezst/terraform-provider-lightningai/internal/client"
)

var _ provider.Provider = &LightningProvider{}
var _ provider.ProviderWithFunctions = &LightningProvider{}

// LightningProvider defines the provider implementation.
type LightningProvider struct {
	version string
}

// LightningProviderModel describes the provider data model.
type LightningProviderModel struct {
	APIKey    types.String `tfsdk:"api_key"`
	UserID    types.String `tfsdk:"user_id"`
	ProjectID types.String `tfsdk:"project_id"`
}

// New returns a new provider factory function.
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &LightningProvider{
			version: version,
		}
	}
}

func (p *LightningProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "lightning"
	resp.Version = p.version
}

func (p *LightningProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Terraform provider for managing Lightning AI Studios.",
		Attributes: map[string]schema.Attribute{
			"api_key": schema.StringAttribute{
				Description: "Lightning AI API key. Can be set via LIGHTNING_API_KEY environment variable.",
				Optional:    true,
				Sensitive:   true,
			},
			"user_id": schema.StringAttribute{
				Description: "Lightning AI user ID. Can be set via LIGHTNING_USER_ID environment variable.",
				Optional:    true,
			},
			"project_id": schema.StringAttribute{
				Description: "Lightning AI project/teamspace ID. Can be set via LIGHTNING_PROJECT_ID environment variable.",
				Optional:    true,
			},
		},
	}
}

func (p *LightningProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config LightningProviderModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiKey := os.Getenv("LIGHTNING_API_KEY")
	if !config.APIKey.IsNull() && !config.APIKey.IsUnknown() {
		apiKey = config.APIKey.ValueString()
	}

	userID := os.Getenv("LIGHTNING_USER_ID")
	if !config.UserID.IsNull() && !config.UserID.IsUnknown() {
		userID = config.UserID.ValueString()
	}

	projectID := os.Getenv("LIGHTNING_PROJECT_ID")
	if !config.ProjectID.IsNull() && !config.ProjectID.IsUnknown() {
		projectID = config.ProjectID.ValueString()
	}

	// Validate API key
	if apiKey == "" {
		resp.Diagnostics.AddError(
			"Missing API Key",
			"The Lightning AI API key must be set via the api_key provider attribute or the LIGHTNING_API_KEY environment variable.",
		)
	} else if strings.TrimSpace(apiKey) == "" {
		resp.Diagnostics.AddError(
			"Invalid API Key",
			"The Lightning AI API key cannot be empty or contain only whitespace.",
		)
	}

	// Validate user ID format (UUID-like or alphanumeric)
	if userID == "" {
		resp.Diagnostics.AddError(
			"Missing User ID",
			"The Lightning AI user ID must be set via the user_id provider attribute or the LIGHTNING_USER_ID environment variable.",
		)
	} else if strings.TrimSpace(userID) == "" {
		resp.Diagnostics.AddError(
			"Invalid User ID",
			"The Lightning AI user ID cannot be empty or contain only whitespace.",
		)
	} else if !isValidID(userID) {
		resp.Diagnostics.AddError(
			"Invalid User ID",
			"The Lightning AI user ID must contain only alphanumeric characters, hyphens, and underscores.",
		)
	}

	// Validate project ID format (UUID-like or alphanumeric)
	if projectID == "" {
		resp.Diagnostics.AddError(
			"Missing Project ID",
			"The Lightning AI project ID must be set via the project_id provider attribute or the LIGHTNING_PROJECT_ID environment variable.",
		)
	} else if strings.TrimSpace(projectID) == "" {
		resp.Diagnostics.AddError(
			"Invalid Project ID",
			"The Lightning AI project ID cannot be empty or contain only whitespace.",
		)
	} else if !isValidID(projectID) {
		resp.Diagnostics.AddError(
			"Invalid Project ID",
			"The Lightning AI project ID must contain only alphanumeric characters, hyphens, and underscores.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	c := client.NewClient(apiKey, userID, projectID)
	resp.DataSourceData = c
	resp.ResourceData = c
}

// isValidID checks if an ID contains only safe characters (alphanumeric, hyphens, underscores)
func isValidID(id string) bool {
	// Allow alphanumeric characters, hyphens, and underscores (typical for UUIDs and IDs)
	validIDPattern := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	return validIDPattern.MatchString(id)
}

func (p *LightningProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewStudioResource,
	}
}

func (p *LightningProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return nil
}

func (p *LightningProvider) Functions(_ context.Context) []func() function.Function {
	return nil
}
