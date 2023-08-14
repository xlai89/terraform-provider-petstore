// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"terraform-provider-petstore/petstoreapi"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ datasource.DataSource              = &UserDataSource{}
	_ datasource.DataSourceWithConfigure = &UserDataSource{}
)

func NewUserDataSource() datasource.DataSource {
	return &UserDataSource{}
}

// UserDataSource defines the data source implementation.
type UserDataSource struct {
	client *petstoreapi.ClientWithResponses
}

// TODO: implement UserDataSourceModel according to "User" from petstore-openapi.yaml
// UserDataSourceModel describes the data source data model.
type UserDataSourceModel struct {
	ConfigurableAttribute types.String `tfsdk:"configurable_attribute"`
	Id                    types.String `tfsdk:"id"`
}

func (d *UserDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user"
}

func (d *UserDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "User data source",

		Attributes: map[string]schema.Attribute{
			// TODO: implement schema for user data source
			"configurable_attribute": schema.StringAttribute{
				MarkdownDescription: "Example configurable attribute",
				Optional:            true,
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "Example identifier",
				Computed:            true,
			},
		},
	}
}

func (d *UserDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*petstoreapi.ClientWithResponses)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *petstoreapi.ClientWithResponses, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.client = client
}

func (d *UserDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data UserDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// If applicable, this is a great opportunity to initialize any necessary
	// provider client data and make a call using it.
	// TODO: implement the API call using petstoreapi client
	userResp := petstoreapi.GetUserByNameResponse{}
	var err error
	if err != nil {
		resp.Diagnostics.AddError(
			"Client error",
			fmt.Sprintf("Unable to get users: %s", err),
		)
	}

	tflog.Trace(ctx, "get users and got a http response", map[string]interface{}{
		"status": userResp.HTTPResponse.StatusCode,
		"url":    fmt.Sprintf("%#v", userResp.HTTPResponse.Request.URL),
		"body":   fmt.Sprintf("%#v", userResp.Body),
	})

	// If status code is not 200, return an error
	if userResp.HTTPResponse.StatusCode != 200 {
		resp.Diagnostics.AddError("Server Error",
			fmt.Sprintf("Unable to get users, got status code: %d", userResp.HTTPResponse.StatusCode))
	}

	// If no users are found, return an error.
	if userResp.JSON200 == nil {
		// TODO: implement an error message "User with name %s not found."
		// use the if statement above as an example.
		tflog.Info(ctx, "to be implemented")
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Save the response value into the Terraform state.
	user := (*userResp.JSON200)

	// TODO: map response body to schema and populate Computed attribute values
	data.Id = types.StringValue(fmt.Sprintf("%d", *user.Id))

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Trace(ctx, "read a user data source")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
