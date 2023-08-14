// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"

	"terraform-provider-petstore/petstoreapi"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &PetResource{}
var _ resource.ResourceWithImportState = &PetResource{}

func NewPetResource() resource.Resource {
	return &PetResource{}
}

// PetResource defines the resource implementation.
type PetResource struct {
	client *petstoreapi.ClientWithResponses
}

// PetResourceModel describes the resource data model.
type PetResourceModel struct {
	Id       types.String      `tfsdk:"id"`
	Name     types.String      `tfsdk:"name"`
	Status   types.String      `tfsdk:"status"`
	Category *petCategoryModel `tfsdk:"category"`
	Tags     *[]petTagModel    `tfsdk:"tags"`

	// TODO: implement the field "PhotoUrls"
}

type petCategoryModel struct {
	Id   types.Int64  `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

type petTagModel struct {
	Id   types.Int64  `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

func (r *PetResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_pet"
}

func (r *PetResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Pet resource",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Pet identifier",
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Pet name",
				Required:            true,
			},
			"category": schema.SingleNestedAttribute{
				MarkdownDescription: "Pet category",
				Required:            true,
				Attributes: map[string]schema.Attribute{
					"id": schema.Int64Attribute{
						MarkdownDescription: "Category identifier",
						Required:            true,
					},
					"name": schema.StringAttribute{
						MarkdownDescription: "Category name",
						Required:            true,
					},
				},
			},
			"tags": schema.ListNestedAttribute{
				MarkdownDescription: "Pet tags",
				Optional:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.Int64Attribute{
							MarkdownDescription: "Tag identifier",
							Optional:            true,
						},
						"name": schema.StringAttribute{
							MarkdownDescription: "Tag name",
							Optional:            true,
						},
					},
				},
			},
			"status": schema.StringAttribute{
				MarkdownDescription: "Pet status",
				Required:            true,
				// TODO: validate status input according to its enum
			},
			// TODO: implement the attribute "photo_urls"
		},
	}
}

func (r *PetResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*petstoreapi.ClientWithResponses)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *petstoreapi.ClientWithResponses, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
}

