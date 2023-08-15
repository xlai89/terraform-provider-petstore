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

data "petstore_user" "example" {
  username = "theUser"
}

resource "petstore_pet" "example" {
  name = "new dog 2"
  category = {
    id   = 1
    name = "dog"
  }
  photo_urls = [
    "url1",
    "url2"
  ]
  status = "pending"
}
