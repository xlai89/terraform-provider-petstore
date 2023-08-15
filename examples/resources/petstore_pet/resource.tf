resource "petstore_pet" "example" {
  name = "new cat 9"
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
