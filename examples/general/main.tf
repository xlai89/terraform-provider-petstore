terraform {
  required_providers {
    petstore = {
      source = "registry.terraform.io/xlai89/petstore"
    }
  }
}

provider "petstore" {
  server = "http://localhost:9090/api/v3"
}

data "petstore_example" "example" {}
