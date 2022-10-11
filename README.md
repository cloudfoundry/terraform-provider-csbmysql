# terraform-provider-csbmysql

This is a highly specialised Terraform provider designed to be used exclusively with the [Cloud Service Broker](https://github.com/cloudfoundry-incubator/cloud-service-broker) ("CSB") to create OSBAPI-compliant service bindings in MySQL. Initially CSB brokerpaks used the original Terraform provider for MySQL, however it is no longer maintained.

## Usage example
```terraform
terraform {
  required_providers {
    csbmysql = {
      source  = "cloudfoundry.org/cloud-service-broker/csbmysql"
      version = "1.0.0"
    }
  }
}

provider "csbmysql" {
  host            = "localhost"
  port            = 3306
  username        = "admin-user"
  password        = "fill-me-in"
  database        = "mysql"
}

resource "csbmysql_binding_user" "binding_user" {
  username = "foo"
  password = "bar"
}
```

## Releasing
To create a new GitHub release, decide on a new version number [according to Semanitc Versioning](https://semver.org/), and then:
1. Create a tag on the main branch with a leading `v`:
   `git tag vX.Y.X`
1. Push the tag:
   `git push --tags`
1. Wait for the GitHub action to run GoReleaser and create the new GitHub release

