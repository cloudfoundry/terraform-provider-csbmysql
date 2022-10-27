# Run "make init" to perform "terraform init"
# The easiest way to get a MySQL is: docker run -e MYSQL_ROOT_PASSWORD="fill-me-in" -t mysql

terraform {
  required_providers {
    csbmysql = {
      source  = "cloudfoundry.org/cloud-service-broker/csbmysql"
      version = "1.0.0"
    }
  }
}

provider "csbmysql" {
  host     = "localhost"
  port     = 3306
  username = "admin-user"
  password = "fill-me-in"
  database = "mysql"
}

resource "csbmysql_binding_user" "binding_user" {
  username = "foo"
  password = "bar"
}