func (r *PetResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan PetResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("from pet_resource.go: output pet plan: %#v", plan))

	// If applicable, this is a great opportunity to initialize any necessary
	// provider client data and make a call using it.
	params := petstoreapi.Pet{
		Name: plan.Name.ValueString(),
		Id:   types.Int64Unknown().ValueInt64Pointer(),
		Category: &petstoreapi.Category{
			Id:   types.Int64Unknown().ValueInt64Pointer(),
			Name: types.StringUnknown().ValueStringPointer(),
		},
		// TODO: activate the field "PhotoUrls" after it's implemented
		// PhotoUrls: []string{},
		Tags:   &[]petstoreapi.Tag{},
		Status: nil,
	}

	if plan.Category != nil {
		if !plan.Category.Id.IsNull() {
			tmpId := plan.Category.Id.ValueInt64()
			params.Category.Id = &tmpId
		}
		if !plan.Category.Name.IsNull() {
			tmpName := plan.Category.Name.ValueString()
			params.Category.Name = &tmpName
		}
	}

	// TODO: add photo urls from the resource schema into api payload
	// pay attention to the types of the source var (plan.PhotoUrls)
	// and the target var (params.PhotoUrls)

	if plan.Tags != nil && len(*plan.Tags) > 0 {
		for _, tag := range *plan.Tags {
			tmp := petstoreapi.Tag{}
			if !tag.Id.IsNull() {
				tmpId := tag.Id.ValueInt64()
				tmp.Id = &tmpId
			}
			if !tag.Name.IsNull() {
				tmpName := tag.Name.ValueString()
				tmp.Name = &tmpName
			}
			(*params.Tags) = append((*params.Tags), tmp)
		}
	}

	if !plan.Status.IsNull() {
		tmp := petstoreapi.PetStatus(plan.Status.ValueString())
		params.Status = &tmp
	}

	// generate a random id for the pet
	tmpId := rand.Int63n(1000)
	params.Id = &tmpId
	tflog.Trace(ctx, "from pet_resource.go: generated pet id", map[string]interface{}{
		"id": tmpId,
	})

	tflog.Trace(ctx, fmt.Sprintf("from pet_resource.go: output pet params for create: %#v", params))

	addResp, err := r.client.AddPetWithResponse(ctx, params)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create pet, got error: %s", err))
		return
	}

	tflog.Trace(ctx, "from pet_resource.go: create pet and got a http response", map[string]interface{}{
		"status": addResp.HTTPResponse.StatusCode,
		"url":    fmt.Sprintf("%#v", addResp.HTTPResponse.Request.URL),
		"body":   fmt.Sprintf("%#v", string(addResp.Body)),
	})

	// If status code is not 200, return an error
	if addResp.JSON200 == nil {
		resp.Diagnostics.AddError("Server Error",
			fmt.Sprintf("Unable to create pet, got status code: %d", addResp.HTTPResponse.StatusCode))
		return
	}

	// Map response body to schema and populate Computed attribute values
	pet := addResp.JSON200

	plan.Id = types.StringValue(fmt.Sprintf("%d", *pet.Id))
	tflog.Trace(ctx, "from pet_resource.go: set pet id", map[string]interface{}{
		"id": *pet.Id,
	})

	if pet.Category != nil {
		plan.Category = &petCategoryModel{
			Id:   types.Int64Value(*pet.Category.Id),
			Name: types.StringValue(*pet.Category.Name),
		}
	}

	// TODO: map photo urls from the api response to resource schema
	// pay attention to the types of the source var (pet.PhotoUrls)
	// and the target var (plan.PhotoUrls)

	if pet.Tags != nil && len(*pet.Tags) > 0 {
		tmp := make([]petTagModel, len(*pet.Tags))
		for i, tag := range *pet.Tags {
			tmp[i] = petTagModel{
				Id:   types.Int64Value(*tag.Id),
				Name: types.StringValue(*tag.Name),
			}
		}
		plan.Tags = &tmp
	}

	if pet.Status != nil {
		switch *pet.Status {
		case petstoreapi.PetStatusAvailable:
			plan.Status = types.StringValue("available")
		case petstoreapi.PetStatusPending:
			plan.Status = types.StringValue("pending")
		case petstoreapi.PetStatusSold:
			plan.Status = types.StringValue("sold")
		}
	}

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Trace(ctx, "from pet_resource.go: created a resource")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *PetResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data PetResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Trace(ctx, "from pet_resource.go: use pet id", map[string]interface{}{
		"id": data.Id.ValueString(),
	})

	// If applicable, this is a great opportunity to initialize any necessary
	// provider client data and make a call using it.
	petId, _ := strconv.ParseInt(data.Id.ValueString(), 10, 64)
	getResp, err := r.client.GetPetByIdWithResponse(ctx, petId)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read pet, got error: %s", err))
		return
	}

	tflog.Trace(ctx, "from pet_resource.go: read a pet and got a http response", map[string]interface{}{
		"status": getResp.HTTPResponse.StatusCode,
		"url":    fmt.Sprintf("%#v", getResp.HTTPResponse.Request.URL),
		"body":   fmt.Sprintf("%#v", string(getResp.Body)),
	})

	// If status code is not 200, return an error
	if getResp.JSON200 == nil {
		resp.Diagnostics.AddError("Server Error",
			fmt.Sprintf("Unable to read pet, got status code: %d", getResp.HTTPResponse.StatusCode))
		return
	}

	// Map response body to schema and populate Computed attribute values
	pet := getResp.JSON200

	data.Name = types.StringValue(pet.Name)

	data.Id = types.StringValue(fmt.Sprintf("%d", *pet.Id))

	if pet.Category != nil {
		data.Category = &petCategoryModel{
			Id:   types.Int64Value(*pet.Category.Id),
			Name: types.StringValue(*pet.Category.Name),
		}
	}

	// TODO: map photo urls from the api response to resource schema
	// pay attention to the types of the source var (pet.PhotoUrls)
	// and the target var (data.PhotoUrls)

	if pet.Tags != nil && len(*pet.Tags) > 0 {
		tmp := make([]petTagModel, len(*pet.Tags))
		for i, tag := range *pet.Tags {
			tmp[i] = petTagModel{
				Id:   types.Int64Value(*tag.Id),
				Name: types.StringValue(*tag.Name),
			}
		}
		data.Tags = &tmp
	}

	if pet.Status != nil {
		switch *pet.Status {
		case petstoreapi.PetStatusAvailable:
			data.Status = types.StringValue("available")
		case petstoreapi.PetStatusPending:
			data.Status = types.StringValue("pending")
		case petstoreapi.PetStatusSold:
			data.Status = types.StringValue("sold")
		}
	}

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Trace(ctx, "from pet_resource.go: read a resource")

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *PetResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan PetResourceModel
	var data PetResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// If applicable, this is a great opportunity to initialize any necessary
	// provider client data and make a call using it.
	params := &petstoreapi.UpdatePetWithFormParams{}

	if !plan.Name.IsNull() && !plan.Name.Equal(data.Name) {
		params.Name = plan.Name.ValueStringPointer()
	}

	// TODO: implement update for status

	// If only status is changed, add name to params because it is required
	if params.Name == nil && params.Status != nil {
		params.Name = data.Name.ValueStringPointer()
	}

	// If params is empty, return
	if params.Name == nil && params.Status == nil {
		tflog.Trace(ctx, "from pet_resource.go: no update params")
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("from pet_resource.go: output pet params for update: %#v", *params))

	tflog.Trace(ctx, "from pet_resource.go: use pet id", map[string]interface{}{
		"id": data.Id.ValueString(),
	})

	petId, _ := strconv.ParseInt(data.Id.ValueString(), 10, 64)
	updateResp, err := r.client.UpdatePetWithFormWithResponse(ctx, petId, params)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update pet, got error: %s", err))
		return
	}

	tflog.Trace(ctx, "from pet_resource.go: update a pet and got a http response", map[string]interface{}{
		"status": updateResp.HTTPResponse.StatusCode,
		"url":    fmt.Sprintf("%#v", updateResp.HTTPResponse.Request.URL),
		"body":   fmt.Sprintf("%#v", string(updateResp.Body)),
	})

	// If status code is not 200, return an error
	if updateResp.JSON200 == nil {
		resp.Diagnostics.AddError("Server Error",
			fmt.Sprintf("Unable to update pet, got status code: %d", updateResp.HTTPResponse.StatusCode))
		return
	}

	// Map response body to schema and populate Computed attribute values
	pet := updateResp.JSON200

	data.Name = types.StringValue(pet.Name)

	data.Id = types.StringValue(fmt.Sprintf("%d", *pet.Id))

	if pet.Category != nil {
		data.Category = &petCategoryModel{
			Id:   types.Int64Value(*pet.Category.Id),
			Name: types.StringValue(*pet.Category.Name),
		}
	}

	// TODO: map photo urls from the api response to resource schema
	// pay attention to the types of the source var (pet.PhotoUrls)
	// and the target var (data.PhotoUrls)

	if pet.Tags != nil && len(*pet.Tags) > 0 {
		tmp := make([]petTagModel, len(*pet.Tags))
		for i, tag := range *pet.Tags {
			tmp[i] = petTagModel{
				Id:   types.Int64Value(*tag.Id),
				Name: types.StringValue(*tag.Name),
			}
		}
		data.Tags = &tmp
	}

	if pet.Status != nil {
		switch *pet.Status {
		case petstoreapi.PetStatusAvailable:
			data.Status = types.StringValue("available")
		case petstoreapi.PetStatusPending:
			data.Status = types.StringValue("pending")
		case petstoreapi.PetStatusSold:
			data.Status = types.StringValue("sold")
		}
	}

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Trace(ctx, "from pet_resource.go: updated a resource")

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

}

func (r *PetResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data PetResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Trace(ctx, "from pet_resource.go: use pet id", map[string]interface{}{
		"id": data.Id.ValueString(),
	})

	// If applicable, this is a great opportunity to initialize any necessary
	// provider client data and make a call using it.
	// TODO: get the pet id and make an api call for delete
	// pay attention to the variable types
	// petId := "fake"
	delResp := &petstoreapi.DeletePetResponse{}
	var err error
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete pet, got error: %s", err))
		return
	}

	tflog.Trace(ctx, "from pet_resource.go: delete a pet and got a http response", map[string]interface{}{
		"status": delResp.HTTPResponse.StatusCode,
		"url":    fmt.Sprintf("%#v", delResp.HTTPResponse.Request.URL),
		"body":   fmt.Sprintf("%#v", string(delResp.Body)),
	})

	// If status code is not 200, return an error
	if delResp.HTTPResponse.StatusCode != 200 {
		resp.Diagnostics.AddError("Server Error",
			fmt.Sprintf("Unable to delete pet, got status code: %d", delResp.HTTPResponse.StatusCode))
		return
	}

}

func (r *PetResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
