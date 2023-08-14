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
	Id        types.Int64  `tfsdk:"id"`
	Username  types.String `tfsdk:"username"`
	Password  types.String `tfsdk:"password"`
	Firstname types.String `tfsdk:"firstname"`
	Lastname  types.String `tfsdk:"lastname"`
	Email     types.String `tfsdk:"email"`
	Phone     types.String `tfsdk:"phone"`
	Status    types.Int64  `tfsdk:"status"`
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
			"id": schema.Int64Attribute{
				MarkdownDescription: "User identifier",
				Computed:            true,
			},
			"username": schema.StringAttribute{
				MarkdownDescription: "User name",
				Required:            true,
			},
			"password": schema.StringAttribute{
				MarkdownDescription: "User password",
				Computed:            true,
				Sensitive:           true,
			},
			"firstname": schema.StringAttribute{
				MarkdownDescription: "User first name",
				Computed:            true,
			},
			"lastname": schema.StringAttribute{
				MarkdownDescription: "User last name",
				Computed:            true,
			},
			"email": schema.StringAttribute{
				MarkdownDescription: "User email",
				Computed:            true,
			},
			"phone": schema.StringAttribute{
				MarkdownDescription: "User phone",
				Computed:            true,
			},
			"status": schema.Int64Attribute{
				MarkdownDescription: "User status",
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
	userResp, err := d.client.GetUserByNameWithResponse(ctx, data.Username.ValueString())
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
		resp.Diagnostics.AddError(
			"User not found",
			fmt.Sprintf("User with name %s not found.", data.Username.ValueString()),
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Save the response value into the Terraform state.
	user := (*userResp.JSON200)

	// TODO: map response body to schema and populate Computed attribute values
	data.Id = types.Int64Value(*user.Id)
	data.Username = types.StringValue(*user.Username)
	data.Password = types.StringValue(*user.Password)
	data.Firstname = types.StringValue(*user.FirstName)
	data.Lastname = types.StringValue(*user.LastName)
	data.Email = types.StringValue(*user.Email)
	data.Phone = types.StringValue(*user.Phone)
	data.Status = types.Int64Value(int64(*user.UserStatus))

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Trace(ctx, "read a user data source")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
