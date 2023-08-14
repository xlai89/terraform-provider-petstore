// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	// TODO: add an import to use the petstore client.
	"terraform-provider-petstore/petstoreapi"
)

// Ensure PetstoreProvider satisfies various provider interfaces.
var _ provider.Provider = &PetstoreProvider{}

// PetstoreProvider defines the provider implementation.
type PetstoreProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// PetstoreProviderModel describes the provider data model.
type PetstoreProviderModel struct {
	Server types.String `tfsdk:"server"`
}

func (p *PetstoreProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "petstore"
	resp.Version = p.version
}

func (p *PetstoreProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"server": schema.StringAttribute{
				MarkdownDescription: "the petstore server",
				Optional:            true,
			},
		},
	}
}

func (p *PetstoreProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data PetstoreProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Configuration values are now available.
	// If practitioner provided a configuration value for any of the
	// attributes, it must be a known value.
	if data.Server.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("server"),
			"Unknown PetStore API Host",
			"The provider cannot create the PetStore API client as there is an unknown configuration value for the PetStore API host. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the PETSTORE_SERVER environment variable.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Default values to environment variables, but override
	// with Terraform configuration value if set.

	// TODO: get server from env var `PETSTORE_SERVER`
	server := os.Getenv("PETSTORE_SERVER")

	if !data.Server.IsNull() {
		server = data.Server.ValueString()
	}

	// If any of the expected configurations are missing, return
	// errors with provider-specific guidance.
	if server == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("server"),
			"Unknown Petstore API Host",
			"The provider cannot create the Petstore API client as there is an unknown configuration value for the Petstore API host. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the PETSTORE_SERVER environment variable.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// TODO: Create a new petstore client using the configuration values
	client, err := petstoreapi.NewClientWithResponses(server)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Petstore API Client",
			"An unexpected error occurred when creating the Petstore API client. "+
				"If the error is not clear, please contact the provider developers.\n\n"+
				"Petstore Client Error: "+err.Error(),
		)
		return
	}

	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *PetstoreProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewExampleResource,
	}
}

func (p *PetstoreProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewUserDataSource,
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &PetstoreProvider{
			version: version,
		}
	}
}
