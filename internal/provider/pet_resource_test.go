// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccPetResource(t *testing.T) {
	resourceName := "petstore_pet.example"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccPetResourceConfig("newdog", "available"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", "newdog"),
					resource.TestCheckResourceAttr(resourceName, "status", "available"),
					// TODO: check if the first photo url is "url1"
				),
			},
			// ImportState testing
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				// This is not normally necessary, but is here because this
				// example code does not have an actual upstream service.
				// Once the Read method is able to refresh information from
				// the upstream service, this can be removed.
				// ImportStateVerifyIgnore: []string{"configurable_attribute", "defaulted"},
			},
			// Update and Read testing
			{
				Config: testAccPetResourceConfig("anotherdog", "pending"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", "anotherdog"),
					resource.TestCheckResourceAttr(resourceName, "status", "pending"),
					// TODO: check if the first photo url is still "url1"
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccPetResourceConfig(name, status string) string {
	return fmt.Sprintf(`
resource "petstore_pet" "example" {
	name = "%s"
	category = {
		id   = 1
		name = "dog"
	}
	photo_urls = [
		"url1",
		"url2"
	]
	status = "%s"
	}
`, name, status)
}
