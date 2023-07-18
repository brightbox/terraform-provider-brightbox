package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure BrightboxProvider satisfies various provider interfaces.
var _ provider.Provider = &BrightboxProvider{}

// BrightboxProvider defines the provider implementation.
type BrightboxProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// BrightboxProviderModel describes the provider data model.
type BrightboxProviderModel struct {
	APIClient types.String `tfsdk:"apiclient"`
	APISecret types.String `tfsdk:"apisecret"`
	UserName  types.String `tfsdk:"username"`
	password  types.String `tfsdk:"password"`
	Account   types.String `tfsdk:"account"`
	APIURL    types.String `tfsdk:"apiurl"`
	OrbitURL  types.String `tfsdk:"orbit_url"`
}

// Metadata should return the metadata for the provider, such as
// a type name and version data.
func (p *BrightboxProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "brightbox"
	resp.Version = p.version
}

// Schema should return the schema for this provider.
func (p *BrightboxProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"account": schema.StringAttribute{
				Optional:    true,
				Description: "Brightbox Cloud Account to operate upon",
			},
			"apiclient": schema.StringAttribute{
				Optional:    true,
				Description: "Brightbox Cloud API Client/OAuth Application ID",
			},
			"apisecret": schema.StringAttribute{
				Optional:    true,
				Description: "Brightbox Cloud API Client/OAuth Application Secret",
			},
			"apiurl": schema.StringAttribute{
				Optional:    true,
				Description: "Brightbox Cloud Api URL for selected Region",
			},
			"orbit_url": schema.StringAttribute{
				Optional:    true,
				Description: "Brightbox Cloud Orbit URL for selected Region",
			},
			"password": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "Brightbox Cloud Password for User Name",
			},
			"username": schema.StringAttribute{
				Optional:    true,
				Description: "Brightbox Cloud User Name",
			},
		},
	}
}

// Configure is called at the beginning of the provider lifecycle, when
// Terraform sends to the provider the values the user specified in the
// provider configuration block. These are supplied in the
// ConfigureProviderRequest argument.
func (p *BrightboxProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data BrightboxProviderModel

	tflog.Debug(ctx, "Configure")
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	data = addDefaultsToConfig(data)

	resp.Diagnostics.Append(validateConfig(ctx, data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Configuration values are now available.
	// if data.Endpoint.IsNull() { /* ... */ }

	// Example client configuration for data sources and resources
	// client := configureClient(ctx, data)
	// resp.DataSourceData = client
	// resp.ResourceData = client
}

// Resources returns a slice of functions to instantiate each Resource
// implementation.
func (p *BrightboxProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{}
}

// DataSources returns a slice of functions to instantiate each datasource
func (p *BrightboxProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{}
}

// New returns a versioned provider
func New(version string) provider.Provider {
	return &BrightboxProvider{
		version: version,
	}
}
